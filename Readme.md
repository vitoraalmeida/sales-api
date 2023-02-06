Não há consenso sobre qual é a forma mais apropriada para organizar um projeto
em Go.

Este projeto utiliza layers.

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
