## Dependências do projeto:
* Go >= 1.18
* Make
* Docker
* Kind (Kubernetes in Docker)
* Kubectl
* kustomize (ferramenta para alterar arquivos yaml de recursos do k8s com base em determinados parâmetros

### utilitários
* expvarmon -> vizualização das infos disponibilizadas pelo expvar da stdlib

## Sobre a estrutura da aplicação

Não há consenso sobre qual é a forma mais apropriada para organizar um projeto
em Go.

Este projeto utiliza camadas.

```
app
    services
        metrics     -> sidecar para o sales-api
            pacotes específicos
        sales-api
            pacotes específicos
    tooling
        logfmt
        sales-admin
business
    core -> nível de entrada para a camada de negócio
    data
        schema
        store
            product
            user
        tests
    sys
        auth
        database
        metrics
        validade
    web
foundation
    docker
    keystore
    logger
    web
    worker
vendos
zarf
```
camadas posicionadas acima, importam camadas posicionados abaixo

####App layer: 

inicia e desliga os serviços/tooling
aceita input externo e provê output externo
Os códigos não são reutilizáveis e não possuem imports entre pacotes que estão em App
inicia -> recebe input -> valida input -> processa input -> retorna output -> desliga

#### Business
A lógica de negócio que precisamos para resolver o problema de negócio que temos

App pode importar business

Core -> provê a API para acessar a camada de negócios. Usa a camada de dados
Data -> camada de dados, fornece acesso aos bancos de dados, a camada de CRUD
camada de app pode acessar direto a camada de data se apenas precisa de CRUD,
não é necessário criar abstrações que não agregam valor em core apenas para 
fazer CRUD. Mas coisas que precisam fazer múltiplas chamadas para a camada de 
dados, chamadas mais complexas, devem passar por core. Nela vão os modelos,
códigos de acesso ao banco etc. Não devem importar uns aos outros
System -> Pacotes que são específicos ao problema que estamos resolvendo, mas
ao mesmo tempo são orientados ao sistema (autenticação, banco de dados, metricas, validação)
Web -> específicos para aplicações web

#### Foundation
App usa, business usa.

Sobre pacotes fundamentais que não estão necessariamente ligados ao problema de negócios.
poderiam até ser usados por outros problemas de negócios, mas não estão prontos
ou não justificam tem seus próprios repositórios

Como se fosse a stdlib do projeto. (não fornecem logs, não agrupa erros)
Não importam uns aos outros

#### Vendor
Código de terceiros que estamos usando
Permite que possamos saber como estão definidos, go to definition, documentação etc

#### ZARF

Configuração para rodar o projeto


## Anotações durante o curso

### Logs
Logs devem ser legíveis e também é interessante ter logs estruturados caso
seja desejável colocar os logs em outros lugares

Códigos fundamentais (foundation) não devem gerar logs, business and application
devem gerar logs.

Logs são importantes para resolver problemas. Então devemos poder passar o
logger pela aplicação onde precisarmos. Loggers devem estar explícitos e não 
escondidos em context. Logs devem sinalizar, mas não devem gerar ruídos!

Existem bons pacotes de logs, até mesmo na stdlib, mas o pacote de logs do 
uber provê logs estruturados com a habilidade de poder fazê-los legíveis

"go.uber.org/zap"
"go.uber.org/zap/zapcore"

### Configs

Apenas main.go deve ter acesso às configurações durante o startup, nenhum outro
pacote e em nenhum outro momento

Menos é mais

Devem existir bons defaults, passiveis de executar a aplicação de cara.

Precisam ser passíveis de modificação pelo menos variáveis de ambiente ou pela
CLI

### Metricas e debug

A stdlib provê pacotes para que possamos ver informções do estado interno da
aplicação (alocações, uso de memória, GC etc)

### Shutdown e agendamento do desligamento

É importante que o servidor pare sozinho se algo inesperado acontecer ou mesmo
se a aplicação for terminada deliberadamente (seja por um comando no terminal
CTRL + C, ou um sinal do kubernetes para interromper o container SIGTERM)

Se for um sinal deliberado, é importante que agendemos a finalização para 
garantir que todo o trabalho seja terminado antes de fechar o servidor. Também
é importante garantirmos que todas as goroutines relacionadas serão finalizadas
para que não fiquem em execução zumbi (quando a goroutine main é finalizada, 
mas as goroutines filhas não), pois consome recursos


### Readiness e liveness

Não são orientados ao negócio, mas são orientados a produção

Readiness -> Podemos indicar que estamos prontos para receber tráfego
Liveness -> Estamos executando 


