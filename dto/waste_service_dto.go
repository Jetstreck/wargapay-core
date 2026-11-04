package dto

type WasteServiceRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       int64  `json:"price"`
	IsActive    *bool  `json:"is_active"`
}
