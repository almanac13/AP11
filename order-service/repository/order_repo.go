package repository

import (
	"database/sql"
	"order-service/domain"
)

type orderRepo struct {
	db *sql.DB
}

func NewOrderRepo(db *sql.DB) *orderRepo {
	return &orderRepo{db: db}
}

func (r *orderRepo) Create(o domain.Order) error {
	_, err := r.db.Exec(
		"INSERT INTO orders (id, customer_id, item_name, amount, status, created_at) VALUES ($1,$2,$3,$4,$5,$6)",
		o.ID, o.CustomerID, o.ItemName, o.Amount, o.Status, o.CreatedAt,
	)
	return err
}

func (r *orderRepo) GetByID(id string) (*domain.Order, error) {
	row := r.db.QueryRow(
		"SELECT id, customer_id, item_name, amount, status, created_at FROM orders WHERE id=$1",
		id,
	)

	var o domain.Order
	err := row.Scan(&o.ID, &o.CustomerID, &o.ItemName, &o.Amount, &o.Status, &o.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &o, nil
}

func (r *orderRepo) GetByCustomerID(customerID string) ([]domain.Order, error) {
	rows, err := r.db.Query(
		"SELECT id, customer_id, item_name, amount, status, created_at FROM orders WHERE customer_id=$1",
		customerID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []domain.Order
	for rows.Next() {
		var o domain.Order
		err := rows.Scan(&o.ID, &o.CustomerID, &o.ItemName, &o.Amount, &o.Status, &o.CreatedAt)
		if err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}

	return orders, nil
}

func (r *orderRepo) UpdateStatus(id string, status string) error {
	_, err := r.db.Exec(
		"UPDATE orders SET status=$1 WHERE id=$2",
		status, id,
	)
	return err
}
