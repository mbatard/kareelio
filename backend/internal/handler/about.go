package handler

import (
	"net/http"

	"github.com/user/kareelio/backend/internal/model"
)

type AboutHandler struct{}

func NewAboutHandler() *AboutHandler {
	return &AboutHandler{}
}

func (h *AboutHandler) Get(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, model.GetAbout())
}
