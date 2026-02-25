package project

import (
	"encoding/json"
	"net/http"

	"github.com/dimasbaguspm/infario/pkgs/response"
)

type handler struct {
	service *Service
}

func RegisterRoutes(mux *http.ServeMux, s Service) {
	h := &handler{service: &s}

	mux.HandleFunc("GET /projects/{id}", h.handleProjectByID)
	mux.HandleFunc("POST /projects", h.handleCreateProject)
	mux.HandleFunc("PATCH /projects/{id}", h.handleUpdateProject)
	mux.HandleFunc("DELETE /projects/{id}", h.handleDeleteProject)
}

// handleCreateProject registers a new project.
// @Summary      Create a project
// @Tags         projects
// @Accept       json
// @Produce      json
// @Param request body CreateProject true "Project Details"
// @Success      201 {object} Project
// @Failure      400 {object} response.ErrorResponse "Bad Request"
// @Failure      500 {object} response.ErrorResponse "Internal Server Error"
// @Router       /projects [post]
func (h *handler) handleCreateProject(w http.ResponseWriter, r *http.Request) {
	var req CreateProject
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	project, err := h.service.CreateNewProject(r.Context(), req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, project)
}

// handleProjectByID retrieves a project by its ID.
// @Summary      Get a project by ID
// @Tags         projects
// @Produce      json
// @Param id path string true "Project ID"
// @Success      200 {object} Project
// @Failure      404 {object} response.ErrorResponse "Project not found"
// @Failure      500 {object} response.ErrorResponse "Internal Server Error"
// @Router       /projects/{id} [get]
func (h *handler) handleProjectByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	project, err := h.service.GetProjectByID(r.Context(), GetSingleProject{ID: id})
	if err != nil {
		response.Error(w, http.StatusNotFound, "Project not found")
		return
	}

	response.JSON(w, http.StatusOK, project)
}

// handleUpdateProject updates an existing project.
// @Summary      Update a project
// @Tags         projects
// @Accept       json
// @Produce      json
// @Param id path string true "Project ID"
// @Param request body UpdateProject true "Updated Project Details"
// @Success      200 {object} Project
// @Failure      400 {object} response.ErrorResponse "Invalid request body"
// @Failure      404 {object} response.ErrorResponse "Project not found"
// @Failure      500 {object} response.ErrorResponse "Internal Server Error"
// @Router       /projects/{id} [patch]
func (h *handler) handleUpdateProject(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var req UpdateProject
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	req.ID = id

	project, err := h.service.UpdateProject(r.Context(), req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, project)
}

// handleDeleteProject deletes a project by its ID.
// @Summary      Delete a project
// @Tags         projects
// @Param id path string true "Project ID"
// @Success      204 "No Content"
// @Failure      404 {object} response.ErrorResponse "Project not found"
// @Failure      500 {object} response.ErrorResponse "Internal Server Error"
// @Router       /projects/{id} [delete]
func (h *handler) handleDeleteProject(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	err := h.service.DeleteProject(r.Context(), DeleteProject{ID: id})
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusNoContent, nil)
}
