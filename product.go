package facturino

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

// Product is a catalog product or service.
type Product struct {
	ID       string `json:"id"`
	Object   string `json:"object"`
	Livemode bool   `json:"livemode"`

	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Reference   string `json:"reference,omitempty"`
	Category    string `json:"category,omitempty"`

	UnitPrice string `json:"unitPrice"`
	VATRate   string `json:"vatRate"`
	VATCode   string `json:"vatCode"`
	Unit      string `json:"unit"`

	Tags []string `json:"tags"`

	PriceHistory []*PriceHistoryEntry `json:"priceHistory"`

	Active  bool   `json:"active"`
	Created string `json:"created"`
	Updated string `json:"updated"`
}

// PriceHistoryEntry records a price change.
type PriceHistoryEntry struct {
	Price     string `json:"price"`
	VATRate   string `json:"vatRate"`
	ChangedAt string `json:"changedAt"`
	ChangedBy string `json:"changedBy"`
}

// ProductParams are the parameters for creating a product. UnitPrice in centimes, VATRate in centipercent.
type ProductParams struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Reference   string   `json:"reference,omitempty"`
	Category    string   `json:"category,omitempty"`
	UnitPrice   int      `json:"unitPrice"`
	VATRate     int      `json:"vatRate"`
	VATCode     string   `json:"vatCode,omitempty"`
	Unit        string   `json:"unit,omitempty"`
	Tags        []string `json:"tags,omitempty"`

	IdempotencyKey string `json:"-"`
}

// ProductUpdateParams are the parameters for updating a product.
type ProductUpdateParams struct {
	Name        string   `json:"name,omitempty"`
	Description string   `json:"description,omitempty"`
	Reference   string   `json:"reference,omitempty"`
	Category    string   `json:"category,omitempty"`
	UnitPrice   int      `json:"unitPrice,omitempty"`
	VATRate     int      `json:"vatRate,omitempty"`
	VATCode     string   `json:"vatCode,omitempty"`
	Unit        string   `json:"unit,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Active      *bool    `json:"active,omitempty"`
}

// ProductService operates on products.
type ProductService struct {
	client *httpClient
}

// Create creates a new product.
func (s *ProductService) Create(params *ProductParams) (*Product, error) {
	var prod Product
	opts := &requestOption{}
	if params.IdempotencyKey != "" {
		opts.idempotencyKey = params.IdempotencyKey
	}
	err := s.client.post("/products", params, &prod, opts)
	if err != nil {
		return nil, err
	}
	return &prod, nil
}

// Get retrieves a product by ID.
func (s *ProductService) Get(id string) (*Product, error) {
	var prod Product
	err := s.client.get(fmt.Sprintf("/products/%s", id), nil, &prod)
	if err != nil {
		return nil, err
	}
	return &prod, nil
}

// Update updates a product.
func (s *ProductService) Update(id string, params *ProductUpdateParams) (*Product, error) {
	var prod Product
	err := s.client.patch(fmt.Sprintf("/products/%s", id), params, &prod)
	if err != nil {
		return nil, err
	}
	return &prod, nil
}

// Delete soft-deletes a product.
func (s *ProductService) Delete(id string) error {
	return s.client.del(fmt.Sprintf("/products/%s", id))
}

// ProductListParams adds product-specific filters to ListParams.
type ProductListParams struct {
	ListParams

	// Q filters products by name prefix.
	Q string
	// Category filters products by category.
	Category string
	// Active filters products by their active flag. Leave nil to return
	// both active and inactive products.
	Active *bool
}

// List returns a paginated iterator over products.
func (s *ProductService) List(params *ProductListParams) *ProductIterator {
	var lp *ListParams
	if params != nil {
		lp = &params.ListParams
	}
	iter := newIterator[*Product](s.client, "/products", lp, decodeProduct)
	if params != nil {
		extra := url.Values{}
		if params.Q != "" {
			extra.Set("q", params.Q)
		}
		if params.Category != "" {
			extra.Set("category", params.Category)
		}
		if params.Active != nil {
			extra.Set("active", strconv.FormatBool(*params.Active))
		}
		if len(extra) > 0 {
			iter.extraParams = extra
		}
	}
	return &ProductIterator{iter: iter}
}

// ImportCSV imports products from base64-encoded CSV content.
func (s *ProductService) ImportCSV(content string) (*ImportResult, error) {
	var result ImportResult
	body := map[string]string{"csv": content}
	err := s.client.post("/products/import", body, &result, nil)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ExportCSV triggers a CSV export.
func (s *ProductService) ExportCSV() (*ExportResult, error) {
	var result ExportResult
	err := s.client.get("/products/export", nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ProductIterator iterates over products.
type ProductIterator struct {
	iter *Iterator[*Product]
}

// Next fetches the next item from the iterator, returning false when done.
func (it *ProductIterator) Next() bool { return it.iter.Next() }

// Product returns the most recently fetched product.
func (it *ProductIterator) Product() *Product { return it.iter.Current() }

// Err returns any error encountered during iteration.
func (it *ProductIterator) Err() error { return it.iter.Err() }

func decodeProduct(raw json.RawMessage) (*Product, error) {
	var p Product
	err := json.Unmarshal(raw, &p)
	return &p, err
}
