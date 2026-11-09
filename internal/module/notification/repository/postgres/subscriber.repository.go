package postgres

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/chand-magar/gamenglam-server/internal/module/notification/models"
	repository "github.com/chand-magar/gamenglam-server/internal/module/notification/repository"
)

type subscriberRepository struct {
	db *pgxpool.Pool
}

func NewSubscriberRepository(db *pgxpool.Pool) repository.SubscriberRepository {
	return &subscriberRepository{db: db}
}

func (r *subscriberRepository) GetAll(ctx context.Context, req *models.SubscriberListRequest) (*models.SubscriberListResponse, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 {
		req.Limit = 10
	}
	offset := (req.Page - 1) * req.Limit

	where := []string{"status != 'D'"}
	args := []interface{}{}
	idx := 1

	if req.Search != "" {
		where = append(where, fmt.Sprintf("(email ILIKE $%d OR name ILIKE $%d)", idx, idx))
		args = append(args, "%"+req.Search+"%")
		idx++
	}
	if req.Status != "" {
		where = append(where, fmt.Sprintf("status = $%d", idx))
		args = append(args, req.Status)
		idx++
	}

	whereClause := "WHERE " + strings.Join(where, " AND ")

	var total int
	if err := r.db.QueryRow(ctx, fmt.Sprintf(
		`SELECT COUNT(*) FROM notification.newsletter_subscribers %s`, whereClause,
	), args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("subscriberRepository.GetAll count: %w", err)
	}

	limitArg, offsetArg := idx, idx+1
	args = append(args, req.Limit, offset)

	rows, err := r.db.Query(ctx, fmt.Sprintf(`
		SELECT subscriber_no, subscriber_id, email, COALESCE(name,''), status, subscribed_at, created_at
		FROM notification.newsletter_subscribers
		%s
		ORDER BY subscribed_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, limitArg, offsetArg), args...)
	if err != nil {
		return nil, fmt.Errorf("subscriberRepository.GetAll query: %w", err)
	}
	defer rows.Close()

	var items []models.Subscriber
	for rows.Next() {
		var s models.Subscriber
		if err := rows.Scan(&s.SubscriberNo, &s.SubscriberId, &s.Email, &s.Name, &s.Status, &s.SubscribedAt, &s.CreatedAt); err != nil {
			return nil, fmt.Errorf("subscriberRepository.GetAll scan: %w", err)
		}
		items = append(items, s)
	}
	if items == nil {
		items = []models.Subscriber{}
	}

	return &models.SubscriberListResponse{
		Items:      items,
		Total:      total,
		Page:       req.Page,
		Limit:      req.Limit,
		TotalPages: int(math.Ceil(float64(total) / float64(req.Limit))),
	}, nil
}

func (r *subscriberRepository) GetActiveEmails(ctx context.Context) ([]string, error) {
	rows, err := r.db.Query(ctx, `
		SELECT email FROM notification.newsletter_subscribers WHERE status = 'A' ORDER BY subscribed_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("subscriberRepository.GetActiveEmails: %w", err)
	}
	defer rows.Close()

	var emails []string
	for rows.Next() {
		var e string
		if err := rows.Scan(&e); err != nil {
			return nil, fmt.Errorf("subscriberRepository.GetActiveEmails scan: %w", err)
		}
		emails = append(emails, e)
	}
	return emails, nil
}

func (r *subscriberRepository) Save(ctx context.Context, req *models.CreateSubscriberRequest) (*models.Subscriber, error) {
	var s models.Subscriber
	err := r.db.QueryRow(ctx, `
		INSERT INTO notification.newsletter_subscribers (email, name)
		VALUES ($1, NULLIF($2,''))
		ON CONFLICT (email) DO UPDATE SET status = 'A', updated_at = NOW()
		RETURNING subscriber_no, subscriber_id, email, COALESCE(name,''), status, subscribed_at, created_at
	`, req.Email, req.Name).Scan(
		&s.SubscriberNo, &s.SubscriberId, &s.Email, &s.Name, &s.Status, &s.SubscribedAt, &s.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("subscriberRepository.Save: %w", err)
	}
	return &s, nil
}

func (r *subscriberRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE notification.newsletter_subscribers SET status = 'D', updated_at = NOW()
		WHERE subscriber_id = $1 AND status != 'D'
	`, id)
	if err != nil {
		return fmt.Errorf("subscriberRepository.Delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("subscriberRepository.Delete: not found")
	}
	return nil
}

func (r *subscriberRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM notification.newsletter_subscribers WHERE email = $1 AND status != 'D')
	`, email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("subscriberRepository.ExistsByEmail: %w", err)
	}
	return exists, nil
}
