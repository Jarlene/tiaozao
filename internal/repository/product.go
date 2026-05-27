package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Jarlene/tiaozao/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProductRepository struct {
	pool *pgxpool.Pool
}

func NewProductRepository(pool *pgxpool.Pool) *ProductRepository {
	return &ProductRepository{pool: pool}
}

func (r *ProductRepository) Create(ctx context.Context, product *model.Product) error {
	return r.pool.QueryRow(ctx,
		`INSERT INTO products (title, description, price, original_price, status, category_id, seller_id, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
		 RETURNING id, created_at, updated_at`,
		product.Title, product.Description, product.Price, product.OriginalPrice,
		product.Status, product.CategoryID, product.SellerID,
	).Scan(&product.ID, &product.CreatedAt, &product.UpdatedAt)
}

func (r *ProductRepository) Update(ctx context.Context, id uint, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}
	updates["updated_at"] = time.Now()
	return executeUpdate(ctx, r.pool, "products", id, updates)
}

func (r *ProductRepository) GetByID(ctx context.Context, id uint) (*model.Product, error) {
	p := &model.Product{}
	row := r.pool.QueryRow(ctx,
		`SELECT id, title, description, price, original_price, status, category_id, seller_id, created_at, updated_at
		 FROM products WHERE id = $1`, id)
	err := row.Scan(&p.ID, &p.Title, &p.Description, &p.Price, &p.OriginalPrice,
		&p.Status, &p.CategoryID, &p.SellerID, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}

	// Load category
	if p.CategoryID != nil {
		p.Category = &model.Category{}
		err := r.pool.QueryRow(ctx,
			`SELECT id, name FROM categories WHERE id = $1`, *p.CategoryID).Scan(&p.Category.ID, &p.Category.Name)
		if err != nil {
			p.Category = nil // silently ignore if category was deleted
		}
	}

	// Load images
	images, err := r.getImages(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	p.Images = images

	return p, nil
}

type ProductListQuery struct {
	Page       int
	Size       int
	CategoryID *uint
	Status     string
	SellerID   *uint
}

func (r *ProductRepository) List(ctx context.Context, q ProductListQuery) ([]model.Product, int64, error) {
	where := []string{"1=1"}
	args := []any{}
	argIdx := 1

	if q.CategoryID != nil {
		where = append(where, fmt.Sprintf("category_id = $%d", argIdx))
		args = append(args, *q.CategoryID)
		argIdx++
	}
	if q.Status != "" {
		where = append(where, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, q.Status)
		argIdx++
	} else {
		where = append(where, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, string(model.ProductStatusActive))
		argIdx++
	}
	if q.SellerID != nil {
		where = append(where, fmt.Sprintf("seller_id = $%d", argIdx))
		args = append(args, *q.SellerID)
		argIdx++
	}

	whereClause := strings.Join(where, " AND ")

	// Count
	var total int64
	countQuery := "SELECT COUNT(*) FROM products WHERE " + whereClause
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Fetch
	offset := (q.Page - 1) * q.Size
	dataQuery := fmt.Sprintf(
		`SELECT id, title, description, price, original_price, status, category_id, seller_id, created_at, updated_at
		 FROM products WHERE %s ORDER BY updated_at DESC LIMIT $%d OFFSET $%d`,
		whereClause, argIdx, argIdx+1)
	dataArgs := append(args, q.Size, offset)

	rows, err := r.pool.Query(ctx, dataQuery, dataArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var products []model.Product
	for rows.Next() {
		var p model.Product
		if err := rows.Scan(&p.ID, &p.Title, &p.Description, &p.Price, &p.OriginalPrice,
			&p.Status, &p.CategoryID, &p.SellerID, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, 0, err
		}
		products = append(products, p)
	}

	return products, total, nil
}

func (r *ProductRepository) AssociateImages(ctx context.Context, productID uint, imageIDs []uint, coverImageID *uint) error {
	for i, imgID := range imageIDs {
		isCover := false
		if coverImageID != nil && imgID == *coverImageID {
			isCover = true
		}
		_, err := r.pool.Exec(ctx,
			`UPDATE product_images SET product_id = $1, sort_order = $2, is_cover = $3 WHERE id = $4`,
			productID, i, isCover, imgID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *ProductRepository) CreateImage(ctx context.Context, url string) (uint, error) {
	var id uint
	err := r.pool.QueryRow(ctx,
		`INSERT INTO product_images (url, sort_order, is_cover) VALUES ($1, 0, false) RETURNING id`,
		url).Scan(&id)
	return id, err
}

func (r *ProductRepository) GetProductSellerID(ctx context.Context, productID uint) (uint, error) {
	var sellerID uint
	err := r.pool.QueryRow(ctx, `SELECT seller_id FROM products WHERE id = $1`, productID).Scan(&sellerID)
	return sellerID, err
}

func (r *ProductRepository) getImages(ctx context.Context, productID uint) ([]model.ProductImage, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, product_id, url, sort_order, is_cover
		 FROM product_images WHERE product_id = $1
		 ORDER BY sort_order ASC`, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var images []model.ProductImage
	for rows.Next() {
		var img model.ProductImage
		if err := rows.Scan(&img.ID, &img.ProductID, &img.URL, &img.SortOrder, &img.IsCover); err != nil {
			return nil, err
		}
		images = append(images, img)
	}
	return images, nil
}
