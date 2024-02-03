CREATE TABLE IF NOT EXISTS users (		
    user_id VARCHAR PRIMARY KEY CHECK(user_id <> ''),
	hash VARCHAR NOT NULL CHECK (hash <> '')
);
	
CREATE TABLE IF NOT EXISTS orders (
	number VARCHAR PRIMARY KEY CHECK(number <> ''),
	user_id VARCHAR NOT NULL CHECK(user_id <> ''),
	uploaded_at timestamp NOT NULL,
	FOREIGN KEY (user_id) REFERENCES users(user_id)
);
	
CREATE TABLE IF NOT EXISTS billing (
	order_number VARCHAR NOT NULL CHECK(order_number <> ''),
	status VARCHAR NOT NULL,
	accrual int, 
	uploaded_at timestamp NOT NULL,
	time timestamp NOT NULL,
	FOREIGN KEY (order_number) REFERENCES orders(number),
	CONSTRAINT unique_order_number_status UNIQUE (order_number, status)
);	