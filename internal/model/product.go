package model


type ProductSummary struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	SKU        string `json:"sku"`
	ImageCount int    `json:"image_count"`
	VideoCount int    `json:"video_count"`
}


type ProductDetail struct {
	ProductSummary
	ImageURLs []string `json:"image_urls"`
	VideoURLs []string `json:"video_urls"`
}

type CreateProductInput struct {
	Name      string   `json:"name"`
	SKU       string   `json:"sku"`
	ImageURLs []string `json:"image_urls"`
	VideoURLs []string `json:"video_urls"`
}


type AddMediaInput struct {
	ImageURLs []string `json:"image_urls"`
	VideoURLs []string `json:"video_urls"`
}