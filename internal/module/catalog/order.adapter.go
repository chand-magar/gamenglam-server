package catalog

import (
	"context"

	orderservices "github.com/chand-magar/gamenglam-server/internal/module/order/services"
	"github.com/chand-magar/gamenglam-server/internal/module/catalog/services"
)

type catalogAdapter struct {
	svc services.ProductVariantService
}

func (a *catalogAdapter) GetVariantByNo(ctx context.Context, variantNo uint32) (*orderservices.CatalogVariant, error) {
	variant, err := a.svc.FindOneByNo(ctx, variantNo)
	if err != nil || variant == nil {
		return nil, err
	}
	return &orderservices.CatalogVariant{
		VariantNo:   variant.VariantNo,
		ProductName: variant.ProductName,
		SKU:         variant.SKU,
		Attributes:  variant.Attributes,
		Price:       variant.Price,
		StockQty:    int(variant.StockQty),
	}, nil
}