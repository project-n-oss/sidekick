package api

import (
	"net/http"

	"go.uber.org/zap"
)

func (a *Api) InternalError(logger *zap.Logger, w http.ResponseWriter, err error) {
	logger.Error("internal error", zap.Error(err))
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (a *Api) BadRequest(logger *zap.Logger, w http.ResponseWriter, err error) {
	logger.Error("bad request", zap.Error(err))
	http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
}
