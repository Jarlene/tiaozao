package repository

import (
	"flea-market/internal/model"

	"gorm.io/gorm"
)

type CategoryRepository interface {
	Create(category *model.Category) error
	Update(category *model.Category) error
	Delete(id uint) error
	FindByID(id uint) (*model.Category, error)
	ListAll() ([]model.Category, error)
	ListByParentID(parentID *uint) ([]model.Category, error)
	CountSubcategories(id uint) (int64, error)
	CountProducts(id uint) (int64, error)
}

type categoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) Create(category *model.Category) error {
	return r.db.Create(category).Error
}

func (r *categoryRepository) Update(category *model.Category) error {
	return r.db.Model(&model.Category{}).Where("id = ?", category.ID).Updates(map[string]interface{}{
		"name": category.Name,
	}).Error
}

func (r *categoryRepository) Delete(id uint) error {
	return r.db.Delete(&model.Category{}, id).Error
}

func (r *categoryRepository) FindByID(id uint) (*model.Category, error) {
	var category model.Category
	err := r.db.First(&category, id).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *categoryRepository) ListAll() ([]model.Category, error) {
	var categories []model.Category
	err := r.db.Order("id ASC").Find(&categories).Error
	return categories, err
}

func (r *categoryRepository) ListByParentID(parentID *uint) ([]model.Category, error) {
	var categories []model.Category
	err := r.db.Where("parent_id = ?", parentID).Order("id ASC").Find(&categories).Error
	return categories, err
}

func (r *categoryRepository) CountSubcategories(id uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.Category{}).Where("parent_id = ?", id).Count(&count).Error
	return count, err
}

func (r *categoryRepository) CountProducts(id uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.Product{}).Where("category_id = ?", id).Count(&count).Error
	return count, err
}
