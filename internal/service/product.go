package service

import (
	"context"
	"errors"
	"io"

	"github.com/Jarlene/tiaozao/internal/model"
	"github.com/Jarlene/tiaozao/internal/repository"
	"github.com/jackc/pgx/v5"
)

type ProductService struct {
	repo         *repository.ProductRepository
	imageService *ImageService
}

func NewProductService(repo *repository.ProductRepository, imgSvc *ImageService) *ProductService {
	return &ProductService{
		repo:         repo,
		imageService: imgSvc,
	}
}

func (s *ProductService) Create(ctx context.Context, product *model.Product, imageIDs []uint, coverImageID *uint) error {
	if product.Title == "" {
		return errors.New("title is required")
	}
	if product.Price < 0 {
		return errors.New("price must be non-negative")
	}
	if err := s.repo.Create(ctx, product); err != nil {
		return err
	}
	if len(imageIDs) > 0 {
		return s.repo.AssociateImages(ctx, product.ID, imageIDs, coverImageID)
	}
	return nil
}

func (s *ProductService) Update(ctx context.Context, id uint, userID string, updates map[string]any) error {
	sellerID, err := s.repo.GetProductSellerID(ctx, id)
	if err != nil {
		return errors.New("product not found")
	}
	uid, _ := parseUint(userID)
	if uid != 0 && sellerID != uid {
		return errors.New("forbidden")
	}
	return s.repo.Update(ctx, id, updates)
}

func (s *ProductService) Delist(ctx context.Context, id uint, userID string) error {
	return s.Update(ctx, id, userID, map[string]any{"status": model.ProductStatusInactive})
}

func (s *ProductService) GetByID(ctx context.Context, id uint) (*model.Product, error) {
	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("product not found")
		}
		return nil, err
	}
	return p, nil
}

func (s *ProductService) List(ctx context.Context, page, size int, categoryID *uint, status string, sellerID *uint) ([]model.Product, int64, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	q := repository.ProductListQuery{
		Page:       page,
		Size:       size,
		CategoryID: categoryID,
		Status:     status,
		SellerID:   sellerID,
	}
	return s.repo.List(ctx, q)
}

func (s *ProductService) SaveImage(ctx context.Context, filename string, reader io.Reader) (string, error) {
	return s.imageService.Save(ctx, filename, reader)
}

func (s *ProductService) CreateImageRecord(ctx context.Context, url string) (uint, error) {
	return s.repo.CreateImage(ctx, url)
}

func parseUint(s string) (uint, error) {
	var id uint
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, errors.New("invalid uint")
		}
		id = id*10 + uint(c-'0')
	}
	return id, nil
}
