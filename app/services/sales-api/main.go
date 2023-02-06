package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"go.uber.org/automaxprocs/maxprocs"
)

// permite realizar ações no programa de acordo com o ambiente
var build = "delevop"

func main() {
	// define o número máximo que threads para o serviço
	// baseado no que está disponível pela máquina ou quotas (k8s)
	if _, err := maxprocs.Set(); err != nil {
		fmt.Println("maxprocs: %w", err)
		os.Exit(1)
	}

	// O número de CPUs disponíveis - o número de goroutines que podem rodar em paralelo
	g := runtime.GOMAXPROCS(0)
	log.Printf("starting sales build[%s] CPU[%d]", build, g)
	defer log.Println("sales ended")

	shutdown := make(chan os.Signal, 1)
	// SIGINT é enviado por CTRL+C e SIGTERM é o sinal que o K8s envia para
	// finalizar a execução dos serviços
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	// bloqueia a execução do main enquanto não vier um sinal na channel
	<-shutdown

	log.Println("stopping service")
}
