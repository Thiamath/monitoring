/*
 * Copyright (C) 2019 Nalej Group - All Rights Reserved
 */

// Handler for metrics collection

package collect

import (
	"net/http"

	"github.com/nalej/derrors"
)

type Handler struct {
	manager *Manager
}

func NewHandler(manager *Manager) (*Handler, derrors.Error) {
	handler := &Handler{
		manager: manager,
	}

	return handler, nil
}

func (h *Handler) Metrics(w http.ResponseWriter, r *http.Request) {
	h.manager.Metrics(w, r)
}
