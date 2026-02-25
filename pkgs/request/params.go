package request

import (
	"mime/multipart"
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

// FileUpload represents parsed multipart file upload data.
// @name FileUpload
type FileUpload struct {
	File *multipart.FileHeader `json:"-" validate:"required"` // Binary file (zip or tar.gz)
}

// ParseFileUpload parses multipart form data and extracts file upload fields.
// MaxSize specifies the maximum file size in bytes (default: 10MB).
// Returns FileUpload struct with projectID, hash, and file header.
func ParseFileUpload(r *http.Request, maxSize int64) (*FileUpload, error) {
	if maxSize <= 0 {
		maxSize = 10 << 20 // 10MB default
	}

	if err := r.ParseMultipartForm(maxSize); err != nil {
		return nil, err
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fileHeader := r.MultipartForm.File["file"][0]

	return &FileUpload{

		File: fileHeader,
	}, nil
}
