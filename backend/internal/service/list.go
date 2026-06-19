package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/pidanou/family-board/internal/model"
	"github.com/pidanou/family-board/internal/repository"
)

type ListService struct {
	lists repository.ListRepository
}

func NewListService(lists repository.ListRepository) *ListService {
	return &ListService{lists: lists}
}

func (s *ListService) Create(ctx context.Context, familyID, name string) (*model.List, error) {
	l := &model.List{ID: uuid.NewString(), FamilyID: familyID, Name: name, CreatedAt: time.Now().UTC()}
	if err := s.lists.CreateList(ctx, l); err != nil {
		return nil, err
	}
	return l, nil
}

func (s *ListService) ListsByFamily(ctx context.Context, familyID string) ([]*model.List, error) {
	return s.lists.ListsByFamilyID(ctx, familyID)
}

func (s *ListService) Delete(ctx context.Context, listID, familyID string) error {
	return s.lists.DeleteList(ctx, listID, familyID)
}

func (s *ListService) Rename(ctx context.Context, listID, familyID, name string) error {
	return s.lists.RenameList(ctx, listID, familyID, name)
}

func (s *ListService) AddItem(ctx context.Context, listID, familyID, name string) (*model.ListItem, error) {
	item := &model.ListItem{ID: uuid.NewString(), ListID: listID, Name: name, Checked: false, CreatedAt: time.Now().UTC()}
	if err := s.lists.CreateItem(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *ListService) ItemsByList(ctx context.Context, listID string) ([]*model.ListItem, error) {
	return s.lists.ItemsByListID(ctx, listID)
}

func (s *ListService) UpdateItem(ctx context.Context, itemID, listID, familyID, name string, checked bool) error {
	return s.lists.UpdateItem(ctx, &model.ListItem{ID: itemID, ListID: listID, Name: name, Checked: checked}, familyID)
}

func (s *ListService) DeleteItem(ctx context.Context, itemID, listID, familyID string) error {
	return s.lists.DeleteItem(ctx, itemID, listID, familyID)
}

func (s *ListService) ClearChecked(ctx context.Context, listID, familyID string) error {
	return s.lists.DeleteCheckedItems(ctx, listID, familyID)
}
