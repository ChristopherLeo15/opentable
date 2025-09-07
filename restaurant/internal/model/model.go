package model

type Restaurant struct {
	ID          int    `json:"id"`
	MetadataID  int    `json:"metadata_id"`
	DisplayName string `json:"display_name"`
}