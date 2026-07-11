package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pidanou/homeboard/internal/model"
)

type ListRepository struct {
	pool *pgxpool.Pool
}

func NewListRepository(pool *pgxpool.Pool) *ListRepository {
	return &ListRepository{pool: pool}
}

func (r *ListRepository) CreateList(ctx context.Context, list *model.List) error {
	return r.pool.QueryRow(ctx,
		`INSERT INTO lists (id, family_id, name, position, created_at)
		 VALUES ($1, $2, $3, (SELECT COALESCE(MAX(position) + 1, 0) FROM lists WHERE family_id = $2), $4)
		 RETURNING position`,
		list.ID, list.FamilyID, list.Name, list.CreatedAt,
	).Scan(&list.Position)
}

func (r *ListRepository) ListsByFamilyID(ctx context.Context, familyID string) ([]*model.List, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, family_id, name, position, created_at FROM lists WHERE family_id = $1 ORDER BY position, created_at`,
		familyID,
	)
	if err != nil {
		return nil, fmt.Errorf("list lists: %w", err)
	}
	defer rows.Close()
	lists := make([]*model.List, 0)
	for rows.Next() {
		l := &model.List{}
		if err := rows.Scan(&l.ID, &l.FamilyID, &l.Name, &l.Position, &l.CreatedAt); err != nil {
			return nil, err
		}
		lists = append(lists, l)
	}
	return lists, rows.Err()
}

func (r *ListRepository) DeleteList(ctx context.Context, listID, familyID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM lists WHERE id = $1 AND family_id = $2`, listID, familyID)
	return err
}

func (r *ListRepository) RenameList(ctx context.Context, listID, familyID, name string) error {
	_, err := r.pool.Exec(ctx, `UPDATE lists SET name = $1 WHERE id = $2 AND family_id = $3`, name, listID, familyID)
	return err
}

func (r *ListRepository) ReorderLists(ctx context.Context, familyID string, orderedIDs []string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	for i, id := range orderedIDs {
		if _, err := tx.Exec(ctx,
			`UPDATE lists SET position = $1 WHERE id = $2 AND family_id = $3`,
			i, id, familyID,
		); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *ListRepository) CreateItem(ctx context.Context, item *model.ListItem) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO list_items (id, list_id, name, checked, created_at) VALUES ($1, $2, $3, $4, $5)`,
		item.ID, item.ListID, item.Name, item.Checked, item.CreatedAt,
	)
	return err
}

func (r *ListRepository) ItemsByListID(ctx context.Context, listID string) ([]*model.ListItem, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, list_id, name, checked, created_at, checked_at, manual_order FROM list_items WHERE list_id = $1`,
		listID,
	)
	if err != nil {
		return nil, fmt.Errorf("list items: %w", err)
	}
	defer rows.Close()
	items := make([]*model.ListItem, 0)
	for rows.Next() {
		it := &model.ListItem{}
		if err := rows.Scan(&it.ID, &it.ListID, &it.Name, &it.Checked, &it.CreatedAt, &it.CheckedAt, &it.ManualOrder); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	return items, rows.Err()
}

func (r *ListRepository) UpdateItem(ctx context.Context, item *model.ListItem, familyID string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE list_items
		 SET name = $1, checked = $2, checked_at = CASE WHEN $2 THEN now() ELSE NULL END
		 WHERE id = $3 AND list_id IN (SELECT id FROM lists WHERE id = $4 AND family_id = $5)`,
		item.Name, item.Checked, item.ID, item.ListID, familyID,
	)
	return err
}

func (r *ListRepository) DeleteItem(ctx context.Context, itemID, listID, familyID string) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM list_items
		 WHERE id = $1 AND list_id IN (SELECT id FROM lists WHERE id = $2 AND family_id = $3)`,
		itemID, listID, familyID,
	)
	return err
}

func (r *ListRepository) DeleteCheckedItems(ctx context.Context, listID, familyID string) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM list_items
		 WHERE checked = true AND list_id IN (SELECT id FROM lists WHERE id = $1 AND family_id = $2)`,
		listID, familyID,
	)
	return err
}

func (r *ListRepository) ReorderItems(ctx context.Context, listID, familyID string, orderedIDs []string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	for i, id := range orderedIDs {
		if _, err := tx.Exec(ctx,
			`UPDATE list_items SET manual_order = $1
			 WHERE id = $2 AND list_id IN (SELECT id FROM lists WHERE id = $3 AND family_id = $4)`,
			i, id, listID, familyID,
		); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}
