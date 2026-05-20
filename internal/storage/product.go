package store

import (
	"errors"
	"strconv"
	"sync"

	"github.com/wricheek84/source_asia_backend/internal/model"
)

// ProductStore handles the thread-safe in-memory product data.
type ProductStore struct {
	mu       sync.RWMutex
	products map[string]*model.ProductSummary
	media    map[string][]string // Maps Product ID to its Image URLs
	videos   map[string][]string // Maps Product ID to its Video URLs
	skuIndex map[string]string   // Maps SKU string to Product ID
}

// NewProductStore initializes an empty product catalog store.
func NewProductStore() *ProductStore {
	return &ProductStore{
		products: make(map[string]*model.ProductSummary),
		media:    make(map[string][]string),
		videos:   make(map[string][]string),
		skuIndex: make(map[string]string),
	}
}
// CreateProduct checks for duplicate SKUs, saves the item split across maps, and returns the full layout.
func (s *ProductStore) CreateProduct(input model.CreateProductInput) (*model.ProductDetail, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 1. Check if the barcode SKU is already taken
	if _, exists := s.skuIndex[input.SKU]; exists {
		return nil, errors.New("a product with this SKU already exists")
	}

	// 2. Generate a clean unique ID (e.g., "prod_1", "prod_2")
	id := "prod_" + strconv.Itoa(len(s.products)+1)

	// 3. Assemble the lightweight display card summary
	summary := &model.ProductSummary{
		ID:         id,
		Name:       input.Name,
		SKU:        input.SKU,
		ImageCount: len(input.ImageURLs),
		VideoCount: len(input.VideoURLs),
	}

	// 4. Save to our split-memory architecture
	s.products[id] = summary
	s.skuIndex[input.SKU] = id
	s.media[id] = input.ImageURLs
	s.videos[id] = input.VideoURLs

	// 5. Build and return the full detailed booklet view
	detail := &model.ProductDetail{
		ProductSummary: *summary,
		ImageURLs:      input.ImageURLs,
		VideoURLs:      input.VideoURLs,
	}

	return detail, nil
}
// GetProducts returns a paginated slice of lightweight product summaries and the total catalog count.
func (s *ProductStore) GetProducts(page, limit int) ([]model.ProductSummary, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	total := len(s.products)
	if total == 0 {
		return []model.ProductSummary{}, 0
	}

	// 1. Gather all items from the lightweight map into a clean list
	allProducts := make([]model.ProductSummary, 0, total)
	for _, p := range s.products {
		allProducts = append(allProducts, *p)
	}

	// 2. Calculate the starting and ending index for the requested page
	start := (page - 1) * limit
	if start >= total {
		return []model.ProductSummary{}, total
	}

	end := start + limit
	if end > total {
		end = total
	}

	// 3. Slice out and return only the requested batch
	return allProducts[start:end], total
}
// GetProductByID combines the lightweight summary and heavy media arrays for a single product.
func (s *ProductStore) GetProductByID(id string) (*model.ProductDetail, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 1. Check if the product even exists in our lightweight map
	summary, exists := s.products[id]
	if !exists {
		return nil, false
	}

	// 2. Fetch the heavy website links from our separate media maps
	imageURLs := s.media[id]
	videoURLs := s.videos[id]

	// 3. Glue them together into the full detailed layout
	detail := &model.ProductDetail{
		ProductSummary: *summary,
		ImageURLs:      imageURLs,
		VideoURLs:      videoURLs,
	}

	return detail, true
}
// AddMedia appends new image and video URLs to an existing product and updates its counters.
func (s *ProductStore) AddMedia(id string, input model.AddMediaInput) (*model.ProductDetail, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 1. Verify the product exists before doing anything
	summary, exists := s.products[id]
	if !exists {
		return nil, errors.New("product not found")
	}

	// 2. Append the new web links onto our existing heavy maps
	s.media[id] = append(s.media[id], input.ImageURLs...)
	s.videos[id] = append(s.videos[id], input.VideoURLs...)

	// 3. Keep our lightweight counters perfectly accurate
	summary.ImageCount = len(s.media[id])
	summary.VideoCount = len(s.videos[id])

	// 4. Return the brand-new updated booklet layout
	detail := &model.ProductDetail{
		ProductSummary: *summary,
		ImageURLs:      s.media[id],
		VideoURLs:      s.videos[id],
	}

	return detail, nil
}