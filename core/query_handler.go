package core

import (
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"go.uber.org/zap"
	"net/http"
)

type requestBody struct {
	Query   string        `json:"query"`
	Params  []interface{} `json:"params"`
	Timeout int32         `json:"timeout"`
}

func RunQueryHandler(logger *zap.SugaredLogger) httprouter.Handle {
	return func(writer http.ResponseWriter, req *http.Request, params httprouter.Params) {
		queryReq := requestBody{
			Timeout: 10,
		}

		if err := json.NewDecoder(req.Body).Decode(&queryReq); err != nil {
			logger.Warn("unable to decode json body")
			writer.WriteHeader(403)
			fmt.Fprintf(writer, "Bad request")
			return
		}

		streamQueryResponse(req.Context(), writer, queryReq, logger)
	}
}
