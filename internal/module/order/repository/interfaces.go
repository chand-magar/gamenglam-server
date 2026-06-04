package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/chand-magar/gamenglam-server/internal/module/order/models"
)

type CartRepository interface {
	GetByCustomerNo(ctx context.Context, customerNo uint32) (*models.Cart, error)
	GetCartItem(ctx context.Context, cartNo uint32, variantNo uint32) (*models.CartItem, error)
	SaveCart(ctx context.Context, customerNo uint32) (*models.Cart, error)
	AddItem(ctx context.Context, cartNo uint32, req *models.AddCartItemRequest, unitPrice float64) (*models.CartItem, error)
	UpdateItem(ctx context.Context, cartItemId uuid.UUID, req *models.UpdateCartItemRequest) (*models.CartItem, error)
	RemoveItem(ctx context.Context, cartItemId uuid.UUID) error
	ClearCart(ctx context.Context, cartNo uint32) error
	GetCartWithItems(ctx context.Context, customerNo uint32) (*models.Cart, error)
}

type OrderRepository interface {
	Save(ctx context.Context, order *models.Order) (*models.Order, error)
	GetAll(ctx context.Context, customerNo uint32) ([]models.Order, error)
	GetById(ctx context.Context, orderId uuid.UUID) (*models.Order, error)
	UpdateStatus(ctx context.Context, orderId uuid.UUID, req *models.UpdateOrderStatusRequest) (*models.Order, error)
	GetStatusHistory(ctx context.Context, orderNo uint32) ([]models.OrderStatusHistory, error)
}