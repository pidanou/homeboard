package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/pidanou/homeboard/internal/model"
	"github.com/pidanou/homeboard/internal/repository"
)

type CategoryService struct {
	categories repository.CategoryRepository
}

func NewCategoryService(categories repository.CategoryRepository) *CategoryService {
	return &CategoryService{categories: categories}
}

func (s *CategoryService) Create(ctx context.Context, familyID, name, color string) (*model.Category, error) {
	category := &model.Category{
		ID:        uuid.NewString(),
		FamilyID:  familyID,
		Name:      name,
		Color:     color,
		CreatedAt: time.Now().UTC(),
	}
	if err := s.categories.Create(ctx, category); err != nil {
		return nil, err
	}
	return category, nil
}

func (s *CategoryService) ListForFamily(ctx context.Context, familyID string) ([]*model.Category, error) {
	return s.categories.ListByFamilyID(ctx, familyID)
}

func (s *CategoryService) Delete(ctx context.Context, categoryID, familyID string) error {
	return s.categories.Delete(ctx, categoryID, familyID)
}

func (s *CategoryService) Update(ctx context.Context, categoryID, familyID, name, color string) (*model.Category, error) {
	cat := &model.Category{ID: categoryID, FamilyID: familyID, Name: name, Color: color}
	if err := s.categories.Update(ctx, cat); err != nil {
		return nil, err
	}
	return cat, nil
}
