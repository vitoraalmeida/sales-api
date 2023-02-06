# build
FROM golang:1.18 as build_sales-api
# Certifica que o binário gerado não será ligado a nada que esteja na imagem de build
ENV CGO_ENABLED 0
# Define qual tipo de build (versão, local, debug, prod) da aplicação
ARG BUILD_REF

COPY . /service

WORKDIR /service
# faz o build da aplicação conforme o ambiente que se quer
#ldflags são flags para interagir com o linker. -X permite acessar variáveis nos programas
RUN go build -ldflags "-X main.build=${BUILD_REF}"

# Imagem contendo apenas a aplicacão - reduz o tamanho

FROM alpine:3.14
ARG BUILD_DATE
ARG BUILD_REF
COPY --from=build_sales-api /service /service
WORKDIR /service
CMD ["./sales-api"]

LABEL org.opencontainers.image.created="{BUILD_DATE}" \
      org.opencontainers.image.title="sales-api" \
      org.opencontainers.image.authors="Vitor Almeida" \
      org.opencontainers.image.source="https://github.com/vitoraalmeida/sales-api/app/sales-api" \
      org.opencontainers.image.revision="${BUILD_REF}"

        
