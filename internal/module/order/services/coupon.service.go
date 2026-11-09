package services

import (
	"context"

	"github.com/google/uuid"

	"github.com/chand-magar/gamenglam-server/internal/module/order/models"
	repository "github.com/chand-magar/gamenglam-server/internal/module/order/repository"
)

type couponService struct {
	repo repository.CouponRepository
}

func NewCouponService(repo repository.CouponRepository) CouponService {
	return &couponService{repo: repo}
}

func (s *couponService) FindAll(ctx context.Context, req *models.CouponListRequest) (*models.CouponListResponse, error) {
	return s.repo.GetAll(ctx, req)
}

func (s *couponService) FindById(ctx context.Context, couponId uuid.UUID) (*models.Coupon, error) {
	coupon, err := s.repo.GetById(ctx, couponId)
	if err != nil {
		return nil, err
	}
	if coupon == nil {
		return nil, nil
	}
	return coupon, nil
}

func (s *couponService) Create(ctx context.Context, req *models.CreateCouponRequest) (*models.Coupon, error) {
	return s.repo.Save(ctx, req)
}

func (s *couponService) Update(ctx context.Context, couponId uuid.UUID, req *models.UpdateCouponRequest) (*models.Coupon, error) {
	return s.repo.Update(ctx, couponId, req)
}

func (s *couponService) Delete(ctx context.Context, couponId uuid.UUID) error {
	return s.repo.Delete(ctx, couponId)
}
