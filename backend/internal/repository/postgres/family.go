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

func (r *FamilyRepository) CreateVirtualMember(ctx context.Context, m *model.VirtualMember) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO virtual_members (id, family_id, name, created_at) VALUES ($1, $2, $3, $4)`,
		m.ID, m.FamilyID, m.Name, m.CreatedAt,
	)
	return err
}

func (r *FamilyRepository) DeleteVirtualMember(ctx context.Context, id, familyID string) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM virtual_members WHERE id = $1 AND family_id = $2`,
		id, familyID,
	)
	return err
}

func (r *FamilyRepository) GetUnlinkedVirtualMembers(ctx context.Context, familyID string) ([]*model.VirtualMember, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, family_id, name, linked_user_id, created_at FROM virtual_members
		 WHERE family_id = $1 AND linked_user_id IS NULL ORDER BY created_at`,
		familyID,
	)
	if err != nil {
		return nil, fmt.Errorf("get unlinked virtual members: %w", err)
	}
	defer rows.Close()
	var members []*model.VirtualMember
	for rows.Next() {
		m := &model.VirtualMember{}
		if err := rows.Scan(&m.ID, &m.FamilyID, &m.Name, &m.LinkedUserID, &m.CreatedAt); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, rows.Err()
}

func (r *FamilyRepository) LinkVirtualMember(ctx context.Context, virtualID, familyID, userID string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err = tx.Exec(ctx,
		`UPDATE virtual_members SET linked_user_id = $1 WHERE id = $2 AND family_id = $3`,
		userID, virtualID, familyID,
	); err != nil {
		return fmt.Errorf("link virtual member: %w", err)
	}
	if _, err = tx.Exec(ctx,
		`UPDATE tasks SET assigned_to = $1 WHERE assigned_to = $2 AND family_id = $3`,
		userID, virtualID, familyID,
	); err != nil {
		return fmt.Errorf("migrate task assignments: %w", err)
	}
	if _, err = tx.Exec(ctx,
		`UPDATE event_attendees SET user_id = $1
		 WHERE user_id = $2 AND event_id IN (SELECT id FROM events WHERE family_id = $3)`,
		userID, virtualID, familyID,
	); err != nil {
		return fmt.Errorf("migrate event attendees: %w", err)
	}
	return tx.Commit(ctx)
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
