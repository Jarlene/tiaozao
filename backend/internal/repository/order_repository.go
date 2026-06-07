package repository

import (
	"flea-market/internal/model"

	"gorm.io/gorm"
)

type OrderRepository interface {
	Create(order *model.Order) error
	FindByID(id uint) (*model.Order, error)
	FindByOrderNo(orderNo string) (*model.Order, error)
	Update(order *model.Order) error

	// 列表查询
	ListByBuyer(buyerID uint, status *model.OrderStatus, page, pageSize int) ([]model.Order, int64, error)
	ListBySeller(sellerID uint, status *model.OrderStatus, page, pageSize int) ([]model.Order, int64, error)

	// 管理员按状态查询所有订单
	ListByStatus(status model.OrderStatus, page, pageSize int) ([]model.Order, int64, error)

	// 库存操作
	DecrementProductStock(productID uint, quantity int) (bool, error) // true=扣减成功, false=库存不足
	IncrementProductStock(productID uint, quantity int) error

	// 状态日志
	CreateStatusLog(log *model.OrderStatusLog) error
	ListStatusLogs(orderID uint) ([]model.OrderStatusLog, error)

	// 事务
	Transaction(fc func(txRepo OrderRepository) error) error

	// GetDB 获取底层数据库连接（用于在同一事务中操作其他表）
	GetDB() *gorm.DB
}

type orderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) Create(order *model.Order) error {
	return r.db.Create(order).Error
}

func (r *orderRepository) FindByID(id uint) (*model.Order, error) {
	var order model.Order
	err := r.db.Preload("Product").
		Preload("Buyer").
		Preload("Seller").
		First(&order, id).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) FindByOrderNo(orderNo string) (*model.Order, error) {
	var order model.Order
	err := r.db.Where("order_no = ?", orderNo).
		Preload("Product").
		Preload("Buyer").
		Preload("Seller").
		First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) Update(order *model.Order) error {
	return r.db.Save(order).Error
}

func (r *orderRepository) DecrementProductStock(productID uint, quantity int) (bool, error) {
	result := r.db.Model(&model.Product{}).
		Where("id = ? AND stock >= ?", productID, quantity).
		UpdateColumn("stock", gorm.Expr("stock - ?", quantity))
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func (r *orderRepository) IncrementProductStock(productID uint, quantity int) error {
	return r.db.Model(&model.Product{}).
		Where("id = ? AND deleted_at IS NULL", productID).
		UpdateColumn("stock", gorm.Expr("stock + ?", quantity)).Error
}

func (r *orderRepository) ListByBuyer(buyerID uint, status *model.OrderStatus, page, pageSize int) ([]model.Order, int64, error) {
	var orders []model.Order
	var total int64

	query := r.db.Model(&model.Order{}).Where("buyer_id = ?", buyerID)
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Preload("Product", func(db *gorm.DB) *gorm.DB {
			return db.Preload("Images")
		}).
		Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&orders).Error
	if err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

func (r *orderRepository) ListBySeller(sellerID uint, status *model.OrderStatus, page, pageSize int) ([]model.Order, int64, error) {
	var orders []model.Order
	var total int64

	query := r.db.Model(&model.Order{}).Where("seller_id = ?", sellerID)
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Preload("Product", func(db *gorm.DB) *gorm.DB {
			return db.Preload("Images")
		}).
		Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&orders).Error
	if err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

// ListByStatus 管理员按订单状态查询所有订单（不分卖家/买家）
func (r *orderRepository) ListByStatus(status model.OrderStatus, page, pageSize int) ([]model.Order, int64, error) {
	var orders []model.Order
	var total int64

	query := r.db.Model(&model.Order{}).Where("status = ?", status)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Preload("Product", func(db *gorm.DB) *gorm.DB {
			return db.Preload("Images")
		}).
		Preload("Buyer").
		Preload("Seller").
		Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&orders).Error
	if err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

func (r *orderRepository) CreateStatusLog(log *model.OrderStatusLog) error {
	return r.db.Create(log).Error
}

func (r *orderRepository) ListStatusLogs(orderID uint) ([]model.OrderStatusLog, error) {
	var logs []model.OrderStatusLog
	err := r.db.Where("order_id = ?", orderID).
		Order("created_at ASC").
		Find(&logs).Error
	return logs, err
}

func (r *orderRepository) Transaction(fc func(txRepo OrderRepository) error) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		txRepo := &orderRepository{db: tx}
		return fc(txRepo)
	})
}

func (r *orderRepository) GetDB() *gorm.DB {
	return r.db
}
