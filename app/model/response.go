package model

type WebResponse struct {
	Success bool                `json:"success"`
	Message string              `json:"message"`
	Data    any                 `json:"data,omitempty"`
	Meta    *Meta               `json:"meta,omitempty"`
	Errors  map[string][]string `json:"errors,omitempty"`
}

type Meta struct {
	CurrentPage int `json:"current_page"`
	PerPage     int `json:"per_page"`
	Total       int `json:"total"`
	LastPage    int `json:"last_page"`
}
