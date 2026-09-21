package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/chand-magar/gamenglam-server/internal/module/customer/models"
	"github.com/chand-magar/gamenglam-server/internal/module/customer/repository"
)

type customerRepository struct {
	db *pgxpool.Pool
}

func NewCustomerRepository(db *pgxpool.Pool) repository.CustomerRepository {
	return &customerRepository{db: db}
}

// ── READ operations ───────────────────────────────────────────────────────────

func (r *customerRepository) GetAll(ctx context.Context) ([]models.Customer, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			customer_no, customer_id, first_name, last_name,
			email_id, gender, dob, mobile_no, status, created_at, updated_at
		FROM customer.customers
		WHERE status != 'D'
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("customerRepository.GetAll: %w", err)
	}
	defer rows.Close()

	var data []models.Customer
	for rows.Next() {
		var d models.Customer
		if err := rows.Scan(
			&d.CustomerNo, &d.CustomerId, &d.FirstName, &d.LastName,
			&d.EmailId, &d.Gender, &d.Dob, &d.MobileNo,
			&d.Status, &d.CreatedAt, &d.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("customerRepository.GetAll: scan: %w", err)
		}
		data = append(data, d)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("customerRepository.GetAll: rows: %w", err)
	}

	return data, nil
}

func (r *customerRepository) GetById(ctx context.Context, Id uuid.UUID) (*models.Customer, error) {
	var d models.Customer
	err := r.db.QueryRow(ctx, `
		SELECT
			customer_no, customer_id, first_name, last_name,
			email_id, gender, dob, mobile_no, status, created_at, updated_at
		FROM customer.customers
		WHERE customer_id = $1 AND status != 'D'
	`, Id).Scan(
		&d.CustomerNo, &d.CustomerId, &d.FirstName, &d.LastName,
		&d.EmailId, &d.Gender, &d.Dob, &d.MobileNo,
		&d.Status, &d.CreatedAt, &d.UpdatedAt, // ← was missing updated_at
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("customerRepository.GetById: %w", err)
	}

	return &d, nil
}

func (r *customerRepository) GetByEmail(ctx context.Context, email string) (*models.Customer, error) {
	var d models.Customer
	err := r.db.QueryRow(ctx, `
		SELECT
			customer_no, customer_id, first_name, last_name,
			email_id, gender, dob, mobile_no, status, created_at, updated_at
		FROM customer.customers
		WHERE email_id = $1 AND status != 'D'
	`, email).Scan(
		&d.CustomerNo, &d.CustomerId, &d.FirstName, &d.LastName,
		&d.EmailId, &d.Gender, &d.Dob, &d.MobileNo,
		&d.Status, &d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("customerRepository.GetByEmail: %w", err)
	}

	return &d, nil
}

func (r *customerRepository) GetByCustomerNo(ctx context.Context, customerNo uint32) (*models.Customer, error) {
	var d models.Customer
	err := r.db.QueryRow(ctx, `
		SELECT
			customer_no, customer_id, first_name, last_name,
			email_id, gender, dob, mobile_no, status, created_at, updated_at
		FROM customer.customers
		WHERE customer_no = $1 AND status != 'D'
	`, customerNo).Scan(
		&d.CustomerNo, &d.CustomerId, &d.FirstName, &d.LastName,
		&d.EmailId, &d.Gender, &d.Dob, &d.MobileNo,
		&d.Status, &d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("customerRepository.GetByCustomerNo: %w", err)
	}

	return &d, nil
}

// ── WRITE operations (ACID — transaction + row lock) ──────────────────────────

func (r *customerRepository) Save(ctx context.Context, customer *models.Customer) (*models.Customer, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("customerRepository.Save: begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	var d models.Customer
	err = tx.QueryRow(ctx, `
		INSERT INTO customer.customers (
			customer_id, first_name, last_name, email_id,
			gender, dob, mobile_no, status
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING
			customer_no, customer_id, first_name, last_name,
			email_id, gender, dob, mobile_no, status, created_at, updated_at
	`, // ← removed trailing comma after last column in INSERT
		customer.CustomerId,
		customer.FirstName,
		customer.LastName,
		customer.EmailId,
		customer.Gender,
		customer.Dob,
		customer.MobileNo,
		models.StatusActive,
	).Scan(
		&d.CustomerNo, &d.CustomerId, &d.FirstName, &d.LastName,
		&d.EmailId, &d.Gender, &d.Dob, &d.MobileNo,
		&d.Status, &d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("customerRepository.Save: insert: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("customerRepository.Save: commit: %w", err)
	}

	return &d, nil
}

func (r *customerRepository) Modify(ctx context.Context, Id uuid.UUID, req *models.UpdateCustomerRequest) (*models.Customer, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("customerRepository.Modify: begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	var lockedNo uint32
	err = tx.QueryRow(ctx, `
		SELECT customer_no FROM customer.customers
		WHERE customer_id = $1 AND status != 'D'
		FOR UPDATE
	`, Id).Scan(&lockedNo)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("customerRepository.Modify: lock row: %w", err)
	}

	var d models.Customer
	err = tx.QueryRow(ctx, `
		UPDATE customer.customers
		SET
			first_name = COALESCE(NULLIF($1, ''), first_name),
			last_name  = COALESCE(NULLIF($2, ''), last_name),
			email_id   = COALESCE(NULLIF($3, ''), email_id),
			gender     = COALESCE(NULLIF($4, ''), gender),
			mobile_no  = COALESCE(NULLIF($5, ''), mobile_no),
			updated_at = NOW()
		WHERE customer_id = $6
		RETURNING
			customer_no, customer_id, first_name, last_name,
			email_id, gender, dob, mobile_no, status, created_at, updated_at
	`, // ← fixed: separate params for first_name/$1 and last_name/$2, removed trailing comma
		req.FirstName,
		req.LastName,
		req.EmailId,
		req.Gender,
		req.MobileNo,
		Id,
	).Scan(
		&d.CustomerNo, &d.CustomerId, &d.FirstName, &d.LastName,
		&d.EmailId, &d.Gender, &d.Dob, &d.MobileNo,
		&d.Status, &d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("customerRepository.Modify: update: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("customerRepository.Modify: commit: %w", err)
	}

	return &d, nil
}

func (r *customerRepository) Remove(ctx context.Context, Id uuid.UUID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("customerRepository.Remove: begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	var lockedNo uint32
	err = tx.QueryRow(ctx, `
		SELECT customer_no FROM customer.customers
		WHERE customer_id = $1 AND status != 'D'
		FOR UPDATE
	`, Id).Scan(&lockedNo)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return pgx.ErrNoRows
		}
		return fmt.Errorf("customerRepository.Remove: lock row: %w", err)
	}

	_, err = tx.Exec(ctx, `
		UPDATE customer.customers
		SET status = 'D', updated_at = NOW()
		WHERE customer_id = $1
	`, Id)
	if err != nil {
		return fmt.Errorf("customerRepository.Remove: soft delete: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("customerRepository.Remove: commit: %w", err)
	}

	return nil
}
