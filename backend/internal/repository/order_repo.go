package repository

import (
	"github.com/Jarlene/tiaozao/backend/internal/model"
	"gorm.io/gorm"
)

// OrderRepository 订单数据仓库
type OrderRepository struct {
	db *gorm.DB
}

// NewOrderRepository 创建订单仓库实例
func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

// Create 创建订单
func (r *OrderRepository) Create(order *model.Order) error {
	return r.db.Create(order).Error
}

// FindByID 根据ID查找订单，预加载关联数据
func (r *OrderRepository) FindByID(id uint) (*model.Order, error) {
	var order model.Order
	err := r.db.Preload("Items.Product.Images").
		Preload("Buyer").
		Preload("Seller").
		First(&order, id).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// FindByOrderNo 根据订单号查找订单
func (r *OrderRepository) FindByOrderNo(orderNo string) (*model.Order, error) {
	var order model.Order
	err := r.db.Where("order_no = ?", orderNo).
		Preload("Items.Product.Images").
		First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// Update 更新订单
func (r *OrderRepository) Update(order *model.Order) error {
	return r.db.Save(order).Error
}

// ListByBuyer 获取买家订单列表
func (r *OrderRepository) ListByBuyer(buyerID uint, page, pageSize int, status string) ([]model.Order, int64, error) {
	var orders []model.Order
	var total int64

	query := r.db.Model(&model.Order{}).Where("buyer_id = ?", buyerID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	query.Count(&total)

	offset := (page - 1) * pageSize
	err := r.db.Where("buyer_id = ?", buyerID).
		Preload("Items.Product.Images").
		Preload("Seller").
		Offset(offset).Limit(pageSize).
		Order("created_at DESC").
		Find(&orders).Error
	return orders, total, err
}

// ListBySeller 获取卖家订单列表
func (r *OrderRepository) ListBySeller(sellerID uint, page, pageSize int, status string) ([]model.Order, int64, error) {
	var orders []model.Order
	var total int64

	query := r.db.Model(&model.Order{}).Where("seller_id = ?", sellerID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	query.Count(&total)

	offset := (page - 1) * pageSize
	err := r.db.Where("seller_id = ?", sellerID).
		Preload("Items.Product.Images").
		Preload("Buyer").
		Offset(offset).Limit(pageSize).
		Order("created_at DESC").
		Find(&orders).Error
	return orders, total, err
}

// DB 暴露原始DB用于事务
func (r *OrderRepository) DB() *gorm.DB {
	return r.db
}
