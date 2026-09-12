-- +goose Up
-- +goose StatementBegin
CREATE TABLE
  IF NOT EXISTS cards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL, -- Foreign key to users table
    token VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL,
    linkedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last4 VARCHAR(24) NOT NULL
  );
CREATE TABLE 
  IF NOT EXISTS payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL UNIQUE, --  Foreign key to orders table
    card_id UUID NOT NULL,  -- Foreign key to cards table
    owner_id UUID NOT NULL, -- Foreign key to users table
    amount BIGINT NOT NULL,
    currency VARCHAR(3) NOT NULL,
    status VARCHAR(20) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS payments;
DROP TABLE IF EXISTS cards;
-- +goose StatementEnd
