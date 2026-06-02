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
	query := `
		SELECT
			customer_no,
			customer_id,
			fullname,
			email_id,
			gender,
			dob,
			mobile_no,
			address,
			status,
			created_at,
			created_by,
			updated_at,
			updated_by
		FROM customer.customers
		WHERE status != 'D'
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("customerRepository.GetAll: %w", err)
	}
	defer rows.Close()

	var data []models.Customer
	for rows.Next() {
		var d models.Customer
		if err := rows.Scan(
			&d.CustomerNo,
			&d.CustomerId,
			&d.FullName,
			&d.EmailId,
			&d.Gender,
			&d.Dob,
			&d.MobileNo,
			&d.Address,
			&d.Status,
			&d.CreatedAt,
			&d.CreatedBy,
			&d.UpdatedAt,
			&d.UpdatedBy,
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
	query := `
		SELECT
			customer_no,
			customer_id,
			fullname,
			email_id,
			gender,
			dob,
			mobile_no,
			address,
			status,
			created_at,
			created_by,
			updated_at,
			updated_by
		FROM customer.customers
		WHERE customer_id = $1
		AND status != 'D'
	`

	var d models.Customer
	err := r.db.QueryRow(ctx, query, Id).Scan(
		&d.CustomerNo,
		&d.CustomerId,
		&d.FullName,
		&d.EmailId,
		&d.Gender,
		&d.Dob,
		&d.MobileNo,
		&d.Address,
		&d.Status,
		&d.CreatedAt,
		&d.CreatedBy,
		&d.UpdatedAt,
		&d.UpdatedBy,
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
	query := `
		SELECT
			customer_no,
			customer_id,
			fullname,
			email_id,
			gender,
			dob,
			mobile_no,
			address,
			status,
			created_at,
			created_by,
			updated_at,
			updated_by
		FROM customer.customers
		WHERE email_id = $1
		AND status != 'D'
	`

	var d models.Customer
	err := r.db.QueryRow(ctx, query, email).Scan(
		&d.CustomerNo,
		&d.CustomerId,
		&d.FullName,
		&d.EmailId,
		&d.Gender,
		&d.Dob,
		&d.MobileNo,
		&d.Address,
		&d.Status,
		&d.CreatedAt,
		&d.CreatedBy,
		&d.UpdatedAt,
		&d.UpdatedBy,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("customerRepository.GetByEmail: %w", err)
	}

	return &d, nil
}

// ── WRITE operations (ACID — transaction + row lock) ──────────────────────────

func (r *customerRepository) Save(ctx context.Context, user *models.Customer) (*models.Customer, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("customerRepository.Save: begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	query := `
		INSERT INTO customer.customers (
			customer_id,
			fullname,
			email_id,
			gender,
			dob,
			mobile_no,
			address,
			status,
			created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING
			customer_no,
			customer_id,
			fullname,
			email_id,
			gender,
			dob,
			mobile_no,
			address,
			status,
			created_at,
			created_by,
			updated_at,
			updated_by
	`

	var d models.Customer
	err = tx.QueryRow(ctx, query,
		user.CustomerId,
		user.FullName,
		user.EmailId,
		user.Gender,
		user.Dob,
		user.MobileNo,
		user.Address,
		models.StatusActive,
		user.CreatedBy,
	).Scan(
		&d.CustomerNo,
		&d.CustomerId,
		&d.FullName,
		&d.EmailId,
		&d.Gender,
		&d.Dob,
		&d.MobileNo,
		&d.Address,
		&d.Status,
		&d.CreatedAt,
		&d.CreatedBy,
		&d.UpdatedAt,
		&d.UpdatedBy,
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

	// Lock row by customer_id — prevents lost updates
	var lockedNo uint32
	err = tx.QueryRow(ctx, `
		SELECT customer_no FROM customer.customers
		WHERE customer_id = $1 AND status != 'D'
	`, Id).Scan(&lockedNo)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("customerRepository.Modify: lock row: %w", err)
	}

	query := `
		UPDATE customer.customers
		SET
			fullname = COALESCE(NULLIF($1, ''), fullname),
			email_id      = COALESCE(NULLIF($2, ''), email_id),
			gender        = COALESCE(NULLIF($3, ''), gender),
			mobile_no     = COALESCE(NULLIF($4, ''), mobile_no),
			address       = CASE WHEN $5::text = '' THEN address ELSE $5::jsonb END,
			updated_at    = NOW(),
			updated_by    = $6
		WHERE customer_id  = $7
		RETURNING
			customer_no,
			customer_id,
			fullname,
			email_id,
			gender,
			dob,
			mobile_no,
			address,
			status,
			created_at,
			created_by,
			updated_at,
			updated_by
	`

	var d models.Customer
	err = tx.QueryRow(ctx, query,
		req.FullName,
		req.EmailId,
		req.Gender,
		req.MobileNo,
		req.Address,
		req.UpdatedBy,
		Id,
	).Scan(
		&d.CustomerNo,
		&d.CustomerId,
		&d.FullName,
		&d.EmailId,
		&d.Gender,
		&d.Dob,
		&d.MobileNo,
		&d.Address,
		&d.Status,
		&d.CreatedAt,
		&d.CreatedBy,
		&d.UpdatedAt,
		&d.UpdatedBy,
	)
	if err != nil {
		return nil, fmt.Errorf("customerRepository.Modify: Modify: %w", err)
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

	// Lock row by customer_id before soft delete
	var lockedNo uint32
	err = tx.QueryRow(ctx, `
		SELECT customer_no FROM customer.customers
		WHERE customer_id = $1 AND status != 'D'
	`, Id).Scan(&lockedNo)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return pgx.ErrNoRows
		}
		return fmt.Errorf("customerRepository.Remove: lock row: %w", err)
	}

	_, err = tx.Exec(ctx, `
		UPDATE customer.customers
		SET
			status     = 'D',
			updated_at = NOW()
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