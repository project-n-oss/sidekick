package api

import (
	"net/http"

	"go.uber.org/zap"
)

func (a *Api) InternalError(w http.ResponseWriter, err error) {
	a.logger.Error("internal error", zap.Error(err))
	w.WriteHeader(500)
}
