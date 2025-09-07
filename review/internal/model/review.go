package model

type Review struct {
	ID           int    `json:"id"`
	RestaurantID int    `json:"restaurant_id"`
	Rating       int    `json:"rating"`
	Comment      string `json:"comment"`
}