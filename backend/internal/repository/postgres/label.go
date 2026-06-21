package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pidanou/homeboard/internal/model"
)

type CategoryRepository struct {
	pool *pgxpool.Pool
}

func NewCategoryRepository(pool *pgxpool.Pool) *CategoryRepository {
	return &CategoryRepository{pool: pool}
}

func (r *CategoryRepository) Create(ctx context.Context, category *model.Category) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO categories (id, family_id, name, color, created_at) VALUES ($1, $2, $3, $4, $5)`,
		category.ID, category.FamilyID, category.Name, category.Color, category.CreatedAt,
	)
	return err
}

func (r *CategoryRepository) ListByFamilyID(ctx context.Context, familyID string) ([]*model.Category, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, family_id, name, color, created_at FROM categories WHERE family_id = $1 ORDER BY created_at`,
		familyID,
	)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	defer rows.Close()

	categories := make([]*model.Category, 0)
	for rows.Next() {
		c := &model.Category{}
		if err := rows.Scan(&c.ID, &c.FamilyID, &c.Name, &c.Color, &c.CreatedAt); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}
	return categories, rows.Err()
}

func (r *CategoryRepository) Delete(ctx context.Context, categoryID, familyID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM categories WHERE id = $1 AND family_id = $2`, categoryID, familyID)
	return err
}

func (r *CategoryRepository) Update(ctx context.Context, category *model.Category) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE categories SET name = $1, color = $2 WHERE id = $3 AND family_id = $4`,
		category.Name, category.Color, category.ID, category.FamilyID,
	)
	return err
}
