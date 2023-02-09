// Package handlers contem um conjunto de funções handler e rotas
// suportadas pela api
package handlers

import (
	"expvar"
	"net/http"
	"net/http/pprof"
	"os"

	"github.com/dimfeld/httptreemux/v5"
	"github.com/vitoraalmeida/sales-api/app/services/sales-api/handlers/debug/checkgrp"
	"github.com/vitoraalmeida/sales-api/app/services/sales-api/handlers/v1/testgrp"
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
func APIMux(cfg APIMuxConfig) *httptreemux.ContextMux {
	// Em APIs, devemos retornar tipos concretos (ou seja, que possuem dados
	// de fato), pois estamos interessados no valor que foi construído. Quem
	// chama a API deve ter o direito de receber o valor concreto e decidir
	// se ele quer tomar alguma ação mais genérica
	mux := httptreemux.NewContextMux() // implementa ServerHTTP

	tgh := testgrp.Handlers{
		Log: cfg.Log,
	}

	mux.Handle(http.MethodGet, "/v1/test", tgh.Test)

	return mux
}
