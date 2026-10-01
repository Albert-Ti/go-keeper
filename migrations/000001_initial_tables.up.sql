CREATE TABLE
  IF NOT EXISTS users (
    uuid UUID PRIMARY KEY DEFAULT uuidv7 (),
    email VARCHAR(255) UNIQUE NOT NULL,
    email_code VARCHAR(32),
    is_confirm_email BOOLEAN DEFAULT FALSE,
    pass VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
  );

CREATE TABLE
  IF NOT EXISTS pass_list (
    id BIGSERIAL PRIMARY KEY,
    user_uuid UUID NOT NULL,
    old_pass VARCHAR(255) NOT NULL,
    FOREIGN KEY (user_uuid) REFERENCES users (uuid)
  );

CREATE TABLE
  IF NOT EXISTS bank_cards (
    id BIGSERIAL PRIMARY KEY,
    user_uuid UUID NOT NULL,
    card_number BYTEA UNIQUE NOT NULL,
    expiry_date DATE NOT NULL,
    active BOOLEAN,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_uuid) REFERENCES users (uuid)
  );

CREATE TABLE
  IF NOT EXISTS arbitrary_data (
    id BIGSERIAL PRIMARY KEY,
    user_uuid UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(255) NOT NULL,
    status SMALLINT NOT NULL DEFAULT 0,
    object_key VARCHAR(255) NOT NULL,
    client_size BIGINT NOT NULL,
    total_size BIGINT DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_uuid) REFERENCES users (uuid)
  );