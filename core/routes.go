package core

import (
	"github.com/julienschmidt/httprouter"
	"go.uber.org/zap"
)

func Routes(logger *zap.SugaredLogger) *httprouter.Router {
	router := httprouter.New()
	router.GET("/", StatsHandler(logger.Named("Stats")))
	router.POST("/", RunQueryHandler(logger.Named("RunQuery")))
	router.POST("/query/:name", NamedQueryHandler(logger.Named("NamedQuery")))
	return router
}
