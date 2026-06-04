package services

import (
	"context"

	"github.com/google/uuid"

	"github.com/chand-magar/gamenglam-server/internal/module/order/models"
)

// CatalogProvider — what order needs from catalog module
type CatalogProvider interface {
	GetVariantByNo(ctx context.Context, variantNo uint32) (*CatalogVariant, error)
}

// CatalogVariant — order's local type, never imports catalog models directly
type CatalogVariant struct {
	VariantNo   uint32
	ProductName string
	SKU         string
	Attributes  map[string]interface{}
	Price       float64
	StockQty    int
}

// CustomerProvider — what order needs from customer module
type CustomerProvider interface {
	GetAddressByNo(ctx context.Context, addressNo uint32) (*CustomerAddress, error)
}

// CustomerAddress — order's local type
type CustomerAddress struct {
	AddressNo    uint32
	AddressLine1 string
	AddressLine2 *string
	City         string
	State        string
	Country      string
	PostalCode   *string
}

type CartService interface {
	GetCart(ctx context.Context, customerNo uint32) (*models.Cart, error)
	AddItem(ctx context.Context, customerNo uint32, req *models.AddCartItemRequest) (*models.CartItem, error)
	UpdateItem(ctx context.Context, cartItemId uuid.UUID, req *models.UpdateCartItemRequest) (*models.CartItem, error)
	RemoveItem(ctx context.Context, cartItemId uuid.UUID) error
	ClearCart(ctx context.Context, customerNo uint32) error
}

type OrderService interface {
	PlaceOrder(ctx context.Context, req *models.PlaceOrderRequest) (*models.Order, error)
	GetOrders(ctx context.Context, customerNo uint32) ([]models.Order, error)
	GetOrder(ctx context.Context, orderId uuid.UUID) (*models.Order, error)
	UpdateStatus(ctx context.Context, orderId uuid.UUID, req *models.UpdateOrderStatusRequest) (*models.Order, error)
	GetStatusHistory(ctx context.Context, orderNo uint32) ([]models.OrderStatusHistory, error)
}