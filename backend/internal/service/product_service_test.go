package service

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/textproto"
	"testing"

	"flea-market/internal/model"
	"flea-market/internal/repository"
	errs "flea-market/pkg/errors"

	"gorm.io/gorm"
)

// ---- Mock ProductRepository ----

type mockProductRepo struct {
	createFunc                func(product *model.Product) error
	findByIDFunc              func(id uint) (*model.Product, error)
	updateFunc                func(product *model.Product) error
	deleteFunc                func(id uint) error
	listFunc                  func(page, pageSize int, status model.ProductStatus) ([]model.Product, int64, error)
	searchFunc                func(keyword string, categoryID *uint, priceMin, priceMax int64, page, pageSize int) ([]model.Product, int64, error)
	listByUserFunc            func(userID uint, status *model.ProductStatus, page, pageSize int) ([]model.Product, int64, error)
	countByUserFunc           func(userID uint) (activeCount, soldCount, inactiveCount int64, err error)
	createImageFunc           func(image *model.ProductImage) error
	updateImageFunc           func(image *model.ProductImage) error
	deleteImageFunc           func(id uint) error
	findImageByIDFunc         func(id uint) (*model.ProductImage, error)
	findImagesByIDsFunc       func(ids []uint) ([]model.ProductImage, error)
	deleteImagesByProductFunc func(productID uint) error
	listImagesByProductFunc   func(productID uint) ([]model.ProductImage, error)
	transactionFunc           func(fc func(txRepo repository.ProductRepository) error) error
}

func (m *mockProductRepo) Create(product *model.Product) error      { return m.createFunc(product) }
func (m *mockProductRepo) FindByID(id uint) (*model.Product, error) { return m.findByIDFunc(id) }
func (m *mockProductRepo) Update(product *model.Product) error      { return m.updateFunc(product) }
func (m *mockProductRepo) Delete(id uint) error                     { return m.deleteFunc(id) }
func (m *mockProductRepo) List(page, pageSize int, status model.ProductStatus) ([]model.Product, int64, error) {
	return m.listFunc(page, pageSize, status)
}
func (m *mockProductRepo) Search(keyword string, categoryID *uint, priceMin, priceMax int64, page, pageSize int) ([]model.Product, int64, error) {
	return m.searchFunc(keyword, categoryID, priceMin, priceMax, page, pageSize)
}
func (m *mockProductRepo) ListByUser(userID uint, status *model.ProductStatus, page, pageSize int) ([]model.Product, int64, error) {
	return m.listByUserFunc(userID, status, page, pageSize)
}
func (m *mockProductRepo) CountByUserAndStatus(userID uint) (activeCount, soldCount, inactiveCount int64, err error) {
	return m.countByUserFunc(userID)
}
func (m *mockProductRepo) CreateImage(image *model.ProductImage) error {
	return m.createImageFunc(image)
}
func (m *mockProductRepo) UpdateImage(image *model.ProductImage) error {
	return m.updateImageFunc(image)
}
func (m *mockProductRepo) DeleteImage(id uint) error { return m.deleteImageFunc(id) }
func (m *mockProductRepo) FindImageByID(id uint) (*model.ProductImage, error) {
	return m.findImageByIDFunc(id)
}
func (m *mockProductRepo) FindImagesByIDs(ids []uint) ([]model.ProductImage, error) {
	return m.findImagesByIDsFunc(ids)
}
func (m *mockProductRepo) DeleteImagesByProduct(productID uint) error {
	return m.deleteImagesByProductFunc(productID)
}
func (m *mockProductRepo) ListImagesByProduct(productID uint) ([]model.ProductImage, error) {
	return m.listImagesByProductFunc(productID)
}
func (m *mockProductRepo) Transaction(fc func(txRepo repository.ProductRepository) error) error {
	return m.transactionFunc(fc)
}

// ---- Mock CategoryRepository ----

type mockCategoryRepo struct {
	createFunc             func(category *model.Category) error
	updateFunc             func(category *model.Category) error
	deleteFunc             func(id uint) error
	findByIDFunc           func(id uint) (*model.Category, error)
	listAllFunc            func() ([]model.Category, error)
	listByParentIDFunc     func(parentID *uint) ([]model.Category, error)
	countSubcategoriesFunc func(id uint) (int64, error)
	countProductsFunc      func(id uint) (int64, error)
}

