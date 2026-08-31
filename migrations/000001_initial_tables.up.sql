CREATE TABLE
  IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuidv7 (),
    email VARCHAR(255) UNIQUE NOT NULL,
    email_token VARCHAR(32),
    is_confirm_email BOOLEAN DEFAULT FALSE,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
  );

CREATE TABLE
  IF NOT EXISTS history (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID NOT NULL,
    old_password VARCHAR(255) NOT NULL,
    life_time INTERVAL,
    FOREIGN KEY (user_id) REFERENCES users (id)
  );

CREATE TABLE
  IF NOT EXISTS bank_cards (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID NOT NULL,
    card_number VARCHAR(19) NOT NULL,
    expiry_date DATE NOT NULL,
    active BOOLEAN,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users (id)
  );