SHELL := /bin/bash

run:
	go run main.go

build:
	#ldflags são flags para interagir com o linker. -X permite acessar variáveis nos programas
	go build -ldflags "-X main.build=local"  # altere a variável build no pacote main para "local"
	# assim conseguimos definir o programa para rodar de formas diferentes a dependenr do ambiente

# modules 
tidy:
	go mod tidy
	go mod vendor # traz o código da dependência para dentro do projeto

VERSION := 1.0

all: sales

sales:
	docker build \
		-f zarf/docker/Dockerfile \
		-t sales-api-amd64:${VERSION} \
		--build-arg BUILD_REF=${VERSION} \
		--build-arg BUILD_DATE=`date -u +"%Y-%m-%dT%H:%M:%SZ"` \
		.

# =========================================================================
# Running from within k8s/kind
KIND_CLUSTER := local-dev-cluster
kind-up:
	kind create cluster \
		--image kindest/node:v1.25.3 \
		--name $(KIND_CLUSTER) \
		--config zarf/k8s/kind/kind-config.yaml
	kubectl config set-context --current --namespace=sales-system

kind-down:
	kind delete cluster --name $(KIND_CLUSTER)

kind-load:
	# disponibiliza a imagem naquele cluster
	kind load docker-image sales-api-amd64:${VERSION} --name ${KIND_CLUSTER}

kind-apply:
	# vai aplicar os patches que vão customizar o deploy dos serviço com base
	# no ambeinte que estamos usando
	kubectl kustomize zarf/k8s/kind/sales-pod | kubectl apply -f -

kind-status:
	kubectl get nodes -o wide
	kubectl get svc -o wide
	kubectl get pods -o wide --watch --all-namespaces

kind-status-service:
	kubectl get pods -o wide --watch

kind-logs:
	kubectl logs -l app=sales --all-containers=true -f --tail=100

kind-describe:
	kubectl describe pod -l app=sales

kind-restart:
	kubectl rollout restart deployment sales-pod

kind-update: all kind-load kind-restart
# vai reiniciar a aplicação no kind quando tiver mudança na config (kustomize)
kind-update-apply: all kind-load kind-apply

