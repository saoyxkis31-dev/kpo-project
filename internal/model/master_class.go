package model

type MasterClass struct {
	ID          int    `json:"id"`
	ProductID   int    `json:"product_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	VideoURL    string `json:"video_url"`
}
