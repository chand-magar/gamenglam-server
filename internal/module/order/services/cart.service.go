package services

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/chand-magar/gamenglam-server/internal/module/order/models"
	"github.com/chand-magar/gamenglam-server/internal/module/order/repository"
)

var (
	ErrCartItemNotFound  = errors.New("cart item not found")
	ErrInsufficientStock = errors.New("insufficient stock")
	ErrVariantNotFound   = errors.New("variant not found")
)

type cartService struct {
	repo            repository.CartRepository
	catalogProvider CatalogProvider
}

func NewCartService(repo repository.CartRepository, catalogProvider CatalogProvider) CartService {
	return &cartService{repo: repo, catalogProvider: catalogProvider}
}

func (s *cartService) GetCart(ctx context.Context, customerId uuid.UUID) (*models.Cart, error) {
	cart, err := s.repo.GetCartWithItems(ctx, customerId)
	if err != nil {
		log.Printf("cartService.GetCart error: %v", err)
		return nil, fmt.Errorf("cartService.GetCart: %w", err)
	}
	if cart == nil {
		return &models.Cart{Items: []models.CartItem{}}, nil
	}
	return cart, nil
}

func (s *cartService) AddItem(ctx context.Context, customerId uuid.UUID, req *models.AddCartItemRequest) (*models.CartItem, error) {
	variant, err := s.catalogProvider.GetVariantByNo(ctx, req.VariantNo)
	if err != nil {
		return nil, fmt.Errorf("cartService.AddItem: %w", err)
	}
	if variant == nil {
		return nil, ErrVariantNotFound
	}
	if variant.StockQty < req.Quantity {
		return nil, ErrInsufficientStock
	}

	cart, err := s.repo.GetByCustomerId(ctx, customerId)
	if err != nil {
		return nil, fmt.Errorf("cartService.AddItem: %w", err)
	}
	if cart == nil {
		cart, err = s.repo.SaveCart(ctx, customerId)
		if err != nil {
			return nil, fmt.Errorf("cartService.AddItem: create cart: %w", err)
		}
	}

	existing, err := s.repo.GetCartItem(ctx, cart.CartNo, req.VariantNo)
	if err != nil {
		return nil, fmt.Errorf("cartService.AddItem: %w", err)
	}
	if existing != nil {
		return s.repo.UpdateItem(ctx, existing.CartItemId, &models.UpdateCartItemRequest{
			Quantity: existing.Quantity + req.Quantity,
		})
	}

	return s.repo.AddItem(ctx, cart.CartNo, req, variant.Price)
}

func (s *cartService) UpdateItem(ctx context.Context, cartItemId uuid.UUID, req *models.UpdateCartItemRequest) (*models.CartItem, error) {
	item, err := s.repo.UpdateItem(ctx, cartItemId, req)
	if err != nil {
		log.Printf("cartService.UpdateItem error: %v", err)
		return nil, fmt.Errorf("cartService.UpdateItem: %w", err)
	}
	if item == nil {
		return nil, ErrCartItemNotFound
	}
	return item, nil
}

func (s *cartService) RemoveItem(ctx context.Context, cartItemId uuid.UUID) error {
	err := s.repo.RemoveItem(ctx, cartItemId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrCartItemNotFound
		}
		log.Printf("cartService.RemoveItem error: %v", err)
		return fmt.Errorf("cartService.RemoveItem: %w", err)
	}
	return nil
}

func (s *cartService) ClearCart(ctx context.Context, customerId uuid.UUID) error {
	cart, err := s.repo.GetByCustomerId(ctx, customerId)
	if err != nil {
		return fmt.Errorf("cartService.ClearCart: %w", err)
	}
	if cart == nil {
		return nil
	}
	return s.repo.ClearCart(ctx, cart.CartNo)
}