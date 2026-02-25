package request

import (
	"net/http"
	"strconv"
)

// PagingParams contains common pagination fields for all list requests.
// @name PagingParams
type PagingParams struct {
	PageNumber int `json:"page_number" query:"pageNumber" validate:"required,min=1"`
	PageSize   int `json:"page_size" query:"pageSize" validate:"required,min=1,max=100"`
}

func (p PagingParams) Offset() int {
	return (p.PageNumber - 1) * p.PageSize
}

func defaultPaging() PagingParams {
	return PagingParams{PageNumber: 1, PageSize: 25}
}

func ParsePaging(r *http.Request) PagingParams {
	p := defaultPaging()

	if val, err := strconv.Atoi(r.URL.Query().Get("pageNumber")); err == nil && val > 0 {
		p.PageNumber = val
	}
	if val, err := strconv.Atoi(r.URL.Query().Get("pageSize")); err == nil && val > 0 {
		if val > 100 {
			val = 100
		}
		p.PageSize = val
	}

	return p
}
