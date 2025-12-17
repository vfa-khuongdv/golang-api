package dto

type Pagination[T any] struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	TotalItems int `json:"total_items"`
	TotalPages int `json:"total_pages"`
	Data       []T `json:"data"`
}
