package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	"github.com/chand-magar/gamenglam-server/internal/module/order/models"
	"github.com/chand-magar/gamenglam-server/internal/module/order/repository"
)

var (
	ErrOrderNotFound   = errors.New("order not found")
	ErrAddressNotFound = errors.New("address not found")
	ErrInvalidStatus   = errors.New("invalid order status transition")
)

type orderService struct {
	repo             repository.OrderRepository
	catalogProvider  CatalogProvider
	customerProvider CustomerProvider
}

func NewOrderService(
	repo repository.OrderRepository,
	catalogProvider CatalogProvider,
	customerProvider CustomerProvider,
) OrderService {
	return &orderService{
		repo:             repo,
		catalogProvider:  catalogProvider,
		customerProvider: customerProvider,
	}
}

func (s *orderService) FindAll(ctx context.Context) ([]models.Order, error) {
	orders, err := s.repo.GetAll(ctx)
	if err != nil {
		log.Printf("orderService.FindAll error: %v", err)
		return nil, fmt.Errorf("orderService.FindAll: %w", err)
	}
	return orders, nil
}

func (s *orderService) PlaceOrder(ctx context.Context, req *models.PlaceOrderRequest) (*models.Order, error) {
	// ── Validate address by UUID ──────────────────────────────────────────────
	address, err := s.customerProvider.GetAddressByID(ctx, req.AddressId)
	if err != nil {
		return nil, fmt.Errorf("orderService.PlaceOrder: validate address: %w", err)
	}
	if address == nil {
		return nil, ErrAddressNotFound
	}

	// ── Validate items and build order items ──────────────────────────────────
	var orderItems []models.OrderItem
	var subtotal float64

	for _, item := range req.Items {
		// Resolve variant by UUID — GetVariantById resolves to variant_no internally
		variant, err := s.catalogProvider.GetVariantById(ctx, item.VariantId)
		if err != nil {
			return nil, fmt.Errorf("orderService.PlaceOrder: validate variant: %w", err)
		}
		if variant == nil {
			return nil, ErrVariantNotFound
		}
		if variant.StockQty < int(item.Quantity) {
			return nil, ErrInsufficientStock
		}

		attributesJSON, err := json.Marshal(variant.Attributes)
		if err != nil {
			return nil, fmt.Errorf("orderService.PlaceOrder: marshal attributes: %w", err)
		}

		// Use unit_price from request if provided, else use variant price
		unitPrice := item.UnitPrice
		if unitPrice == 0 {
			unitPrice = variant.Price
		}

		totalPrice := unitPrice * float64(item.Quantity)
		subtotal += totalPrice

		orderItems = append(orderItems, models.OrderItem{
			VariantNo:   variant.VariantNo, // ← still store variant_no internally
			ProductName: variant.ProductName,
			SKU:         variant.SKU,
			Attributes:  attributesJSON,
			Quantity:    int(item.Quantity),
			UnitPrice:   unitPrice,
			TotalPrice:  totalPrice,
		})
	}

	// ── Build order ───────────────────────────────────────────────────────────
	total := subtotal - req.Discount + req.DeliveryFee
	order := &models.Order{
		OrderId:     uuid.New(),
		CustomerNo:  req.CustomerNo,
		AddressNo:   address.AddressNo,
		OrderRef:    generateOrderRef(),
		Subtotal:    subtotal,
		Discount:    req.Discount,
		Tax:         0,
		DeliveryFee: req.DeliveryFee,
		Total:       total,
		OrderStatus: models.OrderStatusPending,
		Notes:       req.Notes,
		Items:       orderItems,
	}

	// ── Save atomically ───────────────────────────────────────────────────────
	created, err := s.repo.Save(ctx, order)
	if err != nil {
		log.Printf("orderService.PlaceOrder error: %v", err)
		return nil, fmt.Errorf("orderService.PlaceOrder: %w", err)
	}

	return created, nil
}

func (s *orderService) GetOrders(ctx context.Context, customerId uuid.UUID) ([]models.Order, error) {
	orders, err := s.repo.GetByCustomerId(ctx, customerId)
	if err != nil {
		log.Printf("orderService.GetOrders error: %v", err)
		return nil, fmt.Errorf("orderService.GetOrders: %w", err)
	}
	return orders, nil
}

func (s *orderService) GetOrder(ctx context.Context, orderId uuid.UUID) (*models.Order, error) {
	order, err := s.repo.GetById(ctx, orderId)
	if err != nil {
		log.Printf("orderService.GetOrder error: %v", err)
		return nil, fmt.Errorf("orderService.GetOrder: %w", err)
	}
	if order == nil {
		return nil, ErrOrderNotFound
	}
	return order, nil
}

func (s *orderService) UpdateStatus(ctx context.Context, orderId uuid.UUID, req *models.UpdateOrderStatusRequest) (*models.Order, error) {
	existing, err := s.repo.GetById(ctx, orderId)
	if err != nil {
		return nil, fmt.Errorf("orderService.UpdateStatus: %w", err)
	}
	if existing == nil {
		return nil, ErrOrderNotFound
	}

	updated, err := s.repo.UpdateStatus(ctx, orderId, req)
	if err != nil {
		log.Printf("orderService.UpdateStatus error: %v", err)
		return nil, fmt.Errorf("orderService.UpdateStatus: %w", err)
	}

	return updated, nil
}

func (s *orderService) GetStatusHistory(ctx context.Context, orderNo uint32) ([]models.OrderStatusHistory, error) {
	history, err := s.repo.GetStatusHistory(ctx, orderNo)
	if err != nil {
		log.Printf("orderService.GetStatusHistory error: %v", err)
		return nil, fmt.Errorf("orderService.GetStatusHistory: %w", err)
	}
	return history, nil
}

// generateOrderRef creates a human readable order reference e.g. ORD-20260604-A3F2
func generateOrderRef() string {
	return fmt.Sprintf("ORD-%s-%s",
		time.Now().Format("20060102"),
		uuid.New().String()[:4],
	)
}
