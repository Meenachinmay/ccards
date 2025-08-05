package middleware

import (
	"ccards/internal/api/request"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"ccards/pkg/models"
)

const (
	TokyoTimezone = "Asia/Tokyo"
)

type SpendingLimitMiddleware struct {
	db       *sql.DB
	location *time.Location
}

type TimeBasedControl struct {
	StartTime string `json:"start_time"` // Format: "HH:MM"
	EndTime   string `json:"end_time"`   // Format: "HH:MM"
}

type MerchantCategoryControl struct {
	AllowedCategories []string `json:"allowed_categories"`
	BlockedCategories []string `json:"blocked_categories"`
}

func NewSpendingLimitMiddleware(db *sql.DB) *SpendingLimitMiddleware {
	loc, err := time.LoadLocation(TokyoTimezone)
	if err != nil {
		panic(fmt.Sprintf("Failed to load Tokyo timezone: %v", err))
	}

	return &SpendingLimitMiddleware{
		db:       db,
		location: loc,
	}
}

func SpendingLimit(db *sql.DB) gin.HandlerFunc {
	m := NewSpendingLimitMiddleware(db)
	return m.Handle()
}

func (m *SpendingLimitMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		card, err := getCardFromContext(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Card information not found",
			})
			c.Abort()
			return
		}

		req, err := getTransactionRequestFromContext(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Transaction request not found",
			})
			c.Abort()
			return
		}

		// Get spending controls for the card
		controls, err := m.getSpendingControls(c.Request.Context(), card.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to retrieve spending controls",
			})
			c.Abort()
			return
		}

		// Check each active control
		for _, control := range controls {
			if !control.IsActive {
				continue
			}

			switch control.ControlType {
			case "merchant_category":
				if err := m.checkMerchantCategory(control, req.MerchantCategory); err != nil {
					c.JSON(http.StatusForbidden, gin.H{
						"error":             err.Error(),
						"control_type":      "merchant_category",
						"merchant_category": req.MerchantCategory,
					})
					c.Abort()
					return
				}

			case "time_based":
				if err := m.checkTimeBased(control); err != nil {
					currentTime := time.Now().In(m.location)
					c.JSON(http.StatusForbidden, gin.H{
						"error":        err.Error(),
						"control_type": "time_based",
						"current_time": currentTime.Format("15:04"),
						"timezone":     TokyoTimezone,
					})
					c.Abort()
					return
				}

			case "merchant_name":
				// Implement merchant name checking if needed
				continue

			case "location":
				// Implement location-based checking if needed
				continue
			}
		}

		c.Next()
	}
}

