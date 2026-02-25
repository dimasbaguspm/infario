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

// handleGetPagedDeployments lists deployments for a project with pagination.
// @Summary      List deployments for a project
// @Tags         deployments
// @Produce      json
// @Param projectID path string true "Project ID"
// @Param pageNumber query int false "Page number (default: 1)" default(1)
// @Param pageSize query int false "Page size (default: 10, max: 100)" default(10)
// @Success      200 {object} DeploymentPaged
// @Failure      400 {object} response.ErrorResponse "Invalid parameters"
// @Failure      422 {object} response.ErrorResponse "Validation failed"
// @Failure      500 {object} response.ErrorResponse "Internal Server Error"
// @Router       /deployments [get]
func (h *handler) handleGetPagedDeployments(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("projectID")

	params := GetPagedDeployment{
		ProjectID:    projectID,
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
