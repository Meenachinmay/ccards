-- +goose Up
-- +goose StatementBegin
CREATE TABLE companies (
                           id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                           client_id UUID NOT NULL,
                           name VARCHAR(255) NOT NULL,
                           email VARCHAR(255) NOT NULL UNIQUE,
                           password VARCHAR(255) NOT NULL,
                           address TEXT,
                           phone VARCHAR(50),
                           status VARCHAR(50) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'suspended', 'inactive')),
                           created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                           updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_companies_email ON companies(email);
CREATE INDEX idx_companies_client_id ON companies(client_id);
CREATE INDEX idx_companies_status ON companies(status);
CREATE INDEX idx_companies_created_at ON companies(created_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS companies;
-- +goose StatementEnd