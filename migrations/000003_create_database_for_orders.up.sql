CREATE TABLE IF NOT EXISTS orders (
    number 			TEXT 				NOT NULL PRIMARY KEY,
    status 			TEXT        NOT NULL,
		accrual 		NUMERIC,
		uploaded_at TIMESTAMPTZ	NOT NULL,
		user_id     INTEGER     NOT NULL REFERENCES users(id)
);