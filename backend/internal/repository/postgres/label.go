package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pidanou/family-board/internal/model"
)

type LabelRepository struct {
	pool *pgxpool.Pool
}

func NewLabelRepository(pool *pgxpool.Pool) *LabelRepository {
	return &LabelRepository{pool: pool}
}

func (r *LabelRepository) Create(ctx context.Context, label *model.Label) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO labels (id, family_id, name, color, created_at) VALUES ($1, $2, $3, $4, $5)`,
		label.ID, label.FamilyID, label.Name, label.Color, label.CreatedAt,
	)
	return err
}

func (r *LabelRepository) ListByFamilyID(ctx context.Context, familyID string) ([]*model.Label, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, family_id, name, color, created_at FROM labels WHERE family_id = $1 ORDER BY created_at`,
		familyID,
	)
	if err != nil {
		return nil, fmt.Errorf("list labels: %w", err)
	}
	defer rows.Close()

	labels := make([]*model.Label, 0)
	for rows.Next() {
		l := &model.Label{}
		if err := rows.Scan(&l.ID, &l.FamilyID, &l.Name, &l.Color, &l.CreatedAt); err != nil {
			return nil, err
		}
		labels = append(labels, l)
	}
	return labels, rows.Err()
}

func (r *LabelRepository) Delete(ctx context.Context, labelID, familyID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM labels WHERE id = $1 AND family_id = $2`, labelID, familyID)
	return err
}
