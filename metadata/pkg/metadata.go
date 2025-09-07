package metadata

type Metadata struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Location    string `json:"location"`
	CuisineType string `json:"cuisine_type"`
	PriceRange  string `json:"price_range"`
}