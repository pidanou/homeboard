package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pidanou/homeboard/internal/model"
)

type CalendarExportTokenRepository struct {
	pool *pgxpool.Pool
}

func NewCalendarExportTokenRepository(pool *pgxpool.Pool) *CalendarExportTokenRepository {
	return &CalendarExportTokenRepository{pool: pool}
}

func (r *CalendarExportTokenRepository) Create(ctx context.Context, token *model.CalendarExportToken) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO calendar_export_tokens (token, family_id, created_by, created_at)
		 VALUES ($1, $2, $3, $4)`,
		token.Token, token.FamilyID, token.CreatedBy, token.CreatedAt,
	)
	return err
}

func (r *CalendarExportTokenRepository) GetByFamilyID(ctx context.Context, familyID string) (*model.CalendarExportToken, error) {
	t := &model.CalendarExportToken{}
	err := r.pool.QueryRow(ctx,
		`SELECT token, family_id, created_by, created_at FROM calendar_export_tokens WHERE family_id = $1`,
		familyID,
	).Scan(&t.Token, &t.FamilyID, &t.CreatedBy, &t.CreatedAt)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (r *CalendarExportTokenRepository) GetByToken(ctx context.Context, token string) (*model.CalendarExportToken, error) {
	t := &model.CalendarExportToken{}
	err := r.pool.QueryRow(ctx,
		`SELECT token, family_id, created_by, created_at FROM calendar_export_tokens WHERE token = $1`,
		token,
	).Scan(&t.Token, &t.FamilyID, &t.CreatedBy, &t.CreatedAt)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (r *CalendarExportTokenRepository) DeleteByFamilyID(ctx context.Context, familyID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM calendar_export_tokens WHERE family_id = $1`, familyID)
	return err
}
