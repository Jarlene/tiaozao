package service

import (
	"context"
	"testing"

	"github.com/Jarlene/tiaozao/internal/model"
)

func TestCreateProductValidation(t *testing.T) {
	svc := &ProductService{}

	// Empty title
	p := &model.Product{Title: "", Price: 1, Status: model.ProductStatusActive}
	if err := svc.Create(context.Background(), p, nil, nil); err != ErrInvalidInput {
		t.Errorf("empty title: got %v, want ErrInvalidInput", err)
	}

	// Negative price
	p = &model.Product{Title: "x", Price: -1, Status: model.ProductStatusActive}
	if err := svc.Create(context.Background(), p, nil, nil); err != ErrInvalidInput {
		t.Errorf("negative price: got %v, want ErrInvalidInput", err)
	}

	// Invalid status
	p = &model.Product{Title: "x", Price: 1, Status: 99}
	if err := svc.Create(context.Background(), p, nil, nil); err != ErrInvalidInput {
		t.Errorf("invalid status: got %v, want ErrInvalidInput", err)
	}
}
