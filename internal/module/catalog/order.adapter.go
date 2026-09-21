package catalog

import (
	"context"

	"github.com/google/uuid"

	"github.com/chand-magar/gamenglam-server/internal/module/catalog/services"
	orderservices "github.com/chand-magar/gamenglam-server/internal/module/order/services"
)

type catalogAdapter struct {
	svc services.ProductVariantService
}

// GetVariantById — used by order service
func (a *catalogAdapter) GetVariantById(ctx context.Context, variantId uuid.UUID) (*orderservices.CatalogVariant, error) {
	enriched, err := a.svc.FindOneById(ctx, variantId)
	if err != nil || enriched == nil {
		return nil, err
	}
	return &orderservices.CatalogVariant{
		VariantNo:   enriched.VariantNo,
		ProductName: enriched.ProductName,
		SKU:         enriched.SKU,
		Attributes:  enriched.Attributes,
		Price:       enriched.Price,
		StockQty:    int(enriched.StockQty),
	}, nil
}
