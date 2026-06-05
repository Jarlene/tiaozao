package service

import (
	stdErrors "errors"

	"flea-market/internal/model"
	"flea-market/internal/repository"
	"flea-market/pkg/errors"

	"gorm.io/gorm"
)

type CategoryService struct {
	categoryRepo repository.CategoryRepository
}

func NewCategoryService(categoryRepo repository.CategoryRepository) *CategoryService {
	return &CategoryService{categoryRepo: categoryRepo}
}

type CreateCategoryReq struct {
	Name     string `json:"name" binding:"required,max=100"`
	ParentID *uint  `json:"parent_id"`
}

type UpdateCategoryReq struct {
	Name string `json:"name" binding:"required,max=100"`
}

type CategoryTreeItem struct {
	ID       uint                `json:"id"`
	Name     string              `json:"name"`
	ParentID *uint               `json:"parent_id"`
	Children []CategoryTreeItem  `json:"children,omitempty"`
}

// Create 创建分类
func (s *CategoryService) Create(req *CreateCategoryReq) (*model.Category, int, error) {
	// 如果指定了父分类，验证父分类存在
	if req.ParentID != nil {
		if _, err := s.categoryRepo.FindByID(*req.ParentID); err != nil {
			if stdErrors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.ErrCategoryNotFound, nil
			}
			return nil, errors.ErrInternal, err
		}
	}

	category := &model.Category{
		Name:     req.Name,
		ParentID: req.ParentID,
	}

	if err := s.categoryRepo.Create(category); err != nil {
		return nil, errors.ErrInternal, err
	}

	return category, errors.Success, nil
}

// Update 更新分类
func (s *CategoryService) Update(id uint, req *UpdateCategoryReq) (*model.Category, int, error) {
	category, err := s.categoryRepo.FindByID(id)
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.ErrCategoryNotFound, nil
		}
		return nil, errors.ErrInternal, err
	}

	category.Name = req.Name
	if err := s.categoryRepo.Update(category); err != nil {
		return nil, errors.ErrInternal, err
	}

	return category, errors.Success, nil
}

// Delete 删除分类
func (s *CategoryService) Delete(id uint) (int, error) {
	if _, err := s.categoryRepo.FindByID(id); err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			return errors.ErrCategoryNotFound, nil
		}
		return errors.ErrInternal, err
	}

	// 检查是否有子分类
	subCount, err := s.categoryRepo.CountSubcategories(id)
	if err != nil {
		return errors.ErrInternal, err
	}
	if subCount > 0 {
		return errors.ErrCategoryHasChildren, nil
	}

	// 检查是否有商品
	prodCount, err := s.categoryRepo.CountProducts(id)
	if err != nil {
		return errors.ErrInternal, err
	}
	if prodCount > 0 {
		return errors.ErrCategoryHasProducts, nil
	}

	if err := s.categoryRepo.Delete(id); err != nil {
		return errors.ErrInternal, err
	}

	return errors.Success, nil
}

// GetTree 获取分类树
func (s *CategoryService) GetTree() ([]CategoryTreeItem, int, error) {
	categories, err := s.categoryRepo.ListAll()
	if err != nil {
		return nil, errors.ErrInternal, err
	}

	tree := s.buildTree(categories, nil)
	return tree, errors.Success, nil
}

// GetFlatList 获取扁平分类列表
func (s *CategoryService) GetFlatList() ([]model.Category, int, error) {
	categories, err := s.categoryRepo.ListAll()
	if err != nil {
		return nil, errors.ErrInternal, err
	}
	return categories, errors.Success, nil
}

// buildTree 构建树形结构（O(n) 算法，使用 map 按 parent_id 分组）
func (s *CategoryService) buildTree(categories []model.Category, parentID *uint) []CategoryTreeItem {
	childrenMap := make(map[uint][]model.Category)
	var roots []model.Category
	for _, c := range categories {
		if c.ParentID == nil {
			roots = append(roots, c)
		} else {
			childrenMap[*c.ParentID] = append(childrenMap[*c.ParentID], c)
		}
	}

	var build func(pid *uint) []CategoryTreeItem
	build = func(pid *uint) []CategoryTreeItem {
		var items []model.Category
		if pid == nil {
			items = roots
		} else {
			items = childrenMap[*pid]
		}
		tree := make([]CategoryTreeItem, 0, len(items))
		for _, c := range items {
			children := build(&c.ID)
			item := CategoryTreeItem{
				ID:       c.ID,
				Name:     c.Name,
				ParentID: c.ParentID,
			}
			if len(children) > 0 {
				item.Children = children
			}
			tree = append(tree, item)
		}
		return tree
	}

	return build(parentID)
}
