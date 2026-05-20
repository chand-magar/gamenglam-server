package services

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/google/uuid"

	"github.com/chand-magar/gamenglam-server/internal/module/auth/models"
	"github.com/chand-magar/gamenglam-server/internal/module/auth/repository"
)

var (
	ErrSectionNotFound = errors.New("section not found")
)

type sectionService struct {
	repo repository.SectionRepository
}

func NewSectionService(repo repository.SectionRepository) SectionService {
	return &sectionService{repo: repo}
}

func (s *sectionService) FindAll(ctx context.Context) ([]models.Section, error) {
	data, err := s.repo.GetAll(ctx)
	if err != nil {
		log.Printf("sectionService.FindAll error: %v", err)
		return nil, fmt.Errorf("sectionService.FindAll: %w", err)
	}

	return data, nil
}

func (s *sectionService) FindOne(ctx context.Context, Id uuid.UUID) (*models.Section, error) {
	data, err := s.repo.GetById(ctx, Id)
	if err != nil {
		log.Printf("sectionService.FindOne error: %v", err)
		return nil, fmt.Errorf("sectionService.FindOne: %w", err)
	}
	if data == nil {
		return nil, ErrSectionNotFound
	}

	return data, nil
}

func (s *sectionService) Create(ctx context.Context, req *models.CreateSectionRequest) (*models.Section, error) {
	section := &models.Section{
		SectionId:      uuid.New(),
		SectionName:    req.SectionName,
		SectionPath: 	req.SectionPath,
		SectionIcon: 	req.SectionIcon,
		SectionOrder: 	req.SectionOrder,
		Status:      	models.StatusActive,
	}

	created, err := s.repo.Save(ctx, section)
	if err != nil {
		return nil, fmt.Errorf("sectionService.Save: %w", err)
	}

	return created, nil
}

func (s *sectionService) Update(ctx context.Context, Id uuid.UUID, req *models.UpdateSectionRequest) (*models.Section, error) {
	existing, err := s.repo.GetById(ctx, Id)
	if err != nil {
		return nil, fmt.Errorf("sectionService.Update: %w", err)
	}
	if existing == nil {
		return nil, ErrSectionNotFound
	}

	updated, err := s.repo.Modify(ctx, Id, req)
	if err != nil {
		return nil, fmt.Errorf("sectionService.Update: %w", err)
	}

	return updated, nil
}

func (s *sectionService) Delete(ctx context.Context, Id uuid.UUID) error {
	existing, err := s.repo.GetById(ctx, Id)
	if err != nil {
		return fmt.Errorf("sectionService.Delete: %w", err)
	}
	if existing == nil {
		return ErrSectionNotFound
	}

	if err := s.repo.Remove(ctx, Id); err != nil {
		return fmt.Errorf("sectionService.Delete: %w", err)
	}

	return nil
}