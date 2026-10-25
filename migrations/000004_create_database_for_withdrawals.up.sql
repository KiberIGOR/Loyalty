CREATE TABLE IF NOT EXISTS withdrawals (
    order_number  TEXT  			NOT NULL PRIMARY KEY,
		sum 					NUMERIC,
		processed_at  TIMESTAMPTZ NOT NULL,
		user_id     	INTEGER     NOT NULL REFERENCES users(id)
);