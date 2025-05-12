package handlers

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"code.cloudfoundry.org/brokerapi/v13/domain/apiresponses"
	"code.cloudfoundry.org/brokerapi/v13/middlewares"
)

const getCatalogLogKey = "getCatalog"

func (h APIHandler) Catalog(w http.ResponseWriter, req *http.Request) {
	logger := h.logger.Session(req.Context(), getCatalogLogKey)
	requestId := fmt.Sprintf("%v", req.Context().Value(middlewares.RequestIdentityKey))

	services, err := h.serviceBroker.Services(req.Context())
	if err != nil {
		var apiErr *apiresponses.FailureResponse
		switch {
		case errors.As(err, &apiErr):
			logger.Error(apiErr.LoggerAction(), err)
			h.respond(w, apiErr.ValidatedStatusCode(slog.New(logger)), requestId, apiErr.ErrorResponse())
		default:
			logger.Error(unknownErrorKey, err)
			h.respond(w, http.StatusInternalServerError, requestId, apiresponses.ErrorResponse{
				Description: err.Error(),
			})
		}
		return
	}

	catalog := apiresponses.CatalogResponse{
		Services: services,
	}

	h.respond(w, http.StatusOK, requestId, catalog)
}
