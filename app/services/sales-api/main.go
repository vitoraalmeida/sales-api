package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/ardanlabs/conf"
	"github.com/vitoraalmeida/sales-api/app/services/sales-api/handlers"
	"go.uber.org/automaxprocs/maxprocs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

/*
Usar main.go para definir comentários que servem como to-dos
Usar TODO por todo projeto é fácil de perder
*/

// permite realizar ações no programa de acordo com o ambiente
var build = "delevop"

func main() {
	// constroi o logger da aplicação
	log, err := initLogger("SALES-API")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer log.Sync()

	// inicia a aplicação
	if err := run(log); err != nil {
		// os logs seguem um padrão: Context, Chave e valor
		log.Errorw("startup", "ERROR", err)
		os.Exit(1)
	}

}

func run(log *zap.SugaredLogger) error {
	// ======================================================================
	// GOMAXPROCS

	// define o número correto de threads para o serviço baseado no número
	// que está disponível na máquina ou por quotas (k8s)
	if _, err := maxprocs.Set(); err != nil {
		return fmt.Errorf("maxprocs: %w", err)
	}
	log.Infow("startup", "GOMAXPROCS", runtime.GOMAXPROCS(0))

	// ======================================================================
	// Configuration
	cfg := struct {
		conf.Version
		Web struct {
			// default definem zero-values personalizados para os campos no struct
			APIHost   string `conf:"default:0.0.0.0:3000"`
			DebugHost string `conf:"default:0.0.0.0:4000"`
			// timeouts razoáveis, mas os melhores são definidos com testes
			// de carga, debugging, no uso
			ReadTimeout     time.Duration `conf:"default:5s"`
			WriteTimeout    time.Duration `conf:"default:10s"`
			IdleTimeout     time.Duration `conf:"default:120s"`
			ShutdownTimeout time.Duration `conf:"default:20s"`
		}
	}{
		Version: conf.Version{
			//SystemVersionNumber
			SVN:  build,
			Desc: "Example app",
		},
	}

	// variáveis de ambiente e argumentos de cli com o prefíxo definido serão
	// lidos por conf.ParseOSArgs e popularão o struct
	const prefix = "SALES"
	// parse vai tentar fazer o parsing das opções passadas e gerar uma
	// mensagem de help com base na configuração
	help, err := conf.ParseOSArgs(prefix, &cfg)
	if err != nil {
		// imprime mensagem de texto caso o usuário tenha passado --help
		if errors.Is(err, conf.ErrHelpWanted) {
			fmt.Println(help)
			return nil
		}
		return fmt.Errorf("parsing config: %w", err)
	}

	// ======================================================================
	// App starting
	log.Infow("starting service", "version", build)
	defer log.Infow("shutdown complete")

	out, err := conf.String(&cfg)
	if err != nil {
		return fmt.Errorf("generating config for output: %w", err)
	}
	log.Infow("startup", "config", out)

	// ======================================================================
	// Start debug service
	log.Infow("startup", "status", "debug router started", "host", cfg.Web.DebugHost)

	// A função handlers.DebugStandardLibraryMux retorna um mux que escuta e serve todos os
	// endpoints relacionados a debug. Isso inclui os endpoints da stdlib
	debugMux := handlers.DebugStandardLibraryMux()

	// Inicia o serviço que escuta por requisições de debug
	// Se a goroutine main for finalizada sem que mate as goroutines que foram
	// inicializadas por ela, ficarão orfãns e não serão finalizadas
	go func() {
		// ao inves de passar debugMux, poderiamos passar o http.DefaultServeMux
		// e importar o http/pprof para registrar os endpoints de debug e profiling
		// automaticamente, pois o pacote faz isso na função init().
		// O problema é que qualquer outro pacote que usamos pode incluir
		// no defaultServeMux outros endpoints de debug que retornam informações
		// do pprof sem que saibamos. Então é uma boa prática de segurança criamos um
		// mux personalizado com os endpoints de debug, para que saibamos
		// exatamente o que está sendo entregue de informações de debug
		if err := http.ListenAndServe(cfg.Web.DebugHost, debugMux); err != nil {
			log.Errorw("shutdown", "status", "debug router closed", "host", cfg.Web.DebugHost, "ERROR", err)
		}
	}()

	// ======================================================================
	// Start API service
	shutdown := make(chan os.Signal, 1)
	// SIGINT é enviado por CTRL+C e SIGTERM é o sinal que o K8s envia para
	// finalizar a execução dos serviços
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	// bloqueia a execução (main não termina) enquanto não tiver um dos sinais)

	api := http.Server{
		Addr:         cfg.Web.APIHost,
		Handler:      nil,
		ReadTimeout:  cfg.Web.ReadTimeout,
		WriteTimeout: cfg.Web.WriteTimeout,
		IdleTimeout:  cfg.Web.IdleTimeout,
		ErrorLog:     zap.NewStdLog(log.Desugar()),
	}

	// assim que api.Shutdown é chamado, em caso de signals, api.ListenAndServe
	// retorna, então o retorno entra no canal. Como só tem espaço para uma
	// informação no canal, o envio da mensagem (serverErrors <- api.ListenAndServe)
	// ocorre antes da recepção. Então o case err := <-serverErrors não serã
	// selecionado, pois já estamos no caso seguinte, fazendo com que a goroutine
	// que inicia o recebimento de tráfego também seja finalizada e não fique
	// em execução pra sempre, zumbi
	serverErrors := make(chan error, 1)

	// começa a processar tráfego na porta 3000 (serviço principal)
	go func() {
		log.Infow("startup", "status", "api router started", "host", api.Addr)
		// cada vez que chegar uma request uma goroutine filha desta será iniciada
		// para lidar com a requisição
		// se algum erro na goroutine filha acontecer, será introduzido no canal
		// serverErrors, então a goroutine main saberá
		serverErrors <- api.ListenAndServe()
	}()

	select {
	// se ocorrer um erro que não foi esperado, irrecuperável
	case err := <-serverErrors:
		// pode parar o servidor de vez pq não foi um erro esperado
		return fmt.Errorf("Server error: %w", err)
	// caso seja um sinal esperado
	case sig := <-shutdown:
		log.Infow("shutdown", "status", "shutdown started", "signal", sig)
		defer log.Infow("shutdown", "status", "shutdown complete", "signal", sig)

		// agenda a finalização da aplicação com base no timeout definido para
		// que não esperemos para sempre
		ctx, cancel := context.WithTimeout(context.Background(), cfg.Web.ShutdownTimeout)
		defer cancel()

		// api.Shutdown tenta finalizar o serviço de forma controlada (gracefully)
		// fecha os listeners para não receber mais requests e espera todo o trabalho,
		// que estiver sendo feito ser finalizado. Mas se o tempo que foi estipulado
		// no contexto chegar ao fim, retorna um erro
		if err := api.Shutdown(ctx); err != nil {
			// retornando erro, termina imediatamente o serviço
			api.Close()
			return fmt.Errorf("could not stop the server gracefully", err)
		}

	}

	return nil
}

func initLogger(service string) (*zap.SugaredLogger, error) {
	config := zap.NewProductionConfig()
	config.OutputPaths = []string{"stdout"}
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.DisableStacktrace = true
	config.InitialFields = map[string]interface{}{
		"service": "SALES-API",
	}

	log, err := config.Build()
	if err != nil {
		return nil, err
	}

	return log.Sugar(), nil
}
