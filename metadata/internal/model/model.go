package model

import "fmt"

type Metadata struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	CuisineType string `json:"cuisine_type"`
	PriceRange  string `json:"price_range"`
	Address     string `json:"address"`
	City        string `json:"city"`
}

func (m Metadata) Validate() error {
	if m.Name == "" {
		return fmt.Errorf("name is required")
	}
	if m.Address == "" {
		return fmt.Errorf("address is required")
	}
	return nil
}