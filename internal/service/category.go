package service

import (
	"context"
	"errors"

	"github.com/Jarlene/tiaozao/internal/model"
	"github.com/Jarlene/tiaozao/internal/repository"
	"github.com/jackc/pgx/v5"
)

type CategoryService struct {
	repo *repository.CategoryRepository
}

func NewCategoryService(repo *repository.CategoryRepository) *CategoryService {
	return &CategoryService{repo: repo}
}

func (s *CategoryService) Create(ctx context.Context, cat *model.Category) error {
	if cat.Name == "" {
		return errors.New("name is required")
	}
	return s.repo.Create(ctx, cat)
}

func (s *CategoryService) Update(ctx context.Context, id uint, updates map[string]any) error {
	return s.repo.Update(ctx, id, updates)
}

func (s *CategoryService) Delete(ctx context.Context, id uint) error {
	return s.repo.Delete(ctx, id)
}

func (s *CategoryService) GetByID(ctx context.Context, id uint) (*model.Category, error) {
	cat, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("category not found")
		}
		return nil, err
	}
	return cat, nil
}

func (s *CategoryService) List(ctx context.Context) ([]model.Category, error) {
	return s.repo.List(ctx)
}
