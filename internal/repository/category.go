package repository

import (
	"context"
	"time"

	"github.com/Jarlene/tiaozao/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CategoryRepository struct {
	pool *pgxpool.Pool
}

func NewCategoryRepository(pool *pgxpool.Pool) *CategoryRepository {
	return &CategoryRepository{pool: pool}
}

func (r *CategoryRepository) Create(ctx context.Context, cat *model.Category) error {
	return r.pool.QueryRow(ctx,
		`INSERT INTO categories (name, parent_id, sort_order, created_at, updated_at)
		 VALUES ($1, $2, $3, NOW(), NOW())
		 RETURNING id, created_at, updated_at`,
		cat.Name, cat.ParentID, cat.SortOrder,
	).Scan(&cat.ID, &cat.CreatedAt, &cat.UpdatedAt)
}

func (r *CategoryRepository) Update(ctx context.Context, id uint, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}
	updates["updated_at"] = time.Now()
	return executeUpdate(ctx, r.pool, "categories", id, updates)
}

func (r *CategoryRepository) Delete(ctx context.Context, id uint) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Set children's parent_id to NULL
	if _, err := tx.Exec(ctx, `UPDATE categories SET parent_id = NULL, updated_at = NOW() WHERE parent_id = $1`, id); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM categories WHERE id = $1`, id); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *CategoryRepository) GetByID(ctx context.Context, id uint) (*model.Category, error) {
	cat := &model.Category{}
	row := r.pool.QueryRow(ctx,
		`SELECT id, name, parent_id, sort_order, created_at, updated_at FROM categories WHERE id = $1`, id)
	err := row.Scan(&cat.ID, &cat.Name, &cat.ParentID, &cat.SortOrder, &cat.CreatedAt, &cat.UpdatedAt)
	if err != nil {
		return nil, err
	}
	// Load children
	children, err := r.getChildren(ctx, id)
	if err != nil {
		return nil, err
	}
	cat.Children = children
	return cat, nil
}

func (r *CategoryRepository) List(ctx context.Context) ([]model.Category, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, name, parent_id, sort_order, created_at, updated_at
		 FROM categories WHERE parent_id IS NULL
		 ORDER BY sort_order ASC, id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cats []model.Category
	for rows.Next() {
		var c model.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.ParentID, &c.SortOrder, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		children, err := r.getChildren(ctx, c.ID)
		if err != nil {
			return nil, err
		}
		c.Children = children
		cats = append(cats, c)
	}
	return cats, nil
}

func (r *CategoryRepository) getChildren(ctx context.Context, parentID uint) ([]model.Category, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, name, parent_id, sort_order, created_at, updated_at
		 FROM categories WHERE parent_id = $1
		 ORDER BY sort_order ASC, id ASC`, parentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var children []model.Category
	for rows.Next() {
		var c model.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.ParentID, &c.SortOrder, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		children = append(children, c)
	}
	return children, nil
}

func executeUpdate(ctx context.Context, pool *pgxpool.Pool, table string, id uint, updates map[string]any) error {
	query := "UPDATE " + table + " SET "
	args := []any{}
	i := 1
	for k, v := range updates {
		if i > 1 {
			query += ", "
		}
		query += k + " = $" + itoa(i)
		args = append(args, v)
		i++
	}
	query += " WHERE id = $" + itoa(i)
	args = append(args, id)
	_, err := pool.Exec(ctx, query, args...)
	return err
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [12]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
