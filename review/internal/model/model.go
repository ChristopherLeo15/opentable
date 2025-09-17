package model

import "fmt"

type Review struct {
	ID           int    `json:"id"`
	RestaurantID int    `json:"restaurant_id"`
	Rating       int    `json:"rating"`
	Comment      string `json:"comment"`
}

func (r Review) Validate() error {
	if r.RestaurantID <= 0 {
		return fmt.Errorf("restaurant_id must be positive")
	}
	if r.Rating < 1 || r.Rating > 5 {
		return fmt.Errorf("rating must be between 1 and 5")
	}
	return nil
}