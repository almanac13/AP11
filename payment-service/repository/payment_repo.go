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
		"INSERT INTO payments (id, order_id, transaction_id, amount, status, created_at) VALUES ($1,$2,$3,$4,$5,$6)",
		p.ID, p.OrderID, p.TransactionID, p.Amount, p.Status, p.CreatedAt,
	)
	return err
}

func (r *paymentRepo) GetByOrderID(orderID string) (*domain.Payment, error) {
	row := r.db.QueryRow(
		"SELECT id, order_id, transaction_id, amount, status, created_at FROM payments WHERE order_id=$1",
		orderID,
	)

	var p domain.Payment
	err := row.Scan(&p.ID, &p.OrderID, &p.TransactionID, &p.Amount, &p.Status, &p.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &p, nil
}

func (r *paymentRepo) ListByStatus(status string) ([]domain.Payment, error) {
	rows, err := r.db.Query(
		"SELECT id, order_id, transaction_id, amount, status, created_at FROM payments WHERE status=$1 ORDER BY created_at DESC",
		status,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []domain.Payment

	for rows.Next() {
		var p domain.Payment
		err := rows.Scan(&p.ID, &p.OrderID, &p.TransactionID, &p.Amount, &p.Status, &p.CreatedAt)
		if err != nil {
			return nil, err
		}
		payments = append(payments, p)
	}

	return payments, nil
}
