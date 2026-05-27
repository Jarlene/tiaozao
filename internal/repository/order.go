package repository

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/Jarlene/tiaozao/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderRepository struct {
	pool *pgxpool.Pool
}

func NewOrderRepository(pool *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{pool: pool}
}

// generateOrderNo 生成订单号: 20060102150405 + 6位随机数
func generateOrderNo() string {
	return fmt.Sprintf("%s%06d", time.Now().Format("20060102150405"), rand.Intn(1000000))
}

// Create 创建订单（事务）
func (r *OrderRepository) Create(ctx context.Context, order *model.Order) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	order.OrderNo = generateOrderNo()
	order.Status = model.OrderStatusPending
	order.CreatedAt = time.Now()
	order.UpdatedAt = order.CreatedAt

	err = tx.QueryRow(ctx,
		`INSERT INTO orders (order_no, buyer_id, seller_id, total_amount, discount_amount, pay_amount, status, address_id, remark, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		 RETURNING id`,
		order.OrderNo, order.BuyerID, order.SellerID, order.TotalAmount, order.DiscountAmount,
		order.PayAmount, order.Status, order.AddressID, order.Remark, order.CreatedAt, order.UpdatedAt,
	).Scan(&order.ID)
	if err != nil {
		return err
	}

	for i := range order.Items {
		item := &order.Items[i]
		item.OrderID = order.ID
		err = tx.QueryRow(ctx,
			`INSERT INTO order_items (order_id, product_id, product_title, product_image, price, quantity, subtotal)
			 VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
			item.OrderID, item.ProductID, item.ProductTitle, item.ProductImage,
			item.Price, item.Quantity, item.Subtotal,
		).Scan(&item.ID)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

// GetByID 获取订单详情（含明细）
func (r *OrderRepository) GetByID(ctx context.Context, id uint) (*model.Order, error) {
	order := &model.Order{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, order_no, buyer_id, seller_id, total_amount, discount_amount, pay_amount,
		        status, address_id, COALESCE(remark, ''), paid_at, shipped_at, received_at, closed_at,
		        created_at, updated_at
		 FROM orders WHERE id = $1`, id,
	).Scan(&order.ID, &order.OrderNo, &order.BuyerID, &order.SellerID,
		&order.TotalAmount, &order.DiscountAmount, &order.PayAmount,
		&order.Status, &order.AddressID, &order.Remark,
		&order.PaidAt, &order.ShippedAt, &order.ReceivedAt, &order.ClosedAt,
		&order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("order not found")
		}
		return nil, err
	}

	items, err := r.getItems(ctx, order.ID)
	if err != nil {
		return nil, err
	}
	order.Items = items

	return order, nil
}

// UpdateStatus 更新订单状态（含时间戳）
func (r *OrderRepository) UpdateStatus(ctx context.Context, id uint, status string, timestamps map[string]time.Time) error {
	query := "UPDATE orders SET status = $1, updated_at = $2"
	args := []any{status, time.Now()}
	argIdx := 3

	for field, ts := range timestamps {
		query += fmt.Sprintf(", %s = $%d", field, argIdx)
		args = append(args, ts)
		argIdx++
	}

	query += fmt.Sprintf(" WHERE id = $%d", argIdx)
	args = append(args, id)

	_, err := r.pool.Exec(ctx, query, args...)
	return err
}

// ListByBuyer 买家订单列表
func (r *OrderRepository) ListByBuyer(ctx context.Context, buyerID uint, page, size int, status string) ([]model.Order, int64, error) {
	return r.list(ctx, "buyer_id", buyerID, page, size, status)
}

// ListBySeller 卖家订单列表
func (r *OrderRepository) ListBySeller(ctx context.Context, sellerID uint, page, size int, status string) ([]model.Order, int64, error) {
	return r.list(ctx, "seller_id", sellerID, page, size, status)
}

func (r *OrderRepository) list(ctx context.Context, field string, userID uint, page, size int, status string) ([]model.Order, int64, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}

	where := []string{fmt.Sprintf("%s = $1", field)}
	args := []any{userID}
	argIdx := 2

	if status != "" {
		where = append(where, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, status)
		argIdx++
	}

	whereClause := strings.Join(where, " AND ")

	var total int64
	countQuery := "SELECT COUNT(*) FROM orders WHERE " + whereClause
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * size
	dataQuery := fmt.Sprintf(
		`SELECT id, order_no, buyer_id, seller_id, total_amount, discount_amount, pay_amount,
		        status, address_id, COALESCE(remark, ''), paid_at, shipped_at, received_at, closed_at,
		        created_at, updated_at
		 FROM orders WHERE %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		whereClause, argIdx, argIdx+1)
	dataArgs := append(args, size, offset)

	rows, err := r.pool.Query(ctx, dataQuery, dataArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var orders []model.Order
	for rows.Next() {
		var o model.Order
		if err := rows.Scan(&o.ID, &o.OrderNo, &o.BuyerID, &o.SellerID,
			&o.TotalAmount, &o.DiscountAmount, &o.PayAmount,
			&o.Status, &o.AddressID, &o.Remark,
			&o.PaidAt, &o.ShippedAt, &o.ReceivedAt, &o.ClosedAt,
			&o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, 0, err
		}
		orders = append(orders, o)
	}

	// Load items for all orders
	for i := range orders {
		items, err := r.getItems(ctx, orders[i].ID)
		if err != nil {
			return nil, 0, err
		}
		orders[i].Items = items
	}

	return orders, total, nil
}

func (r *OrderRepository) getItems(ctx context.Context, orderID uint) ([]model.OrderItem, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, order_id, product_id, product_title, COALESCE(product_image, ''), price, quantity, subtotal
		 FROM order_items WHERE order_id = $1 ORDER BY id`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.OrderItem
	for rows.Next() {
		var item model.OrderItem
		if err := rows.Scan(&item.ID, &item.OrderID, &item.ProductID, &item.ProductTitle,
			&item.ProductImage, &item.Price, &item.Quantity, &item.Subtotal); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}
