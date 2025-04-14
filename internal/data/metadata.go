package data

import (
	"math"

	"github.com/kharljhon14/starbloom-server/internal/validator"
)

type Metadata struct {
	CurrentPage  int `json:"current_page,omitempty"`
	PageSize     int `json:"page_size,omitempty"`
	FirstPage    int `json:"first_page,omitempty"`
	LastPage     int `json:"last_page,omitempty"`
	TotalRecords int `json:"total_records,omitempty"`
}

func calculateMetadata(totalRecords, page, pageSize int) Metadata {
	if totalRecords == 0 {
		return Metadata{}
	}

	return Metadata{
		CurrentPage:  page,
		PageSize:     pageSize,
		FirstPage:    1,
		LastPage:     int(math.Ceil(float64(totalRecords) / float64(pageSize))),
		TotalRecords: totalRecords,
	}
}

type Filter struct {
	Page     int
	PageSize int
	Sort     string
}

func (f Filter) limit() int {
	return f.PageSize
}

func (f Filter) offset() int {
	return (f.Page - 1) * f.PageSize
}

func (f Filter) sort() string {
	return f.Sort
}

func ValidateFilters(v *validator.Validator, f Filter) {
	v.Check(f.Page > 0, "page", "page must be greater than zero")
	v.Check(f.Page <= 10_000_000, "page", "page must be  a maximum of 10 million")
	v.Check(f.PageSize > 0, "page_size", "page_size must be greater than 0")
	v.Check(f.PageSize <= 1000, "page_size", "page_size must be  a maximum of 1000")
}
