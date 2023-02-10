// Package handlers contem um conjunto de funções handler e rotas
// suportadas pela api
package handlers

import (
	"expvar"
	"net/http"
	"net/http/pprof"
	"os"

	"github.com/vitoraalmeida/sales-api/app/services/sales-api/handlers/debug/checkgrp"
	"github.com/vitoraalmeida/sales-api/app/services/sales-api/handlers/v1/testgrp"
	"github.com/vitoraalmeida/sales-api/business/web/mid"
	"github.com/vitoraalmeida/sales-api/foundation/web"
	"go.uber.org/zap"
)

// Envoltório para rotas de debug com informações fornecidas pelas stdlib
func DebugStandardLibraryMux() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	mux.Handle("/debug/vars", expvar.Handler())

	return mux
}

// DebugMux registra rotas para debug
// Envoltório com as rotas de infos da stdlib mais infos sobre readiness e liveness
func DebugMux(build string, log *zap.SugaredLogger) http.Handler {
	mux := DebugStandardLibraryMux()

	// grupo de handlers referentes a checks (liveness, readiness)
	chg := checkgrp.Handlers{
		Build: build,
		Log:   log,
	}

	mux.HandleFunc("/debug/readiness", chg.Readiness)
	mux.HandleFunc("/debug/liveness", chg.Liveness)

	return mux
}

type APIMuxConfig struct {
	Shutdown chan os.Signal
	Log      *zap.SugaredLogger
}

// APIMux registra rotas relativas à API em sí
// router
func APIMux(cfg APIMuxConfig) *web.App {
	//constrói a web.App que vai conter todas as rotas
	app := web.NewApp(
		cfg.Shutdown,
		// quando desenvolvemos handlers, não precisamos mais loggar, pois o middleware já faz
		mid.Logger(cfg.Log),
	)

	// atrela as rotas da v1 na app
	v1(app, cfg)

	return app
}

// v1 agrupa as rotas relacionadas a versão 1. Registra os handlers
// que são relativos à sua versão
func v1(app *web.App, cfg APIMuxConfig) {
	const version = "v1"

	tgh := testgrp.Handlers{
		Log: cfg.Log,
	}

	app.Handle(http.MethodGet, version, "/test", tgh.Test)
}
