package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pidanou/family-board/internal/model"
)

type FamilyRepository struct {
	pool *pgxpool.Pool
}

func NewFamilyRepository(pool *pgxpool.Pool) *FamilyRepository {
	return &FamilyRepository{pool: pool}
}

func (r *FamilyRepository) Create(ctx context.Context, family *model.Family) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO families (id, name, created_at) VALUES ($1, $2, $3)`,
		family.ID, family.Name, family.CreatedAt,
	)
	return err
}

func (r *FamilyRepository) GetByID(ctx context.Context, id string) (*model.Family, error) {
	family := &model.Family{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, created_at FROM families WHERE id = $1`,
		id,
	).Scan(&family.ID, &family.Name, &family.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("get family by id: %w", err)
	}
	return family, nil
}

func (r *FamilyRepository) AddMember(ctx context.Context, member *model.FamilyMember) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO family_members (family_id, user_id, role, joined_at)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (family_id, user_id) DO NOTHING`,
		member.FamilyID, member.UserID, member.Role, member.JoinedAt,
	)
	return err
}

func (r *FamilyRepository) GetMembers(ctx context.Context, familyID string) ([]*model.FamilyMember, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT fm.family_id, fm.user_id, u.name, u.email, fm.role, fm.joined_at
		 FROM family_members fm JOIN users u ON u.id = fm.user_id
		 WHERE fm.family_id = $1`,
		familyID,
	)
	if err != nil {
		return nil, fmt.Errorf("get family members: %w", err)
	}
	defer rows.Close()

	var members []*model.FamilyMember
	for rows.Next() {
		m := &model.FamilyMember{}
		if err := rows.Scan(&m.FamilyID, &m.UserID, &m.Name, &m.Email, &m.Role, &m.JoinedAt); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, rows.Err()
}

func (r *FamilyRepository) GetFamiliesByUserID(ctx context.Context, userID string) ([]*model.Family, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT f.id, f.name, f.created_at
		 FROM families f
		 JOIN family_members fm ON fm.family_id = f.id
		 WHERE fm.user_id = $1`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("get families by user: %w", err)
	}
	defer rows.Close()

	families := make([]*model.Family, 0)
	for rows.Next() {
		f := &model.Family{}
		if err := rows.Scan(&f.ID, &f.Name, &f.CreatedAt); err != nil {
			return nil, err
		}
		families = append(families, f)
	}
	return families, rows.Err()
}
