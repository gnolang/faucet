package http

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gnolang/faucet/writer"
)

var _ writer.ResponseWriter = (*ResponseWriter)(nil)

type ResponseWriter struct {
	logger *slog.Logger
	w      http.ResponseWriter
}

func New(logger *slog.Logger, w http.ResponseWriter) ResponseWriter {
	return ResponseWriter{
		logger: logger,
		w:      w,
	}
}

func (h ResponseWriter) WriteResponse(response any) {
	if err := json.NewEncoder(h.w).Encode(response); err != nil {
		h.logger.Error(fmt.Sprintf("unable to encode JSON response, %s", err))
	}
}
