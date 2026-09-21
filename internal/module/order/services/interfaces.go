package services

import (
	"context"

	"github.com/google/uuid"

	"github.com/chand-magar/gamenglam-server/internal/module/order/models"
)

type CatalogProvider interface {
	GetVariantById(ctx context.Context, variantId uuid.UUID) (*CatalogVariant, error)
}

type CatalogVariant struct {
	VariantNo   uint32
	ProductName string
	SKU         string
	Attributes  map[string]interface{}
	Price       float64
	StockQty    int
}

type CustomerProvider interface {
	GetAddressByNo(ctx context.Context, addressNo uint32) (*CustomerAddress, error)
	GetAddressByID(ctx context.Context, addressId uuid.UUID) (*CustomerAddress, error)
}

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
	GetCart(ctx context.Context, customerId uuid.UUID) (*models.Cart, error)
	AddItem(ctx context.Context, customerId uuid.UUID, req *models.AddCartItemRequest) (*models.CartItem, error)
	UpdateItem(ctx context.Context, cartItemId uuid.UUID, req *models.UpdateCartItemRequest) (*models.CartItem, error)
	RemoveItem(ctx context.Context, cartItemId uuid.UUID) error
	ClearCart(ctx context.Context, customerId uuid.UUID) error
}

type OrderService interface {
	FindAll(ctx context.Context) ([]models.Order, error)
	PlaceOrder(ctx context.Context, req *models.PlaceOrderRequest) (*models.Order, error)
	GetOrders(ctx context.Context, customerId uuid.UUID) ([]models.Order, error)
	GetOrder(ctx context.Context, orderId uuid.UUID) (*models.Order, error)
	UpdateStatus(ctx context.Context, orderId uuid.UUID, req *models.UpdateOrderStatusRequest) (*models.Order, error)
	GetStatusHistory(ctx context.Context, orderNo uint32) ([]models.OrderStatusHistory, error)
}
