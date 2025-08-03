-- +goose Up
-- +goose StatementBegin
CREATE TABLE cards_to_issue (
                                id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                client_id UUID NOT NULL,
                                card_id UUID NOT NULL DEFAULT gen_random_uuid(),
                                employee_id UUID NOT NULL,
                                employee_email VARCHAR(255) NOT NULL,
                                status VARCHAR(50) NOT NULL DEFAULT 'pending',
                                created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                                updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE cards_to_issue ADD CONSTRAINT chk_cards_to_issue_status CHECK (status IN ('pending', 'generated'));

CREATE INDEX idx_cards_to_issue_client_id ON cards_to_issue(client_id);
CREATE INDEX idx_cards_to_issue_employee_id ON cards_to_issue(employee_id);
CREATE INDEX idx_cards_to_issue_status ON cards_to_issue(status);
CREATE INDEX idx_cards_to_issue_created_at ON cards_to_issue(created_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS cards_to_issue;
-- +goose StatementEnd
