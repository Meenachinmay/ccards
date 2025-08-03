-- +goose Up
-- +goose StatementBegin
CREATE TRIGGER update_cards_to_issue_updated_at BEFORE UPDATE ON cards_to_issue
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS update_cards_to_issue_updated_at ON cards_to_issue;
-- +goose StatementEnd