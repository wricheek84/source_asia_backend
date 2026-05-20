package model

// ProductSummary represents the lightweight entity used for paginated listing.
type ProductSummary struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	SKU        string `json:"sku"`
	ImageCount int    `json:"image_count"`
	VideoCount int    `json:"video_count"`
}

// ProductDetail represents the full layout returned on a specific detail view.
type ProductDetail struct {
	ProductSummary
	ImageURLs []string `json:"image_urls"`
	VideoURLs []string `json:"video_urls"`
}

// CreateProductInput defines the fields we accept when creating a product.
type CreateProductInput struct {
	Name      string   `json:"name"`
	SKU       string   `json:"sku"`
	ImageURLs []string `json:"image_urls"`
	VideoURLs []string `json:"video_urls"`
}

// AddMediaInput defines the structure for appending media links.
type AddMediaInput struct {
	ImageURLs []string `json:"image_urls"`
	VideoURLs []string `json:"video_urls"`
}