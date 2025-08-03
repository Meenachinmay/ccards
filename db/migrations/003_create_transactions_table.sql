-- +goose Up
-- +goose StatementBegin
CREATE TABLE transactions (
                              id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                              card_id UUID NOT NULL REFERENCES cards(id) ON DELETE CASCADE,
                              company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
                              transaction_type VARCHAR(50) NOT NULL CHECK (transaction_type IN ('purchase', 'charge')),
                              amount DECIMAL(15, 2) NOT NULL,
                              merchant_name VARCHAR(255),
                              merchant_category VARCHAR(100),
                              description TEXT DEFAULT 'company purchase',
                              status VARCHAR(50) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'completed', 'failed')),
                              processed_at TIMESTAMP WITH TIME ZONE,
                              created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                              updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_transactions_card_id ON transactions(card_id);
CREATE INDEX idx_transactions_company_id ON transactions(company_id);
CREATE INDEX idx_transactions_status ON transactions(status);
CREATE INDEX idx_transactions_created_at ON transactions(created_at);
CREATE INDEX idx_transactions_type ON transactions(transaction_type);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS transactions;
-- +goose StatementEnd