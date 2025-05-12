package handlers

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"code.cloudfoundry.org/brokerapi/v13/domain"
	"code.cloudfoundry.org/brokerapi/v13/domain/apiresponses"
	"code.cloudfoundry.org/brokerapi/v13/internal/blog"
	"code.cloudfoundry.org/brokerapi/v13/middlewares"
)

const lastOperationLogKey = "lastOperation"

func (h APIHandler) LastOperation(w http.ResponseWriter, req *http.Request) {
	instanceID := req.PathValue("instance_id")
	pollDetails := domain.PollDetails{
		PlanID:        req.FormValue("plan_id"),
		ServiceID:     req.FormValue("service_id"),
		OperationData: req.FormValue("operation"),
	}

	logger := h.logger.Session(req.Context(), lastOperationLogKey, blog.InstanceID(instanceID))

	logger.Info("starting-check-for-operation")

	requestId := fmt.Sprintf("%v", req.Context().Value(middlewares.RequestIdentityKey))

	lastOperation, err := h.serviceBroker.LastOperation(req.Context(), instanceID, pollDetails)
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

	logger.Info("done-check-for-operation", slog.Any("state", lastOperation.State))

	lastOperationResponse := apiresponses.LastOperationResponse{
		State:       lastOperation.State,
		Description: lastOperation.Description,
	}

	h.respond(w, http.StatusOK, requestId, lastOperationResponse)
}
