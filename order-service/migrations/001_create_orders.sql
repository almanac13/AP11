CREATE TABLE orders (
    id TEXT PRIMARY KEY,
    customer_id TEXT,
    item_name TEXT,
    amount BIGINT,
    status TEXT,
    created_at TIMESTAMP
);