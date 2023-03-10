// grupos nesse contexto são conjuntos de funções que lidam com requests
// relacionados
package testgrp

import (
	"context"
	"net/http"

	"github.com/vitoraalmeida/sales-api/foundation/web"
	"go.uber.org/zap"
)

type Handlers struct {
	Log *zap.SugaredLogger
}

// handlers para usar em desenvolvimento, testes, experimentos etc
func (h Handlers) Test(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	status := struct {
		Status string
	}{
		Status: "OK",
	}

	return web.Respond(ctx, w, status, http.StatusOK)
}
