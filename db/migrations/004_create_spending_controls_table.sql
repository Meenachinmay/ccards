-- +goose Up
-- +goose StatementBegin
CREATE TABLE spending_controls (
                                   id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                   card_id UUID NOT NULL REFERENCES cards(id) ON DELETE CASCADE,
                                   control_type VARCHAR(50) NOT NULL CHECK (control_type IN ('merchant_category', 'merchant_name', 'time_based', 'location')),
                                   control_value JSONB NOT NULL,
                                   is_active BOOLEAN DEFAULT true,
                                   created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                                   updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_spending_controls_card_id ON spending_controls(card_id);
CREATE INDEX idx_spending_controls_type ON spending_controls(control_type);
CREATE INDEX idx_spending_controls_is_active ON spending_controls(is_active);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS spending_controls;
-- +goose StatementEnd