package web

import (
	"context"
	"net/http"
	"os"
	"syscall"

	"github.com/dimfeld/httptreemux/v5"
)

// Um Handler no nosso pacote web é uma função que vai receber a requisição
// juntamente com o contexto e chamar internamente a função handle que obedece
// à interface que o pacote http do go impõe
type Handler func(ctx context.Context, w http.ResponseWriter, r *http.Request) error

// embedding permite que o tipo que está usando tenha acesso aos campos, métodos etc
// App pode ser tudo que um httptreemux é, ou seja, é um Http Router compatível
// com o pacote http do Go e pode ser usando como Handler
type App struct {
	*httptreemux.ContextMux //embedding -> o tipo App agora tem embutido um *httptreemux.ContextMux
	shutdown                chan os.Signal
	mw                      []Middleware
}

// A instância de App será compartilhadad (ponteiro) por todo o sistema,
// permitindo que o signal de shutdown possa ser enviado de qualquer ponto
// para qualquer ponto
func NewApp(shutdown chan os.Signal, mw ...Middleware) *App {
	// ...Middleware permite que passemos 0 ou um slice com pelo menos um elemento
	return &App{
		ContextMux: httptreemux.NewContextMux(), // implementa ServerHTTP
		shutdown:   shutdown,
		mw:         mw,
	}
}

func (a *App) SignalShutdown() {
	a.shutdown <- syscall.SIGTERM
}

// group será a versão da api ex.: /v1/...
func (a *App) Handle(method string, group string, path string, handler Handler, mw ...Middleware) {
	// usamos middlewares locais para processos que queremos realizar somente
	// em determinadas rotas
	handler = wrapMiddleware(mw, handler)
	// depois usamos middleware de level de aplicação pra aplicar código que serve
	// de forma geral
	handler = wrapMiddleware(a.mw, handler)

	// a função que nosso mux vai receber tem uma assinatura condizente
	// com que ele espera, que é o padrão de handlers do pacote http
	h := func(w http.ResponseWriter, r *http.Request) {
		// Mas dentro dela nós podemos invocar a função que usamos de fato
		// para o processamento com todo o contexto que queremos

		// PRE CODE PROCESSING
		// logging Stared
		if err := handler(r.Context(), w, r); err != nil {
			// logging error - handle it
			// ERROR HANDLING
			return
		}
		// logging ended
		// POST CODE PROCESSING
	}

	finalPath := path
	if group != "" {
		finalPath = "/" + group + path
	}

	a.ContextMux.Handle(method, finalPath, h)
}
