package postgres

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/chand-magar/gamenglam-server/internal/module/catalog/models"
	repository "github.com/chand-magar/gamenglam-server/internal/module/catalog/repository"
)

type reviewRepository struct {
	db *pgxpool.Pool
}

func NewReviewRepository(db *pgxpool.Pool) repository.ReviewRepository {
	return &reviewRepository{db: db}
}

func (r *reviewRepository) GetByProductId(ctx context.Context, productId uuid.UUID) ([]models.Review, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			pr.review_no,
			pr.review_id,
			pr.product_no,
			p.product_id,
			pr.customer_no,
			c.first_name || ' ' || c.last_name AS customer_name,
			pr.rating,
			COALESCE(pr.title, '') AS title,
			pr.body,
			pr.created_at
		FROM catalog.product_reviews pr
		JOIN catalog.products p        ON p.product_no  = pr.product_no
		JOIN customer.customers c      ON c.customer_no = pr.customer_no
		WHERE p.product_id = $1
		  AND pr.status != 'D'
		ORDER BY pr.created_at DESC
	`, productId)
	if err != nil {
		return nil, fmt.Errorf("reviewRepository.GetByProductId: %w", err)
	}
	defer rows.Close()

	var reviews []models.Review
	for rows.Next() {
		var rv models.Review
		if err := rows.Scan(
			&rv.ReviewNo,
			&rv.ReviewId,
			&rv.ProductNo,
			&rv.ProductId,
			&rv.CustomerNo,
			&rv.CustomerName,
			&rv.Rating,
			&rv.Title,
			&rv.Body,
			&rv.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("reviewRepository.GetByProductId scan: %w", err)
		}
		reviews = append(reviews, rv)
	}
	return reviews, nil
}

func (r *reviewRepository) Save(ctx context.Context, review *models.Review) (*models.Review, error) {
	// Resolve product_id → product_no
	var productNo uint32
	err := r.db.QueryRow(ctx, `
		SELECT product_no FROM catalog.products WHERE product_id = $1 AND status != 'D'
	`, review.ProductId).Scan(&productNo)
	if err != nil {
		return nil, fmt.Errorf("reviewRepository.Save: resolve product_no: %w", err)
	}
	review.ProductNo = productNo

	var reviewNo uint32
	err = r.db.QueryRow(ctx, `
		INSERT INTO catalog.product_reviews
			(review_id, product_no, customer_no, rating, title, body, status)
		VALUES ($1, $2, $3, $4, NULLIF($5, ''), $6, 'A')
		RETURNING review_no
	`,
		review.ReviewId,
		review.ProductNo,
		review.CustomerNo,
		review.Rating,
		review.Title,
		review.Body,
	).Scan(&reviewNo)
	if err != nil {
		return nil, fmt.Errorf("reviewRepository.Save: %w", err)
	}

	// Fetch full record with customer name
	err = r.db.QueryRow(ctx, `
		SELECT
			pr.review_no,
			pr.review_id,
			pr.product_no,
			p.product_id,
			pr.customer_no,
			c.first_name || ' ' || c.last_name AS customer_name,
			pr.rating,
			COALESCE(pr.title, '') AS title,
			pr.body,
			pr.created_at
		FROM catalog.product_reviews pr
		JOIN catalog.products p   ON p.product_no  = pr.product_no
		JOIN customer.customers c ON c.customer_no = pr.customer_no
		WHERE pr.review_no = $1
	`, reviewNo).Scan(
		&review.ReviewNo,
		&review.ReviewId,
		&review.ProductNo,
		&review.ProductId,
		&review.CustomerNo,
		&review.CustomerName,
		&review.Rating,
		&review.Title,
		&review.Body,
		&review.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("reviewRepository.Save: fetch after insert: %w", err)
	}
	return review, nil
}

func (r *reviewRepository) GetAll(ctx context.Context, req *models.ReviewListRequest) (*models.ReviewListResponse, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 {
		req.Limit = 10
	}
	offset := (req.Page - 1) * req.Limit

	where := []string{"pr.status != 'D'"}
	args := []interface{}{}
	idx := 1

	if req.Search != "" {
		where = append(where, fmt.Sprintf(
			"(p.product_name ILIKE $%d OR c.first_name ILIKE $%d OR c.last_name ILIKE $%d OR pr.title ILIKE $%d OR pr.body ILIKE $%d)",
			idx, idx, idx, idx, idx,
		))
		args = append(args, "%"+req.Search+"%")
		idx++
	}
	if req.Rating > 0 {
		where = append(where, fmt.Sprintf("pr.rating = $%d", idx))
		args = append(args, req.Rating)
		idx++
	}
	if req.Status != "" {
		where = append(where, fmt.Sprintf("pr.status = $%d", idx))
		args = append(args, req.Status)
		idx++
	}

	whereClause := "WHERE " + strings.Join(where, " AND ")

	var total int
	err := r.db.QueryRow(ctx, fmt.Sprintf(`
		SELECT COUNT(*)
		FROM catalog.product_reviews pr
		JOIN catalog.products p   ON p.product_no  = pr.product_no
		JOIN customer.customers c ON c.customer_no = pr.customer_no
		%s
	`, whereClause), args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("reviewRepository.GetAll count: %w", err)
	}

	limitArg := idx
	offsetArg := idx + 1
	args = append(args, req.Limit, offset)

	rows, err := r.db.Query(ctx, fmt.Sprintf(`
		SELECT
			pr.review_no,
			pr.review_id,
			pr.product_no,
			p.product_id,
			p.product_name,
			pr.customer_no,
			c.first_name || ' ' || c.last_name AS customer_name,
			pr.rating,
			COALESCE(pr.title, '') AS title,
			pr.body,
			pr.status,
			pr.created_at
		FROM catalog.product_reviews pr
		JOIN catalog.products p   ON p.product_no  = pr.product_no
		JOIN customer.customers c ON c.customer_no = pr.customer_no
		%s
		ORDER BY pr.created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, limitArg, offsetArg), args...)
	if err != nil {
		return nil, fmt.Errorf("reviewRepository.GetAll query: %w", err)
	}
	defer rows.Close()

	var items []models.AdminReview
	for rows.Next() {
		var rv models.AdminReview
		if err := rows.Scan(
			&rv.ReviewNo, &rv.ReviewId,
			&rv.ProductNo, &rv.ProductId, &rv.ProductName,
			&rv.CustomerNo, &rv.CustomerName,
			&rv.Rating, &rv.Title, &rv.Body, &rv.Status, &rv.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("reviewRepository.GetAll scan: %w", err)
		}
		items = append(items, rv)
	}
	if items == nil {
		items = []models.AdminReview{}
	}

	totalPages := int(math.Ceil(float64(total) / float64(req.Limit)))
	return &models.ReviewListResponse{
		Items:      items,
		Total:      total,
		Page:       req.Page,
		Limit:      req.Limit,
		TotalPages: totalPages,
	}, nil
}

func (r *reviewRepository) Delete(ctx context.Context, reviewId uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE catalog.product_reviews SET status = 'D', updated_at = NOW()
		WHERE review_id = $1 AND status != 'D'
	`, reviewId)
	if err != nil {
		return fmt.Errorf("reviewRepository.Delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("reviewRepository.Delete: review not found")
	}
	return nil
}
