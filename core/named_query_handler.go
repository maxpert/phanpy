package core

import (
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"go.uber.org/zap"
	"net/http"
	"os"
)

func NamedQueryHandler(logger *zap.SugaredLogger) httprouter.Handle {
	queries, err := loadQueryConfig(os.Getenv("NAMED_QUERIES"))
	if err != nil {
		queries = make([]namedQueryConfig, 0)
	}

	queryMap := prepareQueryConfigMap(queries)
	logger.Infow("Loaded", "queries", queryMap)
	return func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		queryInfo, ok := queryMap[params.ByName("name")]
		if !ok {
			writer.WriteHeader(404)
			fmt.Fprintf(writer, "Query not found")
			return
		}

		queryParams := make([]interface{}, 0)
		if err := json.NewDecoder(request.Body).Decode(&queryParams); err != nil {
			writer.WriteHeader(403)
			fmt.Fprintf(writer, "Bad parameters")
			return
		}

		r := requestBody{
			Timeout: queryInfo.Timeout,
			Query:   queryInfo.SQL,
			Params:  queryParams,
		}
		streamQueryResponse(request.Context(), writer, r, logger)
	}
}
