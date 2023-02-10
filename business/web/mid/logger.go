package mid

import (
	"context"
	"net/http"
	"time"

	"github.com/vitoraalmeida/sales-api/foundation/web"
	"go.uber.org/zap"
)

func Logger(log *zap.SugaredLogger) web.Middleware {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			traceID := "00000000000000000000"
			statusCode := http.StatusOK
			now := time.Now()

			log.Infow(
				"request started",
				"traceid", traceID,
				"method", r.Method,
				"path", r.URL.Path,
				"remoteaddr", r.RemoteAddr,
			)

			err := handler(ctx, w, r)

			log.Infow(
				"request ended",
				"traceid", traceID,
				"method", r.Method,
				"path", r.URL.Path,
				"remoteaddr", r.RemoteAddr,
				"statuscode", statusCode,
				"since", time.Since(now),
			)

			return err

		}
		return h

	}
	return m
}