func (m *mockCategoryRepo) Create(category *model.Category) error     { return notImplErr("Create") }
func (m *mockCategoryRepo) Update(category *model.Category) error     { return notImplErr("Update") }
func (m *mockCategoryRepo) Delete(id uint) error                      { return notImplErr("Delete") }
func (m *mockCategoryRepo) FindByID(id uint) (*model.Category, error) { return m.findByIDFunc(id) }
func (m *mockCategoryRepo) ListAll() ([]model.Category, error)        { return notImplErr2("ListAll") }
func (m *mockCategoryRepo) ListByParentID(parentID *uint) ([]model.Category, error) {
	return notImplErr2("ListByParentID")
}
func (m *mockCategoryRepo) CountSubcategories(id uint) (int64, error) { return 0, nil }
func (m *mockCategoryRepo) CountProducts(id uint) (int64, error)      { return 0, nil }

func notImplErr(name string) error                      { panic("unexpected call: " + name) }
func notImplErr2(name string) ([]model.Category, error) { panic("unexpected call: " + name) }

// ---- Mock FileStorage ----

type mockMinIO struct{}

func (m *mockMinIO) ProductBucket() string { return "products" }
func (m *mockMinIO) UploadFile(bucket string, file io.Reader, header *multipart.FileHeader) (string, error) {
	return "test/object-key.jpg", nil
}
func (m *mockMinIO) DeleteFile(bucket string, objectKey string) error { return nil }
func (m *mockMinIO) GetFileURL(bucket string, objectKey string) string {
	return "http://minio:9000/" + bucket + "/" + objectKey
}

// ---- Helpers ----

func makeProduct(id uint, userID uint, title string, price int64, status model.ProductStatus) *model.Product {
	return &model.Product{ID: id, Title: title, Price: price, Status: status, UserID: userID}
}

func newTestService(p mockProductRepo, c mockCategoryRepo) *ProductService {
	return &ProductService{productRepo: &p, categoryRepo: &c, minio: &mockMinIO{}}
}

// ---- Tests ----

