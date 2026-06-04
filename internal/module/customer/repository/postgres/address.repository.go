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

type addressRepository struct {
	db *pgxpool.Pool
}

func NewAddressRepository(db *pgxpool.Pool) repository.AddressRepository {
	return &addressRepository{db: db}
}

func scanAddress(rows interface{ Scan(...any) error }) (*models.Address, error) {
	var a models.Address
	if err := rows.Scan(
		&a.AddressNo, &a.AddressId, &a.CustomerNo,
		&a.Label, &a.AddressLine1, &a.AddressLine2,
		&a.City, &a.State, &a.Country, &a.PostalCode,
		&a.IsDefault, &a.Status, &a.CreatedAt, &a.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &a, nil
}

const addressColumns = `
	address_no, address_id, customer_no,
	label, address_line1, address_line2,
	city, state, country, postal_code,
	is_default, status, created_at, updated_at
`

func (r *addressRepository) GetAllByCustomerNo(ctx context.Context, customerNo uint32) ([]models.Address, error) {
	rows, err := r.db.Query(ctx, `
		SELECT `+addressColumns+`
		FROM customer.addresses
		WHERE customer_no = $1 AND status != 'D'
		ORDER BY is_default DESC, created_at ASC
	`, customerNo)
	if err != nil {
		return nil, fmt.Errorf("addressRepository.GetAllByCustomerNo: %w", err)
	}
	defer rows.Close()

	var addresses []models.Address
	for rows.Next() {
		a, err := scanAddress(rows)
		if err != nil {
			return nil, fmt.Errorf("addressRepository.GetAllByCustomerNo: scan: %w", err)
		}
		addresses = append(addresses, *a)
	}
	return addresses, rows.Err()
}

func (r *addressRepository) GetByNo(ctx context.Context, addressNo uint32) (*models.Address, error) {
	a, err := scanAddress(r.db.QueryRow(ctx, `
		SELECT `+addressColumns+`
		FROM customer.addresses
		WHERE address_no = $1 AND status != 'D'
	`, addressNo))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("addressRepository.GetByNo: %w", err)
	}
	return a, nil
}

func (r *addressRepository) GetById(ctx context.Context, Id uuid.UUID) (*models.Address, error) {
	a, err := scanAddress(r.db.QueryRow(ctx, `
		SELECT `+addressColumns+`
		FROM customer.addresses
		WHERE address_id = $1 AND status != 'D'
	`, Id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("addressRepository.GetById: %w", err)
	}
	return a, nil
}

func (r *addressRepository) Save(ctx context.Context, address *models.Address) (*models.Address, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("addressRepository.Save: begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	// If this is default — unset existing default first
	if address.IsDefault {
		_, err = tx.Exec(ctx, `
			UPDATE customer.addresses
			SET is_default = FALSE, updated_at = NOW()
			WHERE customer_no = $1 AND is_default = TRUE
		`, address.CustomerNo)
		if err != nil {
			return nil, fmt.Errorf("addressRepository.Save: unset default: %w", err)
		}
	}

	a, err := scanAddress(tx.QueryRow(ctx, `
		INSERT INTO customer.addresses (
			address_id, customer_no, label,
			address_line1, address_line2,
			city, state, country, postal_code,
			is_default, status
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING `+addressColumns,
		uuid.New(), address.CustomerNo, address.Label,
		address.AddressLine1, address.AddressLine2,
		address.City, address.State, address.Country, address.PostalCode,
		address.IsDefault, models.StatusActive,
	))
	if err != nil {
		return nil, fmt.Errorf("addressRepository.Save: insert: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("addressRepository.Save: commit: %w", err)
	}
	return a, nil
}

func (r *addressRepository) Modify(ctx context.Context, Id uuid.UUID, req *models.UpdateAddressRequest) (*models.Address, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("addressRepository.Modify: begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	// Lock row
	var lockedNo uint32
	err = tx.QueryRow(ctx, `
		SELECT address_no FROM customer.addresses
		WHERE address_id = $1 AND status != 'D'
		FOR UPDATE
	`, Id).Scan(&lockedNo)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("addressRepository.Modify: lock row: %w", err)
	}

	// Unset existing default if this becomes default
	if req.IsDefault != nil && *req.IsDefault {
		var customerNo uint32
		_ = r.db.QueryRow(ctx, `SELECT customer_no FROM customer.addresses WHERE address_id = $1`, Id).Scan(&customerNo)
		_, err = tx.Exec(ctx, `
			UPDATE customer.addresses
			SET is_default = FALSE, updated_at = NOW()
			WHERE customer_no = $1 AND is_default = TRUE AND address_id != $2
		`, customerNo, Id)
		if err != nil {
			return nil, fmt.Errorf("addressRepository.Modify: unset default: %w", err)
		}
	}

	a, err := scanAddress(tx.QueryRow(ctx, `
		UPDATE customer.addresses
		SET
			label         = COALESCE(NULLIF($1, ''), label),
			address_line1 = COALESCE(NULLIF($2, ''), address_line1),
			address_line2 = COALESCE($3, address_line2),
			city          = COALESCE(NULLIF($4, ''), city),
			state         = COALESCE(NULLIF($5, ''), state),
			country       = COALESCE(NULLIF($6, ''), country),
			postal_code   = COALESCE($7, postal_code),
			is_default    = COALESCE($8, is_default),
			updated_at    = NOW()
		WHERE address_id  = $9
		RETURNING `+addressColumns,
		req.Label, req.AddressLine1, req.AddressLine2,
		req.City, req.State, req.Country, req.PostalCode,
		req.IsDefault, Id,
	))
	if err != nil {
		return nil, fmt.Errorf("addressRepository.Modify: update: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("addressRepository.Modify: commit: %w", err)
	}
	return a, nil
}

func (r *addressRepository) Remove(ctx context.Context, Id uuid.UUID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("addressRepository.Remove: begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	var lockedNo uint32
	err = tx.QueryRow(ctx, `
		SELECT address_no FROM customer.addresses
		WHERE address_id = $1 AND status != 'D'
		FOR UPDATE
	`, Id).Scan(&lockedNo)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return pgx.ErrNoRows
		}
		return fmt.Errorf("addressRepository.Remove: lock row: %w", err)
	}

	_, err = tx.Exec(ctx, `
		UPDATE customer.addresses
		SET status = 'D', updated_at = NOW()
		WHERE address_id = $1
	`, Id)
	if err != nil {
		return fmt.Errorf("addressRepository.Remove: soft delete: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("addressRepository.Remove: commit: %w", err)
	}
	return nil
}