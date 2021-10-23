package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jackc/pgproto3/v2"
	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"
	"net/http"
	"os"
	"strconv"
	"time"
)

var arrayStart = []byte("[")
var arrayEnd = []byte("]")
var arraySeparator = []byte(",")

var flushRowBatch = readOsEnvInt("FLUSH_ROW_BATCH", 100)

func readOsEnvInt(name string, fallback int) int {
	v := os.Getenv(name)
	if v == "" {
		return fallback
	}

	ret, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}

	return ret
}

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
	rowsCount := 0

	_, _ = w.Write(arrayStart)
	defer func() {
		_, _ = w.Write(arrayEnd)
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

		if rowsCount != 0 {
			_, _ = w.Write(arraySeparator)
		}

		_, err = w.Write(serRowMap)
		if err != nil {
			logger.Warn("connection reset by client, terminating...", err.Error())
			break
		}

		rowsCount++
		if int(rowsCount % flushRowBatch) == 0 {
			flusher.Flush()
		}
	}
}

func streamQueryResponse(
	ctx context.Context,
	w http.ResponseWriter,
	queryReq requestBody,
	logger *zap.SugaredLogger,
) {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(queryReq.Timeout)*time.Second)
	defer cancel()

	rs, err := connectionPool.Query(ctx, queryReq.Query, queryReq.Params...)
	if err != nil {
		w.WriteHeader(500)
		_, _ = fmt.Fprintf(w, "Error: %v", err)
		return
	}

	defer rs.Close()
	streamResponse(w, rs, logger)
}
