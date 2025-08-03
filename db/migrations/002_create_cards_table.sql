-- +goose Up
-- +goose StatementBegin
CREATE TABLE cards (
                       id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                       company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
                       card_number VARCHAR(16) NOT NULL UNIQUE,
                       card_holder_name VARCHAR(255) NOT NULL,
                       employee_id VARCHAR(100),
                       employee_email VARCHAR(255),
                       card_type VARCHAR(50) NOT NULL DEFAULT 'virtual',
                       status VARCHAR(50) NOT NULL DEFAULT 'active',
                       balance DECIMAL(15, 2) NOT NULL DEFAULT 0.00,
                       spending_limit DECIMAL(15, 2),
                       daily_limit DECIMAL(15, 2),
                       monthly_limit DECIMAL(15, 2),
                       expiry_date DATE NOT NULL,
                       cvv_hash VARCHAR(255) NOT NULL,
                       last_four VARCHAR(4) NOT NULL,
                       created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                       updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                       blocked_at TIMESTAMP WITH TIME ZONE,
                       blocked_reason TEXT
);

ALTER TABLE cards ADD CONSTRAINT chk_card_type CHECK (card_type IN ('virtual', 'physical'));
ALTER TABLE cards ADD CONSTRAINT chk_card_status CHECK (status IN ('active', 'blocked', 'expired', 'cancelled'));
ALTER TABLE cards ADD CONSTRAINT chk_balance CHECK (balance >= 0);

CREATE INDEX idx_cards_company_id ON cards(company_id);
CREATE INDEX idx_cards_card_number ON cards(card_number);
CREATE INDEX idx_cards_status ON cards(status);
CREATE INDEX idx_cards_employee_email ON cards(employee_email);
CREATE INDEX idx_cards_expiry_date ON cards(expiry_date);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS cards;
-- +goose StatementEnd
