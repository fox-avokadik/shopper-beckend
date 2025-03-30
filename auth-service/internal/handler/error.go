package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/status"

	"auth-service/internal/models"
)

func CustomErrorHandler(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, r *http.Request, err error) {
	w.Header().Set("Content-Type", "application/json")

	st := status.Convert(err)
	code := runtime.HTTPStatusFromCode(st.Code())
	w.WriteHeader(code)

	response := struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}{
		Code:    getMessageCode(st.Message()),
		Message: st.Message(),
	}

	json.NewEncoder(w).Encode(response)
}

func getMessageCode(msg string) string {
	if err, exists := models.MessageToError[msg]; exists {
		return err.Code
	}
	return models.ErrInternal.Code
}
