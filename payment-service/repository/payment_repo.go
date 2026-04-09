package repository

import (
	"database/sql"
	"payment-service/domain"
)

type paymentRepo struct {
	db *sql.DB
}

func NewPaymentRepo(db *sql.DB) *paymentRepo {
	return &paymentRepo{db: db}
}

func (r *paymentRepo) Create(p domain.Payment) error {
	_, err := r.db.Exec(
		"INSERT INTO payments (id, order_id, transaction_id, amount, status) VALUES ($1,$2,$3,$4,$5)",
		p.ID, p.OrderID, p.TransactionID, p.Amount, p.Status,
	)
	return err
}

func (r *paymentRepo) GetByOrderID(orderID string) (*domain.Payment, error) {
	row := r.db.QueryRow(
		"SELECT id, order_id, transaction_id, amount, status FROM payments WHERE order_id=$1",
		orderID,
	)

	var p domain.Payment
	err := row.Scan(&p.ID, &p.OrderID, &p.TransactionID, &p.Amount, &p.Status)
	if err != nil {
		return nil, err
	}

	return &p, nil
}
