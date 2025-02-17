CREATE TABLE IF NOT EXISTS users (
                                     id SERIAL PRIMARY KEY,
                                     username VARCHAR(255) NOT NULL UNIQUE,
                                     coins INT NOT NULL DEFAULT 1000 CHECK (coins >= 0)
);

CREATE TABLE IF NOT EXISTS transactions (
                                            id SERIAL PRIMARY KEY,
                                            sender_id INT REFERENCES users(id) ON DELETE SET NULL,
                                            receiver_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                                            amount INT NOT NULL CHECK (amount > 0),
                                            created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS inventory (
                                         user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                                         item_name VARCHAR(50) NOT NULL,
                                         quantity INT NOT NULL DEFAULT 1 CHECK (quantity > 0),
                                         PRIMARY KEY (user_id, item_name)
);

CREATE TABLE IF NOT EXISTS auth (
                       id SERIAL PRIMARY KEY,
                       username VARCHAR(255) UNIQUE NOT NULL,
                       password_hash VARCHAR(255) NOT NULL,
                       created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS merch (
                                     id SERIAL PRIMARY KEY,
                                     name VARCHAR(255) UNIQUE NOT NULL,
                                     price INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_transactions_sender ON transactions(sender_id);
CREATE INDEX IF NOT EXISTS idx_transactions_receiver ON transactions(receiver_id);
CREATE INDEX IF NOT EXISTS idx_inventory_user ON inventory(user_id);