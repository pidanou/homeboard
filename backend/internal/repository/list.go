package repository

import (
	"context"

	"github.com/pidanou/homeboard/internal/model"
)

type ListRepository interface {
	CreateList(ctx context.Context, list *model.List) error
	ListsByFamilyID(ctx context.Context, familyID string) ([]*model.List, error)
	DeleteList(ctx context.Context, listID, familyID string) error
	RenameList(ctx context.Context, listID, familyID, name string) error
	CreateItem(ctx context.Context, item *model.ListItem) error
	ItemsByListID(ctx context.Context, listID string) ([]*model.ListItem, error)
	UpdateItem(ctx context.Context, item *model.ListItem, familyID string) error
	DeleteItem(ctx context.Context, itemID, listID, familyID string) error
	DeleteCheckedItems(ctx context.Context, listID, familyID string) error
}
