package deployment

import (
	"net/http"

	"github.com/dimasbaguspm/infario/pkgs/request"
	"github.com/dimasbaguspm/infario/pkgs/response"
)

type handler struct {
	service *Service
}

func RegisterRoutes(mux *http.ServeMux, s Service) {
	h := &handler{service: &s}

	mux.HandleFunc("GET /deployments", h.handleGetPagedDeployments)
	mux.HandleFunc("GET /deployments/{id}", h.handleGetDeployment)
	mux.HandleFunc("POST /deployments/upload", h.handleUpload)
}

// handleGetDeployment retrieves a deployment by its ID.
// @Summary      Get a deployment by ID
// @Tags         deployments
// @Produce      json
// @Param id path string true "Deployment ID"
// @Success      200 {object} Deployment
// @Failure      404 {object} response.ErrorResponse "Deployment not found"
// @Failure      500 {object} response.ErrorResponse "Internal Server Error"
// @Router       /deployments/{id} [get]
func (h *handler) handleGetDeployment(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	deployment, err := h.service.GetDeploymentByID(r.Context(), GetSingleDeployment{ID: id})
	if err != nil {
		if fields := response.MapValidationErrors(err); len(fields) > 0 {
			response.Error(w, http.StatusUnprocessableEntity, "Validation failed", fields)
			return
		}
		response.Error(w, http.StatusNotFound, "Deployment not found")
		return
	}

	response.JSON(w, http.StatusOK, deployment)
}

// handleGetPagedDeployments lists all deployments with pagination.
// @Summary      List all deployments
// @Tags         deployments
// @Produce      json
// @Param pageNumber query int false "Page number (default: 1)" default(1)
// @Param pageSize query int false "Page size (default: 25, max: 100)" default(25)
// @Success      200 {object} DeploymentPaged
// @Failure      400 {object} response.ErrorResponse "Invalid parameters"
// @Failure      422 {object} response.ErrorResponse "Validation failed"
// @Failure      500 {object} response.ErrorResponse "Internal Server Error"
// @Router       /deployments [get]
func (h *handler) handleGetPagedDeployments(w http.ResponseWriter, r *http.Request) {
	params := GetPagedDeployment{
		PagingParams: request.ParsePaging(r),
	}

	page, err := h.service.GetPagedDeployments(r.Context(), params)
	if err != nil {
		if fields := response.MapValidationErrors(err); len(fields) > 0 {
			response.Error(w, http.StatusUnprocessableEntity, "Validation failed", fields)
			return
		}
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, page)
}

// handleUpload uploads a deployment artifact (zip or tar.gz) with default 30-day TTL.
// @Summary      Upload a deployment artifact
// @Tags         deployments
// @Accept       mpfd
// @Produce      json
// @Param project_id formData string true "Project ID"
// @Param hash formData string true "Content-addressable hash"
// @Param entry_path formData string true "URL path prefix for Traefik routing"
// @Param file formData file true "Binary file (zip or tar.gz)"
// @Success      201 {object} Deployment
// @Failure      400 {object} response.ErrorResponse "Invalid request"
// @Failure      422 {object} response.ErrorResponse "Validation failed"
// @Failure      500 {object} response.ErrorResponse "Internal Server Error"
// @Router       /deployments/upload [post]
func (h *handler) handleUpload(w http.ResponseWriter, r *http.Request) {
	upload, err := request.ParseFileUpload(r, 10<<20) // 10MB max
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid multipart form or missing file")
		return
	}

	req := UploadDeployment{
		ProjectID:  r.FormValue("project_id"),
		Hash:       r.FormValue("hash"),
		EntryPath:  r.FormValue("entry_path"),
		FileUpload: *upload,
	}

	deployment, err := h.service.Upload(r.Context(), req)
	if err != nil {
		if fields := response.MapValidationErrors(err); len(fields) > 0 {
			response.Error(w, http.StatusUnprocessableEntity, "Validation failed", fields)
			return
		}
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, deployment)
}
