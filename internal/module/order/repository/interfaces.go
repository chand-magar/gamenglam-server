package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/chand-magar/gamenglam-server/internal/module/order/models"
)

type CartRepository interface {
	GetByCustomerId(ctx context.Context, customerId uuid.UUID) (*models.Cart, error)
	GetCartItem(ctx context.Context, cartNo uint32, variantNo uint32) (*models.CartItem, error)
	SaveCart(ctx context.Context, customerId uuid.UUID) (*models.Cart, error)
	AddItem(ctx context.Context, cartNo uint32, req *models.AddCartItemRequest, unitPrice float64) (*models.CartItem, error)
	UpdateItem(ctx context.Context, cartItemId uuid.UUID, req *models.UpdateCartItemRequest) (*models.CartItem, error)
	RemoveItem(ctx context.Context, cartItemId uuid.UUID) error
	ClearCart(ctx context.Context, cartNo uint32) error
	GetCartWithItems(ctx context.Context, customerId uuid.UUID) (*models.Cart, error)
}

type OrderRepository interface {
	GetAll(ctx context.Context) ([]models.Order, error) // ← ADD
	Save(ctx context.Context, order *models.Order) (*models.Order, error)
	GetByCustomerId(ctx context.Context, customerId uuid.UUID) ([]models.Order, error) // ← RENAME from GetAll
	GetById(ctx context.Context, orderId uuid.UUID) (*models.Order, error)
	UpdateStatus(ctx context.Context, orderId uuid.UUID, req *models.UpdateOrderStatusRequest) (*models.Order, error)
	GetStatusHistory(ctx context.Context, orderNo uint32) ([]models.OrderStatusHistory, error)
}
