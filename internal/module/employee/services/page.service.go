package services

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/google/uuid"

	"github.com/chand-magar/gamenglam-server/internal/module/employee/models"
	"github.com/chand-magar/gamenglam-server/internal/module/employee/repository"
)

var (
	ErrPageNotFound = errors.New("page not found")
)

type pageService struct {
	repo repository.PageRepository
}

func NewPageService(repo repository.PageRepository) PageService {
	return &pageService{repo: repo}
}

func (s *pageService) FindAll(ctx context.Context) ([]models.Page, error) {
	data, err := s.repo.GetAll(ctx)
	if err != nil {
		log.Printf("pageService.FindAll error: %v", err)
		return nil, fmt.Errorf("pageService.FindAll: %w", err)
	}

	return data, nil
}

func (s *pageService) FindOne(ctx context.Context, Id uuid.UUID) (*models.Page, error) {
	data, err := s.repo.GetById(ctx, Id)
	if err != nil {
		log.Printf("pageService.FindOne error: %v", err)
		return nil, fmt.Errorf("pageService.FindOne: %w", err)
	}
	if data == nil {
		return nil, ErrPageNotFound
	}

	return data, nil
}

func (s *pageService) Create(ctx context.Context, req *models.CreatePageRequest) (*models.Page, error) {
	page := &models.Page{
		PageId:    uuid.New(),
		SectionNo: req.SectionNo,
		PageName:  req.PageName,
		PagePath:  req.PagePath,
		PageOrder: req.PageOrder,
		Status:    models.StatusActive,
	}

	created, err := s.repo.Save(ctx, page)
	if err != nil {
		return nil, fmt.Errorf("pageService.Save: %w", err)
	}

	return created, nil
}

func (s *pageService) Update(ctx context.Context, Id uuid.UUID, req *models.UpdatePageRequest) (*models.Page, error) {
	existing, err := s.repo.GetById(ctx, Id)
	if err != nil {
		return nil, fmt.Errorf("pageService.Update: %w", err)
	}
	if existing == nil {
		return nil, ErrPageNotFound
	}

	updated, err := s.repo.Modify(ctx, Id, req)
	if err != nil {
		return nil, fmt.Errorf("pageService.Update: %w", err)
	}

	return updated, nil
}

func (s *pageService) Delete(ctx context.Context, Id uuid.UUID) error {
	existing, err := s.repo.GetById(ctx, Id)
	if err != nil {
		return fmt.Errorf("pageService.Delete: %w", err)
	}
	if existing == nil {
		return ErrPageNotFound
	}

	if err := s.repo.Remove(ctx, Id); err != nil {
		return fmt.Errorf("pageService.Delete: %w", err)
	}

	return nil
}
