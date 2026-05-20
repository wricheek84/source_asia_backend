package validation

import (
	"errors"
	"net/url"
	"strings"

	"github.com/wricheek84/source_asia_backend/internal/model"
)

// ValidateProduct checks constraints for creating a product.
func ValidateProduct(input model.CreateProductInput) error {
	if strings.TrimSpace(input.Name) == "" {
		return errors.New("product name cannot be empty")
	}
	if strings.TrimSpace(input.SKU) == "" {
		return errors.New("sku cannot be empty")
	}
	if len(input.ImageURLs) > 20 {
		return errors.New("image urls cannot exceed 20 items")
	}
	if len(input.VideoURLs) > 20 {
		return errors.New("video urls cannot exceed 20 items")
	}
	for _, u := range input.ImageURLs {
		if !IsValidURL(u) {
			return errors.New("invalid image url format: " + u)
		}
	}
	for _, u := range input.VideoURLs {
		if !IsValidURL(u) {
			return errors.New("invalid video url format: " + u)
		}
	}
	return nil
}

// ValidateMedia checks constraints for adding media.
func ValidateMedia(input model.AddMediaInput) error {
	if len(input.ImageURLs) == 0 && len(input.VideoURLs) == 0 {
		return errors.New("at least one image or video url must be provided")
	}
	if len(input.ImageURLs) > 20 || len(input.VideoURLs) > 20 {
		return errors.New("media arrays cannot exceed 20 items")
	}
	for _, u := range input.ImageURLs {
		if !IsValidURL(u) {
			return errors.New("invalid image url format: " + u)
		}
	}
	for _, u := range input.VideoURLs {
		if !IsValidURL(u) {
			return errors.New("invalid video url format: " + u)
		}
	}
	return nil
}

// IsValidURL verifies if a string is a properly formatted web link.
func IsValidURL(toCheck string) bool {
	u, err := url.ParseRequestURI(toCheck)
	if err != nil {
		return false
	}
	return u.Scheme == "http" || u.Scheme == "https"
}