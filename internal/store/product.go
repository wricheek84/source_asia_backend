package store

import (
	"errors"
	"strconv"
	"sync"

	"github.com/wricheek84/source_asia_backend/internal/model"
)


type ProductStore struct {
	mu       sync.RWMutex
	products map[string]*model.ProductSummary
	media    map[string][]string 
	videos   map[string][]string 
	skuIndex map[string]string  
}


func NewProductStore() *ProductStore {
	return &ProductStore{
		products: make(map[string]*model.ProductSummary),
		media:    make(map[string][]string),
		videos:   make(map[string][]string),
		skuIndex: make(map[string]string),
	}
}

func (s *ProductStore) CreateProduct(input model.CreateProductInput) (*model.ProductDetail, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	
	if _, exists := s.skuIndex[input.SKU]; exists {
		return nil, errors.New("a product with this SKU already exists")
	}
	id := "prod_" + strconv.Itoa(len(s.products)+1)

	
	summary := &model.ProductSummary{
		ID:         id,
		Name:       input.Name,
		SKU:        input.SKU,
		ImageCount: len(input.ImageURLs),
		VideoCount: len(input.VideoURLs),
	}

	
	s.products[id] = summary
	s.skuIndex[input.SKU] = id
	s.media[id] = input.ImageURLs
	s.videos[id] = input.VideoURLs

	
	detail := &model.ProductDetail{
		ProductSummary: *summary,
		ImageURLs:      input.ImageURLs,
		VideoURLs:      input.VideoURLs,
	}

	return detail, nil
}

func (s *ProductStore) GetProducts(page, limit int) ([]model.ProductSummary, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	total := len(s.products)
	if total == 0 {
		return []model.ProductSummary{}, 0
	}

	
	allProducts := make([]model.ProductSummary, 0, total)
	for _, p := range s.products {
		allProducts = append(allProducts, *p)
	}

	
	start := (page - 1) * limit
	if start >= total {
		return []model.ProductSummary{}, total
	}

	end := start + limit
	if end > total {
		end = total
	}

	return allProducts[start:end], total
}

func (s *ProductStore) GetProductByID(id string) (*model.ProductDetail, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	
	summary, exists := s.products[id]
	if !exists {
		return nil, false
	}

	imageURLs := s.media[id]
	videoURLs := s.videos[id]

	
	detail := &model.ProductDetail{
		ProductSummary: *summary,
		ImageURLs:      imageURLs,
		VideoURLs:      videoURLs,
	}

	return detail, true
}

func (s *ProductStore) AddMedia(id string, input model.AddMediaInput) (*model.ProductDetail, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	summary, exists := s.products[id]
	if !exists {
		return nil, errors.New("product not found")
	}

	
	s.media[id] = append(s.media[id], input.ImageURLs...)
	s.videos[id] = append(s.videos[id], input.VideoURLs...)

	
	summary.ImageCount = len(s.media[id])
	summary.VideoCount = len(s.videos[id])

	
	detail := &model.ProductDetail{
		ProductSummary: *summary,
		ImageURLs:      s.media[id],
		VideoURLs:      s.videos[id],
	}

	return detail, nil
}