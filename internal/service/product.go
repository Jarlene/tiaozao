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
	commentRepo  CommentStore
	imageService *ImageService
}

func NewProductService(repo *repository.ProductRepository, commentRepo CommentStore, imgSvc *ImageService) *ProductService {
	return &ProductService{
		repo:         repo,
		commentRepo:  commentRepo,
		imageService: imgSvc,
	}
}

func (s *ProductService) Create(ctx context.Context, product *model.Product, imageIDs []uint, coverImageID *uint) error {
	if product.Title == "" {
		return ErrInvalidInput
	}
	if product.Price < 0 {
		return ErrInvalidInput
	}
	if !product.Status.Valid() {
		return ErrInvalidInput
	}
	if err := s.repo.Create(ctx, product); err != nil {
		return err
	}
	if len(imageIDs) > 0 {
		return s.repo.AssociateImages(ctx, product.ID, imageIDs, coverImageID)
	}
	return nil
}

func (s *ProductService) Update(ctx context.Context, id uint, userID uint, updates map[string]any) error {
	sellerID, err := s.repo.GetProductSellerID(ctx, id)
	if err != nil {
		return ErrNotFound
	}
	if userID == 0 || sellerID != userID {
		return ErrForbidden
	}
	return s.repo.Update(ctx, id, updates)
}

func (s *ProductService) Delist(ctx context.Context, id uint, userID uint) error {
	sellerID, err := s.repo.GetProductSellerID(ctx, id)
	if err != nil {
		return ErrNotFound
	}
	if userID == 0 || sellerID != userID {
		return ErrForbidden
	}
	return s.repo.Update(ctx, id, map[string]any{"status": model.ProductStatusOffline})
}

func (s *ProductService) GetByID(ctx context.Context, id uint) (*model.Product, error) {
	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	// Attach aggregate rating summary from reviews (aggregation logic).
	if s.commentRepo != nil {
		rating, err := s.commentRepo.GetProductRating(ctx, p.ID)
		if err != nil {
			return nil, err
		}
		p.Rating = rating
	}
	return p, nil
}

func (s *ProductService) List(ctx context.Context, page, size int, categoryID *uint, status int, sellerID *uint) ([]model.Product, int64, error) {
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
