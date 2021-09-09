package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jackc/pgproto3/v2"
	"github.com/jackc/pgx/v4"
	"github.com/julienschmidt/httprouter"
	"github.com/sony/gobreaker"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type requestBody struct {
	Query   string        `json:"query"`
	Params  []interface{} `json:"params"`
	Timeout int32         `json:"timeout"`
}

var arrayStart = []byte("[")
var arrayEnd = []byte("]")
var arraySeperator = []byte(",")

func rowToMap(description []pgproto3.FieldDescription, row []interface{}) (map[string]interface{}, error) {
	if len(description) != len(row) {
		return nil, errors.New("column size does not match for row mapping")
	}

	ret := make(map[string]interface{})
	for i := 0; i < len(row); i++ {
		ret[string(description[i].Name)] = row[i]
	}

	return ret, nil
}

func streamResponse(w http.ResponseWriter, rows pgx.Rows, logger *zap.SugaredLogger) {
	flusher, _ := w.(http.Flusher)

	w.Header().Set("Mode", "Stream")
	w.WriteHeader(200)

	fields := rows.FieldDescriptions()
	first := true

	w.Write(arrayStart)
	flusher.Flush()
	defer func() {
		w.Write(arrayEnd)
		flusher.Flush()
	}()

	for rows.Next() {
		row, err := rows.Values()
		if err != nil {
			logger.Error("unable to load row", err)
			break
		}

		rowMap, err := rowToMap(fields, row)
		if err != nil {
			logger.Error("unable to map rows", err)
			break
		}

		serRowMap, err := json.Marshal(rowMap)
		if err != nil {
			logger.Error("unable to serialize row", err)
			break
		}

		if first {
			first = false
		} else {
			w.Write(arraySeperator)
		}

		_, err = w.Write(serRowMap)
		if err != nil {
			logger.Warn("connection reset by client, terminating...", err.Error())
			break
		}

		flusher.Flush()
	}
}

func prepareCircuitBreaker(name string) *gobreaker.CircuitBreaker {
	return gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name: name,
		Timeout: 15 * time.Second,
		Interval: 10 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return float32(counts.ConsecutiveFailures) / float32(counts.Requests) > 0.50
		},
		MaxRequests: 10,
	})
}

func RunQueryHandler(logger *zap.SugaredLogger) httprouter.Handle {
	cb := prepareCircuitBreaker("database")

	return func(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
		queryReq := requestBody{
			Timeout: 10,
		}

		if err := json.NewDecoder(req.Body).Decode(&queryReq); err != nil {
			logger.Warn("unable to decode json body")
			w.WriteHeader(403)
			fmt.Fprintf(w, "Bad request")
			return
		}

		ctx, cancel := context.WithTimeout(req.Context(), time.Duration(queryReq.Timeout) * time.Second)
		defer cancel()

		result, err := cb.Execute(func() (interface{}, error) {
			return connectionPool.Query(ctx, queryReq.Query, queryReq.Params...)
		})
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprintf(w, "Error: %v", err)
			return
		}

		rs := result.(pgx.Rows)
		defer rs.Close()
		streamResponse(w, rs, logger)
	}
}
