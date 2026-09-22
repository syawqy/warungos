package model

import "time"

// Pagination holds standard pagination parameters.
type Pagination struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
	Total    int `json:"total"`
}

// Offset returns the database offset for the current page.
func (p *Pagination) Offset() int {
	if p.Page < 1 {
		p.Page = 1
	}
	return (p.Page - 1) * p.PageSize
}

// APIResponse is the standard JSON envelope for all API responses.
type APIResponse struct {
	Success    bool        `json:"success"`
	Data       interface{} `json:"data,omitempty"`
	Error      string      `json:"error,omitempty"`
	Pagination *Pagination `json:"pagination,omitempty"`
	Meta       interface{} `json:"meta,omitempty"`
	Timestamp  time.Time   `json:"timestamp"`
}

// SuccessResponse creates a success APIResponse.
func SuccessResponse(data interface{}) APIResponse {
	return APIResponse{
		Success:   true,
		Data:      data,
		Timestamp: time.Now(),
	}
}

// SuccessResponseWithPagination creates a success APIResponse with pagination.
func SuccessResponseWithPagination(data interface{}, p *Pagination) APIResponse {
	return APIResponse{
		Success:    true,
		Data:       data,
		Pagination: p,
		Timestamp:  time.Now(),
	}
}

// ErrorResponse creates an error APIResponse.
func ErrorResponse(err string) APIResponse {
	return APIResponse{
		Success:   false,
		Error:     err,
		Timestamp: time.Now(),
	}
}
