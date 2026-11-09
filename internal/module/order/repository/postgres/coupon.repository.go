package postgres

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/chand-magar/gamenglam-server/internal/module/order/models"
)

type CouponRepository struct {
	db *pgxpool.Pool
}

func NewCouponRepository(db *pgxpool.Pool) *CouponRepository {
	return &CouponRepository{db: db}
}

func (r *CouponRepository) GetAll(ctx context.Context, req *models.CouponListRequest) (*models.CouponListResponse, error) {
	page, limit := req.Page, req.Limit
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	args := []any{}
	where := []string{"status != 'D'"}
	i := 1

	if req.Status != "" && req.Status != "all" {
		where = append(where, fmt.Sprintf("status = $%d", i))
		args = append(args, req.Status)
		i++
	}
	if req.Search != "" {
		where = append(where, fmt.Sprintf("(UPPER(code) LIKE UPPER($%d) OR description ILIKE $%d)", i, i))
		args = append(args, "%"+req.Search+"%")
		i++
	}

	whereClause := "WHERE " + strings.Join(where, " AND ")

	var total int
	countQ := fmt.Sprintf(`SELECT COUNT(*) FROM orders.coupons %s`, whereClause)
	if err := r.db.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("couponRepository.GetAll count: %w", err)
	}

	args = append(args, limit, offset)
	listQ := fmt.Sprintf(`
		SELECT coupon_no, coupon_id, code, COALESCE(description,''), discount_type,
		       discount_value, min_order_amount, max_uses, used_count,
		       valid_from, valid_until, status, created_at, updated_at
		FROM orders.coupons
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, i, i+1)

	rows, err := r.db.Query(ctx, listQ, args...)
	if err != nil {
		return nil, fmt.Errorf("couponRepository.GetAll query: %w", err)
	}
	defer rows.Close()

	var items []models.Coupon
	for rows.Next() {
		var c models.Coupon
		if err := rows.Scan(
			&c.CouponNo, &c.CouponId, &c.Code, &c.Description,
			&c.DiscountType, &c.DiscountValue, &c.MinOrderAmount,
			&c.MaxUses, &c.UsedCount, &c.ValidFrom, &c.ValidUntil,
			&c.Status, &c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("couponRepository.GetAll scan: %w", err)
		}
		items = append(items, c)
	}
	if items == nil {
		items = []models.Coupon{}
	}

	return &models.CouponListResponse{
		Items:      items,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: int(math.Ceil(float64(total) / float64(limit))),
	}, nil
}

func (r *CouponRepository) GetById(ctx context.Context, couponId uuid.UUID) (*models.Coupon, error) {
	var c models.Coupon
	err := r.db.QueryRow(ctx, `
		SELECT coupon_no, coupon_id, code, COALESCE(description,''), discount_type,
		       discount_value, min_order_amount, max_uses, used_count,
		       valid_from, valid_until, status, created_at, updated_at
		FROM orders.coupons
		WHERE coupon_id = $1 AND status != 'D'
	`, couponId).Scan(
		&c.CouponNo, &c.CouponId, &c.Code, &c.Description,
		&c.DiscountType, &c.DiscountValue, &c.MinOrderAmount,
		&c.MaxUses, &c.UsedCount, &c.ValidFrom, &c.ValidUntil,
		&c.Status, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("couponRepository.GetById: %w", err)
	}
	return &c, nil
}

func (r *CouponRepository) Save(ctx context.Context, req *models.CreateCouponRequest) (*models.Coupon, error) {
	var c models.Coupon
	err := r.db.QueryRow(ctx, `
		INSERT INTO orders.coupons
		    (code, description, discount_type, discount_value, min_order_amount,
		     max_uses, valid_from, valid_until)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING coupon_no, coupon_id, code, COALESCE(description,''), discount_type,
		          discount_value, min_order_amount, max_uses, used_count,
		          valid_from, valid_until, status, created_at, updated_at
	`,
		strings.ToUpper(req.Code), req.Description, req.DiscountType, req.DiscountValue,
		req.MinOrderAmount, req.MaxUses, req.ValidFrom, req.ValidUntil,
	).Scan(
		&c.CouponNo, &c.CouponId, &c.Code, &c.Description,
		&c.DiscountType, &c.DiscountValue, &c.MinOrderAmount,
		&c.MaxUses, &c.UsedCount, &c.ValidFrom, &c.ValidUntil,
		&c.Status, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("couponRepository.Save: %w", err)
	}
	return &c, nil
}

func (r *CouponRepository) Update(ctx context.Context, couponId uuid.UUID, req *models.UpdateCouponRequest) (*models.Coupon, error) {
	sets := []string{"updated_at = NOW()"}
	args := []any{}
	i := 1

	if req.Description != nil {
		sets = append(sets, fmt.Sprintf("description = $%d", i))
		args = append(args, *req.Description)
		i++
	}
	if req.DiscountType != nil {
		sets = append(sets, fmt.Sprintf("discount_type = $%d", i))
		args = append(args, *req.DiscountType)
		i++
	}
	if req.DiscountValue != nil {
		sets = append(sets, fmt.Sprintf("discount_value = $%d", i))
		args = append(args, *req.DiscountValue)
		i++
	}
	if req.MinOrderAmount != nil {
		sets = append(sets, fmt.Sprintf("min_order_amount = $%d", i))
		args = append(args, *req.MinOrderAmount)
		i++
	}
	if req.MaxUses != nil {
		sets = append(sets, fmt.Sprintf("max_uses = $%d", i))
		args = append(args, *req.MaxUses)
		i++
	}
	if req.ValidFrom != nil {
		sets = append(sets, fmt.Sprintf("valid_from = $%d", i))
		args = append(args, *req.ValidFrom)
		i++
	}
	if req.ValidUntil != nil {
		sets = append(sets, fmt.Sprintf("valid_until = $%d", i))
		args = append(args, *req.ValidUntil)
		i++
	}
	if req.Status != nil {
		sets = append(sets, fmt.Sprintf("status = $%d", i))
		args = append(args, *req.Status)
		i++
	}

	args = append(args, couponId)
	query := fmt.Sprintf(`
		UPDATE orders.coupons
		SET %s
		WHERE coupon_id = $%d AND status != 'D'
		RETURNING coupon_no, coupon_id, code, COALESCE(description,''), discount_type,
		          discount_value, min_order_amount, max_uses, used_count,
		          valid_from, valid_until, status, created_at, updated_at
	`, strings.Join(sets, ", "), i)

	var c models.Coupon
	err := r.db.QueryRow(ctx, query, args...).Scan(
		&c.CouponNo, &c.CouponId, &c.Code, &c.Description,
		&c.DiscountType, &c.DiscountValue, &c.MinOrderAmount,
		&c.MaxUses, &c.UsedCount, &c.ValidFrom, &c.ValidUntil,
		&c.Status, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("couponRepository.Update: %w", err)
	}
	return &c, nil
}

func (r *CouponRepository) Delete(ctx context.Context, couponId uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		`UPDATE orders.coupons SET status = 'D', updated_at = NOW() WHERE coupon_id = $1`,
		couponId,
	)
	if err != nil {
		return fmt.Errorf("couponRepository.Delete: %w", err)
	}
	return nil
}