func TestCreateProduct_Success(t *testing.T) {
	catID := uint(1)
	svc := newTestService(
		mockProductRepo{
			createFunc: func(product *model.Product) error { product.ID = 1; return nil },
			transactionFunc: func(fc func(txRepo repository.ProductRepository) error) error {
				return fc(&mockProductRepo{createFunc: func(product *model.Product) error { product.ID = 1; return nil }})
			},
		},
		mockCategoryRepo{findByIDFunc: func(id uint) (*model.Category, error) { return &model.Category{ID: id, Name: "Test"}, nil }},
	)

	req := &CreateProductReq{Title: "测试商品", Description: "描述", Price: 1000, CategoryID: &catID}
	product, code, err := svc.Create(1, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != errs.Success {
		t.Fatalf("expected success code %d, got %d", errs.Success, code)
	}
	if product.Title != "测试商品" || product.Price != 1000 {
		t.Fatalf("product mismatch: title=%s price=%d", product.Title, product.Price)
	}
	if product.Status != model.ProductStatusActive {
		t.Fatalf("expected status active (%d), got %d", model.ProductStatusActive, product.Status)
	}
}

func TestCreateProduct_CategoryNotFound(t *testing.T) {
	svc := newTestService(
		mockProductRepo{},
		mockCategoryRepo{findByIDFunc: func(id uint) (*model.Category, error) { return nil, gorm.ErrRecordNotFound }},
	)
	_, code, err := svc.Create(1, &CreateProductReq{Title: "测试", Price: 1000, CategoryID: uintPtr(999)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != errs.ErrCategoryNotFound {
		t.Fatalf("expected ErrCategoryNotFound (%d), got %d", errs.ErrCategoryNotFound, code)
	}
}

func TestGetByID_Success(t *testing.T) {
	svc := newTestService(
		mockProductRepo{findByIDFunc: func(id uint) (*model.Product, error) {
			return makeProduct(id, 1, "商品A", 500, model.ProductStatusActive), nil
		}},
		mockCategoryRepo{},
	)
	product, code, err := svc.GetByID(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != errs.Success {
		t.Fatalf("expected success, got %d", code)
	}
	if product.Title != "商品A" {
		t.Fatalf("expected '商品A', got '%s'", product.Title)
	}
}

func TestGetByID_NotFound(t *testing.T) {
	svc := newTestService(
		mockProductRepo{findByIDFunc: func(id uint) (*model.Product, error) { return nil, gorm.ErrRecordNotFound }},
		mockCategoryRepo{},
	)
	_, code, err := svc.GetByID(999)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != errs.ErrProductNotFound {
		t.Fatalf("expected ErrProductNotFound (%d), got %d", errs.ErrProductNotFound, code)
	}
}

func TestUpdate_Success(t *testing.T) {
	var prod *model.Product
	svc := newTestService(
		mockProductRepo{
			findByIDFunc: func(id uint) (*model.Product, error) {
				if prod == nil {
					prod = makeProduct(id, 1, "旧标题", 500, model.ProductStatusActive)
				}
				return prod, nil
			},
			updateFunc: func(product *model.Product) error { prod = product; return nil },
		},
		mockCategoryRepo{},
	)
	product, code, err := svc.Update(1, 1, &UpdateProductReq{Title: "新标题", Price: 2000})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != errs.Success || product.Title != "新标题" || product.Price != 2000 {
		t.Fatalf("update failed: code=%d title=%s price=%d", code, product.Title, product.Price)
	}
}

func TestUpdate_Forbidden(t *testing.T) {
	svc := newTestService(
		mockProductRepo{findByIDFunc: func(id uint) (*model.Product, error) {
			return makeProduct(id, 2, "商品", 500, model.ProductStatusActive), nil
		}},
		mockCategoryRepo{},
	)
	_, code, err := svc.Update(1, 1, &UpdateProductReq{Title: "hack", Price: 100})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != errs.ErrForbidden {
		t.Fatalf("expected ErrForbidden (%d), got %d", errs.ErrForbidden, code)
	}
}

func TestUpdate_NotFound(t *testing.T) {
	svc := newTestService(
		mockProductRepo{findByIDFunc: func(id uint) (*model.Product, error) { return nil, gorm.ErrRecordNotFound }},
		mockCategoryRepo{},
	)
	_, code, err := svc.Update(1, 999, &UpdateProductReq{Title: "新", Price: 100})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != errs.ErrProductNotFound {
		t.Fatalf("expected ErrProductNotFound (%d), got %d", errs.ErrProductNotFound, code)
	}
}

func TestUpdate_WithNewCategory(t *testing.T) {
	catID := uint(5)
	var prod *model.Product
	svc := newTestService(
		mockProductRepo{
			findByIDFunc: func(id uint) (*model.Product, error) {
				if prod == nil {
					prod = makeProduct(id, 1, "商品", 500, model.ProductStatusActive)
				}
				return prod, nil
			},
			updateFunc: func(product *model.Product) error { prod = product; return nil },
		},
		mockCategoryRepo{findByIDFunc: func(id uint) (*model.Category, error) { return &model.Category{ID: id, Name: "电子"}, nil }},
	)
	product, code, err := svc.Update(1, 1, &UpdateProductReq{Title: "商品", Price: 500, CategoryID: &catID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != errs.Success {
		t.Fatalf("expected success, got %d", code)
	}
	if product.CategoryID == nil || *product.CategoryID != catID {
		t.Fatalf("expected category_id %d", catID)
	}
}

func TestDelete_Success(t *testing.T) {
	svc := newTestService(
		mockProductRepo{
			findByIDFunc: func(id uint) (*model.Product, error) {
				p := makeProduct(id, 1, "商品", 500, model.ProductStatusActive)
				p.Images = []model.ProductImage{}
				return p, nil
			},
			updateFunc: func(product *model.Product) error {
				if product.Status != model.ProductStatusInactive {
					t.Fatal("status should become inactive on delete")
				}
				return nil
			},
		},
		mockCategoryRepo{},
	)
	code, err := svc.Delete(1, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != errs.Success {
		t.Fatalf("expected success, got %d", code)
	}
}

func TestDelete_Forbidden(t *testing.T) {
	svc := newTestService(
		mockProductRepo{findByIDFunc: func(id uint) (*model.Product, error) {
			return makeProduct(id, 2, "商品", 500, model.ProductStatusActive), nil
		}},
		mockCategoryRepo{},
	)
	code, err := svc.Delete(1, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != errs.ErrForbidden {
		t.Fatalf("expected ErrForbidden (%d), got %d", errs.ErrForbidden, code)
	}
}

func TestDelete_NotFound(t *testing.T) {
	svc := newTestService(
		mockProductRepo{findByIDFunc: func(id uint) (*model.Product, error) { return nil, gorm.ErrRecordNotFound }},
		mockCategoryRepo{},
	)
	code, err := svc.Delete(1, 999)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != errs.ErrProductNotFound {
		t.Fatalf("expected ErrProductNotFound (%d), got %d", errs.ErrProductNotFound, code)
	}
}

func TestList_Pagination(t *testing.T) {
	svc := newTestService(
		mockProductRepo{
			listFunc: func(page, pageSize int, status model.ProductStatus) ([]model.Product, int64, error) {
				if status != model.ProductStatusActive {
					t.Fatalf("expected active filter, got %d", status)
				}
				products := make([]model.Product, 2)
				for i := range products {
					products[i] = model.Product{ID: uint(i + 1), Title: "商品", Price: 1000, Status: model.ProductStatusActive}
				}
				return products, 25, nil
			},
		},
		mockCategoryRepo{},
	)
	result, code, err := svc.List(1, 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != errs.Success {
		t.Fatalf("expected success, got %d", code)
	}
	if result.Total != 25 || result.TotalPages != 2 {
		t.Fatalf("expected total=25 pages=2, got total=%d pages=%d", result.Total, result.TotalPages)
	}
	if len(result.Items.([]ProductListItem)) != 2 {
		t.Fatalf("expected 2 items, got %d", len(result.Items.([]ProductListItem)))
	}
}

func TestList_InvalidPageDefaults(t *testing.T) {
	svc := newTestService(
		mockProductRepo{
			listFunc: func(page, pageSize int, status model.ProductStatus) ([]model.Product, int64, error) {
				if page != 1 || pageSize != 20 {
					t.Fatalf("expected page=1 pageSize=20, got page=%d pageSize=%d", page, pageSize)
				}
				return []model.Product{}, 0, nil
			},
		},
		mockCategoryRepo{},
	)
	_, code, err := svc.List(0, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != errs.Success {
		t.Fatalf("expected success, got %d", code)
	}
}

func TestList_ClampPageSize(t *testing.T) {
	svc := newTestService(
		mockProductRepo{
			listFunc: func(page, pageSize int, status model.ProductStatus) ([]model.Product, int64, error) {
				if pageSize > 50 {
					t.Fatalf("pageSize should be clamped to 50, got %d", pageSize)
				}
				return []model.Product{}, 0, nil
			},
		},
		mockCategoryRepo{},
	)
	_, code, err := svc.List(1, 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != errs.Success {
		t.Fatalf("expected success, got %d", code)
	}
}

func TestListMyProducts_Success(t *testing.T) {
	userID := uint(1)
	activeStatus := model.ProductStatusActive
	svc := newTestService(
		mockProductRepo{
			listByUserFunc: func(uid uint, s *model.ProductStatus, page, pageSize int) ([]model.Product, int64, error) {
				if uid != userID || s == nil || *s != activeStatus {
					t.Fatalf("unexpected filter params")
				}
				return []model.Product{{ID: 1, Title: "我的商品", Price: 1500, Status: model.ProductStatusActive, UserID: userID}}, 1, nil
			},
		},
		mockCategoryRepo{},
	)
	result, code, err := svc.ListMyProducts(userID, &activeStatus, 1, 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != errs.Success || result.Total != 1 {
		t.Fatalf("expected success total=1, got code=%d total=%d", code, result.Total)
	}
}

func TestListMyProducts_NoStatusFilter(t *testing.T) {
	svc := newTestService(
		mockProductRepo{
			listByUserFunc: func(uid uint, s *model.ProductStatus, page, pageSize int) ([]model.Product, int64, error) {
				if s != nil {
					t.Fatal("expected nil status filter for 'all'")
				}
				return []model.Product{
					{ID: 1, Title: "商品1", Price: 1000, Status: model.ProductStatusActive, UserID: uid},
					{ID: 2, Title: "商品2", Price: 2000, Status: model.ProductStatusInactive, UserID: uid},
				}, 2, nil
			},
		},
		mockCategoryRepo{},
	)
	result, code, err := svc.ListMyProducts(1, nil, 1, 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != errs.Success || result.Total != 2 {
		t.Fatalf("expected success total=2, got code=%d total=%d", code, result.Total)
	}
}

func TestGetMyCounts_Success(t *testing.T) {
	svc := newTestService(
		mockProductRepo{countByUserFunc: func(uid uint) (int64, int64, int64, error) { return 10, 3, 2, nil }},
		mockCategoryRepo{},
	)
	counts, code, err := svc.GetMyCounts(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != errs.Success {
		t.Fatalf("expected success, got %d", code)
	}
	if counts["active"] != 10 || counts["sold"] != 3 || counts["inactive"] != 2 {
		t.Fatalf("unexpected counts: %v", counts)
	}
}

func TestUpdateStatus_ActivateAndDeactivate(t *testing.T) {
	t.Run("deactivate", func(t *testing.T) {
		svc := newTestService(
			mockProductRepo{
				findByIDFunc: func(id uint) (*model.Product, error) {
					return makeProduct(id, 1, "商品", 1000, model.ProductStatusActive), nil
				},
				updateFunc: func(product *model.Product) error {
					if product.Status != model.ProductStatusInactive {
						t.Fatal("expected inactive")
					}
					return nil
				},
			},
			mockCategoryRepo{},
		)
		code, err := svc.UpdateStatus(1, 1, model.ProductStatusInactive)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if code != errs.Success {
			t.Fatalf("expected success, got %d", code)
		}
	})
	t.Run("reactivate", func(t *testing.T) {
		svc := newTestService(
			mockProductRepo{
				findByIDFunc: func(id uint) (*model.Product, error) {
					return makeProduct(id, 1, "商品", 1000, model.ProductStatusInactive), nil
				},
				updateFunc: func(product *model.Product) error {
					if product.Status != model.ProductStatusActive {
						t.Fatal("expected active on reactivation")
					}
					return nil
				},
			},
			mockCategoryRepo{},
		)
		code, err := svc.UpdateStatus(1, 1, model.ProductStatusActive)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if code != errs.Success {
			t.Fatalf("expected success, got %d", code)
		}
	})
	t.Run("mark as sold", func(t *testing.T) {
		svc := newTestService(
			mockProductRepo{
				findByIDFunc: func(id uint) (*model.Product, error) {
					return makeProduct(id, 1, "商品", 1000, model.ProductStatusActive), nil
				},
				updateFunc: func(product *model.Product) error {
					if product.Status != model.ProductStatusSold {
						t.Fatal("expected sold")
					}
					return nil
				},
			},
			mockCategoryRepo{},
		)
		code, err := svc.UpdateStatus(1, 1, model.ProductStatusSold)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if code != errs.Success {
			t.Fatalf("expected success, got %d", code)
		}
	})
}

func TestUpdateStatus_NotFound(t *testing.T) {
	svc := newTestService(
		mockProductRepo{findByIDFunc: func(id uint) (*model.Product, error) { return nil, gorm.ErrRecordNotFound }},
		mockCategoryRepo{},
	)
	code, err := svc.UpdateStatus(1, 999, model.ProductStatusInactive)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != errs.ErrProductNotFound {
		t.Fatalf("expected ErrProductNotFound (%d), got %d", errs.ErrProductNotFound, code)
	}
}

func TestUpdateStatus_Forbidden(t *testing.T) {
	svc := newTestService(
		mockProductRepo{findByIDFunc: func(id uint) (*model.Product, error) {
			return makeProduct(id, 2, "商品", 1000, model.ProductStatusActive), nil
		}},
		mockCategoryRepo{},
	)
	code, err := svc.UpdateStatus(1, 1, model.ProductStatusInactive)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != errs.ErrForbidden {
		t.Fatalf("expected ErrForbidden (%d), got %d", errs.ErrForbidden, code)
	}
}

func TestSearch_WithKeyword(t *testing.T) {
	svc := newTestService(
		mockProductRepo{
			searchFunc: func(keyword string, categoryID *uint, priceMin, priceMax int64, page, pageSize int) ([]model.Product, int64, error) {
				if keyword != "手机" {
					t.Fatalf("expected '手机', got '%s'", keyword)
				}
				return []model.Product{{ID: 1, Title: "手机", Price: 200000, Status: model.ProductStatusActive}}, 1, nil
			},
		},
		mockCategoryRepo{},
	)
	result, code, err := svc.Search("手机", nil, 0, 0, 1, 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != errs.Success || result.Total != 1 {
		t.Fatalf("expected success total=1, got code=%d total=%d", code, result.Total)
	}
}

func TestSearch_WithPriceAndCategory(t *testing.T) {
	catID := uint(1)
	svc := newTestService(
		mockProductRepo{
			searchFunc: func(keyword string, categoryID *uint, priceMin, priceMax int64, page, pageSize int) ([]model.Product, int64, error) {
				if priceMin != 1000 || priceMax != 5000 {
					t.Fatalf("price range mismatch: min=%d max=%d", priceMin, priceMax)
				}
				if categoryID == nil || *categoryID != 1 {
					t.Fatalf("expected categoryID 1")
				}
				return []model.Product{{ID: 1, Title: "商品", Price: 2000, Status: model.ProductStatusActive}}, 1, nil
			},
		},
		mockCategoryRepo{},
	)
	result, code, err := svc.Search("", &catID, 1000, 5000, 1, 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != errs.Success || result.Total != 1 {
		t.Fatalf("expected success total=1, got code=%d total=%d", code, result.Total)
	}
}

func TestGetImage_Success(t *testing.T) {
	svc := newTestService(
		mockProductRepo{findImageByIDFunc: func(id uint) (*model.ProductImage, error) {
			return &model.ProductImage{ID: id, ObjectKey: "img/abc.jpg", SortOrder: 0}, nil
		}},
		mockCategoryRepo{},
	)
	image, code, err := svc.GetImage(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != errs.Success {
		t.Fatalf("expected success, got %d", code)
	}
	if image.URL == "" {
		t.Fatal("expected non-empty URL")
	}
}

func TestGetImage_NotFound(t *testing.T) {
	svc := newTestService(
		mockProductRepo{findImageByIDFunc: func(id uint) (*model.ProductImage, error) { return nil, gorm.ErrRecordNotFound }},
		mockCategoryRepo{},
	)
	_, code, err := svc.GetImage(999)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != errs.ErrImageNotFound {
		t.Fatalf("expected ErrImageNotFound (%d), got %d", errs.ErrImageNotFound, code)
	}
}

func TestDeleteImage_Success(t *testing.T) {
	svc := newTestService(
		mockProductRepo{
			findImageByIDFunc: func(id uint) (*model.ProductImage, error) {
				return &model.ProductImage{ID: id, UserID: 1, ObjectKey: "img/abc.jpg"}, nil
			},
			deleteImageFunc: func(id uint) error { return nil },
		},
		mockCategoryRepo{},
	)
	code, err := svc.DeleteImage(1, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != errs.Success {
		t.Fatalf("expected success, got %d", code)
	}
}

func TestDeleteImage_Forbidden(t *testing.T) {
	svc := newTestService(
		mockProductRepo{findImageByIDFunc: func(id uint) (*model.ProductImage, error) {
			return &model.ProductImage{ID: id, UserID: 2, ObjectKey: "img/abc.jpg"}, nil
		}},
		mockCategoryRepo{},
	)
	code, err := svc.DeleteImage(1, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != errs.ErrForbidden {
		t.Fatalf("expected ErrForbidden (%d), got %d", errs.ErrForbidden, code)
	}
}

func TestDeleteImage_NotFound(t *testing.T) {
	svc := newTestService(
		mockProductRepo{findImageByIDFunc: func(id uint) (*model.ProductImage, error) { return nil, gorm.ErrRecordNotFound }},
		mockCategoryRepo{},
	)
	code, err := svc.DeleteImage(1, 999)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != errs.ErrImageNotFound {
		t.Fatalf("expected ErrImageNotFound (%d), got %d", errs.ErrImageNotFound, code)
	}
}

func TestAllowedImageTypes(t *testing.T) {
	for _, mime := range []string{"image/jpeg", "image/png", "image/webp"} {
		if _, ok := allowedImageTypes[mime]; !ok {
			t.Fatalf("%s should be allowed", mime)
		}
	}
	for _, mime := range []string{"image/gif", "image/bmp", "image/svg+xml", "application/pdf"} {
		if _, ok := allowedImageTypes[mime]; ok {
			t.Fatalf("%s should NOT be allowed", mime)
		}
	}
}

func uintPtr(v uint) *uint { return &v }

// testFile 模拟 multipart.File，用于 UploadImage 测试
type testFile struct {
	*bytes.Reader
}

func (f *testFile) ReadAt(p []byte, off int64) (int, error) { return f.Reader.ReadAt(p, off) }
func (f *testFile) Seek(offset int64, whence int) (int64, error) {
	return f.Reader.Seek(offset, whence)
}
func (f *testFile) Close() error { return nil }

// JPEG 文件头（前 512 字节中应包含 FF D8 FF）
var jpegHeader = append([]byte{0xFF, 0xD8, 0xFF, 0xE0}, bytes.Repeat([]byte{0x00}, 508)...)

// PNG 文件头
var pngHeader = append([]byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}, bytes.Repeat([]byte{0x00}, 504)...)

// 非图片文件头
var textHeader = []byte("this is not an image file")

func makeUploadHeader(filename, contentType string, size int) *multipart.FileHeader {
	return &multipart.FileHeader{
		Filename: filename,
		Header:   textproto.MIMEHeader{"Content-Type": {contentType}},
		Size:     int64(size),
	}
}

func TestUploadImage_Success(t *testing.T) {
	svc := newTestService(
		mockProductRepo{
			createImageFunc: func(image *model.ProductImage) error {
				image.ID = 1
				return nil
			},
		},
		mockCategoryRepo{},
	)

	t.Run("jpeg", func(t *testing.T) {
		file := &testFile{bytes.NewReader(jpegHeader)}
		header := makeUploadHeader("test.jpg", "image/jpeg", len(jpegHeader))
		image, code, err := svc.UploadImage(1, file, header)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if code != errs.Success {
			t.Fatalf("expected success, got %d", code)
		}
		if image.URL == "" {
			t.Fatal("expected non-empty URL")
		}
	})

	t.Run("png", func(t *testing.T) {
		file := &testFile{bytes.NewReader(pngHeader)}
		header := makeUploadHeader("test.png", "image/png", len(pngHeader))
		image, code, err := svc.UploadImage(1, file, header)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if code != errs.Success {
			t.Fatalf("expected success, got %d", code)
		}
		if image.URL == "" {
			t.Fatal("expected non-empty URL")
		}
	})

	t.Run("webp by extension fallback", func(t *testing.T) {
		// Use small content that won't be detected as image; rely on extension fallback
		data := []byte("RIFF\x00\x00\x00\x00WEBPVP8 ")
		data = append(data, bytes.Repeat([]byte{0x00}, 500)...)
		file := &testFile{bytes.NewReader(data)}
		header := makeUploadHeader("test.webp", "application/octet-stream", len(data))
		image, code, err := svc.UploadImage(1, file, header)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if code != errs.Success {
			t.Fatalf("expected success, got %d", code)
		}
		if image.URL == "" {
			t.Fatal("expected non-empty URL")
		}
	})
}

func TestUploadImage_InvalidContent(t *testing.T) {
	svc := newTestService(mockProductRepo{}, mockCategoryRepo{})

	t.Run("non-image content and non-image extension rejected", func(t *testing.T) {
		file := &testFile{bytes.NewReader(textHeader)}
		header := makeUploadHeader("readme.txt", "text/plain", len(textHeader))
		_, code, err := svc.UploadImage(1, file, header)
		if err != nil {
			t.Fatal("unexpected error:", err)
		}
		if code != errs.ErrFileFormat {
			t.Fatalf("expected ErrFileFormat (%d), got %d", errs.ErrFileFormat, code)
		}
	})
}

func TestUploadImage_FileTooLarge(t *testing.T) {
	svc := newTestService(mockProductRepo{}, mockCategoryRepo{})
	data := make([]byte, 6*1024*1024) // 6MB > 5MB limit
	file := &testFile{bytes.NewReader(data)}
	header := makeUploadHeader("large.jpg", "image/jpeg", len(data))
	_, code, err := svc.UploadImage(1, file, header)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != errs.ErrFileTooLarge {
		t.Fatalf("expected ErrFileTooLarge (%d), got %d", errs.ErrFileTooLarge, code)
	}
}
