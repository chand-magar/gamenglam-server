package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/chand-magar/gamenglam-server/internal/module/catalog/models"
	"github.com/chand-magar/gamenglam-server/internal/module/catalog/repository"
)

type variantImageRepository struct {
	db *pgxpool.Pool
}

func NewVariantImageRepository(db *pgxpool.Pool) repository.VariantImageRepository {
	return &variantImageRepository{db: db}
}

const imageColumns = `image_no, image_id, variant_no, image_url, sort_order, is_primary, created_at`

func scanImage(row pgx.Row) (*models.ProductVariantImage, error) {
	var img models.ProductVariantImage
	err := row.Scan(
		&img.ImageNo, &img.ImageId, &img.VariantNo,
		&img.ImageURL, &img.SortOrder, &img.IsPrimary, &img.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &img, nil
}

func (r *variantImageRepository) GetByVariantId(ctx context.Context, variantId uuid.UUID) ([]models.ProductVariantImage, error) {
	rows, err := r.db.Query(ctx, `
		SELECT `+imageColumns+`
		FROM catalog.product_variant_images
		WHERE variant_no = (
			SELECT variant_no FROM catalog.product_variants WHERE variant_id = $1
		)
		ORDER BY is_primary DESC, sort_order ASC, created_at ASC
	`, variantId)
	if err != nil {
		return nil, fmt.Errorf("variantImageRepository.GetByVariantId: %w", err)
	}
	defer rows.Close()

	var images []models.ProductVariantImage
	for rows.Next() {
		img, err := scanImage(rows)
		if err != nil {
			return nil, fmt.Errorf("variantImageRepository.GetByVariantId: scan: %w", err)
		}
		images = append(images, *img)
	}
	return images, rows.Err()
}

func (r *variantImageRepository) Save(ctx context.Context, img *models.ProductVariantImage) (*models.ProductVariantImage, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("variantImageRepository.Save: begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	// If primary — unset existing primary first
	if img.IsPrimary {
		_, err = tx.Exec(ctx, `
			UPDATE catalog.product_variant_images
			SET is_primary = FALSE
			WHERE variant_no = $1 AND is_primary = TRUE
		`, img.VariantNo)
		if err != nil {
			return nil, fmt.Errorf("variantImageRepository.Save: unset primary: %w", err)
		}
	}

	saved, err := scanImage(tx.QueryRow(ctx, `
		INSERT INTO catalog.product_variant_images
			(image_id, variant_no, image_url, sort_order, is_primary)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING `+imageColumns,
		uuid.New(), img.VariantNo, img.ImageURL, img.SortOrder, img.IsPrimary,
	))
	if err != nil {
		return nil, fmt.Errorf("variantImageRepository.Save: insert: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("variantImageRepository.Save: commit: %w", err)
	}
	return saved, nil
}

func (r *variantImageRepository) BulkSave(ctx context.Context, variantNo uint32, urls []string) ([]models.ProductVariantImage, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("variantImageRepository.BulkSave: begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	var images []models.ProductVariantImage
	for i, url := range urls {
		isPrimary := i == 0
		img, err := scanImage(tx.QueryRow(ctx, `
			INSERT INTO catalog.product_variant_images
				(image_id, variant_no, image_url, sort_order, is_primary)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING `+imageColumns,
			uuid.New(), variantNo, url, i, isPrimary,
		))
		if err != nil {
			return nil, fmt.Errorf("variantImageRepository.BulkSave: insert: %w", err)
		}
		images = append(images, *img)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("variantImageRepository.BulkSave: commit: %w", err)
	}
	return images, nil
}

func (r *variantImageRepository) Remove(ctx context.Context, imageId uuid.UUID) error {
	_, err := r.db.Exec(ctx, `
		DELETE FROM catalog.product_variant_images WHERE image_id = $1
	`, imageId)
	if err != nil {
		return fmt.Errorf("variantImageRepository.Remove: %w", err)
	}
	return nil
}

func (r *variantImageRepository) RemoveAllByVariantId(ctx context.Context, variantId uuid.UUID) error {
	_, err := r.db.Exec(ctx, `
		DELETE FROM catalog.product_variant_images
		WHERE variant_no = (
			SELECT variant_no FROM catalog.product_variants WHERE variant_id = $1
		)
	`, variantId)
	if err != nil {
		return fmt.Errorf("variantImageRepository.RemoveAllByVariantId: %w", err)
	}
	return nil
}

func (r *variantImageRepository) GetById(ctx context.Context, imageId uuid.UUID) (*models.ProductVariantImage, error) {
	img, err := scanImage(r.db.QueryRow(ctx, `
		SELECT `+imageColumns+`
		FROM catalog.product_variant_images
		WHERE image_id = $1
	`, imageId))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("variantImageRepository.GetById: %w", err)
	}
	return img, nil
}

func (r *variantImageRepository) GetVariantNoById(ctx context.Context, variantId uuid.UUID) (uint32, error) {
	var variantNo uint32
	err := r.db.QueryRow(ctx, `
		SELECT variant_no FROM catalog.product_variants WHERE variant_id = $1
	`, variantId).Scan(&variantNo)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, nil
		}
		return 0, fmt.Errorf("variantImageRepository.GetVariantNoById: %w", err)
	}
	return variantNo, nil
}

func (r *variantImageRepository) GetPrimaryByVariantIds(ctx context.Context, variantIds []uuid.UUID) (map[string]string, error) {
	if len(variantIds) == 0 {
		return map[string]string{}, nil
	}

	placeholders := make([]string, len(variantIds))
	args := make([]interface{}, len(variantIds))
	for i, id := range variantIds {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	query := fmt.Sprintf(`
		SELECT DISTINCT ON (pv.variant_id)
			pv.variant_id::text,
			pvi.image_url
		FROM catalog.product_variant_images pvi
		JOIN catalog.product_variants pv ON pv.variant_no = pvi.variant_no
		WHERE pv.variant_id IN (%s)
		ORDER BY pv.variant_id, pvi.is_primary DESC, pvi.sort_order ASC
	`, strings.Join(placeholders, ","))

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("variantImageRepository.GetPrimaryByVariantIds: %w", err)
	}
	defer rows.Close()

	result := make(map[string]string)
	for rows.Next() {
		var variantId string
		var imageURL string
		if err := rows.Scan(&variantId, &imageURL); err != nil {
			return nil, fmt.Errorf("variantImageRepository.GetPrimaryByVariantIds: scan: %w", err)
		}
		result[variantId] = imageURL
	}
	return result, rows.Err()
}
