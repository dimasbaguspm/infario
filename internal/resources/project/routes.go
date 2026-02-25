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

func (h *handler) handleProjectByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	project, err := h.service.GetProjectByID(r.Context(), GetSingleProject{ID: id})
	if err != nil {
		response.Error(w, http.StatusNotFound, "Project not found")
		return
	}

	response.JSON(w, http.StatusOK, project)
}

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

func (h *handler) handleDeleteProject(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	err := h.service.DeleteProject(r.Context(), DeleteProject{ID: id})
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusNoContent, nil)
}
