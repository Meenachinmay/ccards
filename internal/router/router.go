package router

import (
	"ccards/internal/card"
	"ccards/internal/client"
	"ccards/internal/transaction"
	"ccards/pkg/config"
	"ccards/pkg/middleware"
	"database/sql"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type Router struct {
	engine             *gin.Engine
	clientHandler      *client.Handler
	cardHandler        *card.Handler
	transactionHandler *transaction.Handler
	config             *config.Config
	redisClient        *redis.Client
	db                 *sql.DB
}

type RouterConfig struct {
	ClientHandler      *client.Handler
	CardHandler        *card.Handler
	TransactionHandler *transaction.Handler
	Config             *config.Config
	RedisClient        *redis.Client
	DB                 *sql.DB
}

func NewRouter(cfg RouterConfig) *Router {
	return &Router{
		engine:             gin.New(),
		clientHandler:      cfg.ClientHandler,
		cardHandler:        cfg.CardHandler,
		transactionHandler: cfg.TransactionHandler,
		config:             cfg.Config,
		redisClient:        cfg.RedisClient,
		db:                 cfg.DB,
	}
}

func (r *Router) Setup() *gin.Engine {
	r.engine.Use(gin.Logger())
	r.engine.Use(gin.Recovery())
	r.engine.Use(middleware.CORS())

	r.engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	admin := r.engine.Group("/admin")
	{
		admin.POST("/company/register", r.clientHandler.RegisterCompany)
	}

	authGroup := r.engine.Group("/auth/company")
	{
		authGroup.POST("/login", r.clientHandler.Login)
	}

	apiGroup := r.engine.Group("/api")
	clientAuthMiddleware := middleware.NewClientAuthMiddleware(r.config, r.redisClient)
	apiGroup.Use(clientAuthMiddleware.ClientAuth())
	{
		companyGroup := apiGroup.Group("/company")
		{
			companyGroup.GET("", r.clientHandler.GetCompany)
			companyGroup.POST("/upload-csv", r.clientHandler.UploadCardCSV)
			companyGroup.GET("/card-to-issue", r.clientHandler.GetCardsToIssue)
			companyGroup.POST("/issue-cards", r.clientHandler.IssueNewCards)
		}

		cardGroup := apiGroup.Group("/cards")
		{
			cardGroup.GET("", r.cardHandler.GetCards)
			cardGroup.POST("/update/spending-limit", r.cardHandler.UpdateSpendingLimit) // get company id from context and send card id as query params
			cardGroup.POST("/update/block", r.cardHandler.Block)                        // companyID, cardID
			cardGroup.POST("/update/unblock", r.cardHandler.Unblock)                    // companyID, cardID
			cardGroup.POST("/update/charge", r.cardHandler.Charge)                      // companyID, cardID, amount
			cardGroup.POST("/update/spending-control", r.cardHandler.UpdateSpendingControl)

			transactionGroup := cardGroup.Group("/transactions")
			{
				transactionGroup.Use(
					middleware.ValidCard(r.db),
					middleware.UsableCard(),
					middleware.SufficientAmount(),
					middleware.WithinDailyLimit(r.db),
					middleware.SpendingLimit(r.db),
				)

				transactionGroup.POST("", r.transactionHandler.Pay)
			}

			cardGroup.GET("/transactions", r.transactionHandler.GetTransactionHistory)
			cardGroup.GET("/transactions/:id", r.transactionHandler.GetTransaction)
		}
	}

	return r.engine
}

func (r *Router) Engine() *gin.Engine {
	return r.engine
}