func (m *SpendingLimitMiddleware) getSpendingControls(ctx context.Context, cardID uuid.UUID) ([]*models.SpendingControl, error) {
	query := `
		SELECT id, card_id, control_type, control_value, is_active, created_at, updated_at
		FROM spending_controls
		WHERE card_id = $1 AND is_active = true
		ORDER BY created_at DESC
	`

	rows, err := m.db.QueryContext(ctx, query, cardID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var controls []*models.SpendingControl
	for rows.Next() {
		var control models.SpendingControl
		var controlValueJSON []byte

		err := rows.Scan(
			&control.ID,
			&control.CardID,
			&control.ControlType,
			&controlValueJSON,
			&control.IsActive,
			&control.CreatedAt,
			&control.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		control.ControlValue = json.RawMessage(controlValueJSON)
		controls = append(controls, &control)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return controls, nil
}

func (m *SpendingLimitMiddleware) checkMerchantCategory(control *models.SpendingControl, merchantCategory string) error {
	var categoryControl MerchantCategoryControl

	if err := json.Unmarshal([]byte(control.ControlValue.(json.RawMessage)), &categoryControl); err != nil {
		return fmt.Errorf("invalid merchant category control configuration: %w", err)
	}

	merchantCategory = strings.ToLower(strings.TrimSpace(merchantCategory))

	for _, blocked := range categoryControl.BlockedCategories {
		if merchantCategory == strings.ToLower(strings.TrimSpace(blocked)) {
			return &SpendingControlError{
				Type:    "merchant_category",
				Message: fmt.Sprintf("Transaction blocked: merchant category '%s' is not allowed", merchantCategory),
			}
		}
	}

	if len(categoryControl.AllowedCategories) > 0 {
		allowed := false
		for _, allowedCat := range categoryControl.AllowedCategories {
			if merchantCategory == strings.ToLower(strings.TrimSpace(allowedCat)) {
				allowed = true
				break
			}
		}
		if !allowed {
			return &SpendingControlError{
				Type:    "merchant_category",
				Message: fmt.Sprintf("Transaction blocked: merchant category '%s' is not in allowed list", merchantCategory),
			}
		}
	}

	return nil
}

func (m *SpendingLimitMiddleware) checkTimeBased(control *models.SpendingControl) error {
	var timeControl TimeBasedControl

	if err := json.Unmarshal([]byte(control.ControlValue.(json.RawMessage)), &timeControl); err != nil {
		return fmt.Errorf("invalid time-based control configuration: %w", err)
	}

	now := time.Now().In(m.location)

	startHour, startMinute, err := parseTimeString(timeControl.StartTime)
	if err != nil {
		return fmt.Errorf("invalid start time format: %w", err)
	}

	endHour, endMinute, err := parseTimeString(timeControl.EndTime)
	if err != nil {
		return fmt.Errorf("invalid end time format: %w", err)
	}

	currentHour, currentMinute := now.Hour(), now.Minute()

	currentMinutes := currentHour*60 + currentMinute
	startMinutes := startHour*60 + startMinute
	endMinutes := endHour*60 + endMinute

	isAllowed := false

	if startMinutes <= endMinutes {
		isAllowed = currentMinutes >= startMinutes && currentMinutes <= endMinutes
	} else {
		isAllowed = currentMinutes >= startMinutes || currentMinutes <= endMinutes
	}

	if !isAllowed {
		return &SpendingControlError{
			Type: "time_based",
			Message: fmt.Sprintf(
				"Transaction blocked: outside allowed time window (%s - %s JST)",
				timeControl.StartTime,
				timeControl.EndTime,
			),
		}
	}

	return nil
}

func parseTimeString(timeStr string) (hour, minute int, err error) {
	parts := strings.Split(timeStr, ":")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid time format, expected HH:MM")
	}

	if _, err := fmt.Sscanf(parts[0], "%d", &hour); err != nil {
		return 0, 0, fmt.Errorf("invalid hour: %w", err)
	}

	if _, err := fmt.Sscanf(parts[1], "%d", &minute); err != nil {
		return 0, 0, fmt.Errorf("invalid minute: %w", err)
	}

	if hour < 0 || hour > 23 {
		return 0, 0, fmt.Errorf("hour must be between 0 and 23")
	}
	if minute < 0 || minute > 59 {
		return 0, 0, fmt.Errorf("minute must be between 0 and 59")
	}

	return hour, minute, nil
}

func getCardFromContext(c *gin.Context) (*models.Card, error) {
	cardInterface, exists := c.Get("card")
	if !exists {
		return nil, fmt.Errorf("card not found in context")
	}

	card, ok := cardInterface.(*models.Card)
	if !ok {
		return nil, fmt.Errorf("invalid card type in context")
	}

	return card, nil
}

func getTransactionRequestFromContext(c *gin.Context) (*request.Transaction, error) {
	reqInterface, exists := c.Get("transaction_request")
	if !exists {
		return nil, fmt.Errorf("transaction request not found in context")
	}

	req, ok := reqInterface.(*request.Transaction)
	if !ok {
		return nil, fmt.Errorf("invalid transaction request type in context")
	}

	return req, nil
}

type SpendingControlError struct {
	Type    string
	Message string
}

func (e *SpendingControlError) Error() string {
	return e.Message
}

func CreateInitialSpendingControls(db *sql.DB, cardID uuid.UUID) error {
	merchantControl := MerchantCategoryControl{
		AllowedCategories: []string{"food"},
		BlockedCategories: []string{},
	}

	merchantControlJSON, err := json.Marshal(merchantControl)
	if err != nil {
		return fmt.Errorf("failed to marshal merchant control: %w", err)
	}

	timeControl := TimeBasedControl{
		StartTime: "09:00",
		EndTime:   "18:00",
	}

	timeControlJSON, err := json.Marshal(timeControl)
	if err != nil {
		return fmt.Errorf("failed to marshal time control: %w", err)
	}

	insertQuery := `
		INSERT INTO spending_controls (card_id, control_type, control_value, is_active)
		VALUES ($1, $2, $3, $4)
	`

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err = tx.Exec(insertQuery, cardID, "merchant_category", merchantControlJSON, true); err != nil {
		return fmt.Errorf("failed to insert merchant category control: %w", err)
	}

	if _, err = tx.Exec(insertQuery, cardID, "time_based", timeControlJSON, true); err != nil {
		return fmt.Errorf("failed to insert time-based control: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
