# PGE Client-Server API

Projeto implementado com base nos requisitos da atividade "Desafio Client-Server-API" no curso Pós Go Expert da FullCycle.

O servidor é responsável por buscar a cotação do do dólar (USD-BRL) a partir da [AwesomeAPI](https://economia.awesomeapi.com.br/), persistir a cotação em um banco SQLite com [GORM](https://gorm.io/) e expor essa cotação por meio de uma API HTTP.

Já o cliente consome essa API e salva o valor da cotação em um arquivo `.txt`.

## Arquitetura

O projeto é dividido em duas aplicações independentes:

- **`server/`** - Servidor HTTP responsável por buscar, persistir e disponibilizar a cotação USD-BRL.
- **`client/`** - Cliente HTTP responsável por requisitar a cotação ao servidor e salvar em arquivo.

## Estrutura do Projeto

```text
.
├── client/
│   ├── cotacao.txt  # Arquivo gerado pelo cliente com a cotação
│   └── main.go      # Lógica do cliente
├── server/
│   ├── cotacao.db   # Banco SQLite utilizado pelo servidor
│   └── main.go      # Lógica do servidor
├── test/
│   └── cotacao.http # Requisição HTTP para testes
├── go.mod
├── go.sum
└── README.md
```

## Como Executar

É necessário executar o servidor e o cliente em terminais separados.

### 1) Iniciar o servidor

```bash
cd server
go run main.go
```

O servidor ficará disponível em: `http://localhost:8080`

### 2) Executar o cliente

Em outro terminal, a partir da raiz do projeto:

```bash
cd client
go run main.go
```

Após a execução, o arquivo `client/cotacao.txt` será criado ou atualizado com o seguinte formato:

```text
Dólar: 5.3716
```

## Endpoint Disponível

| Método | Rota       | Descrição                                                                                       |
| ------ | ---------- | ----------------------------------------------------------------------------------------------- |
| `GET`  | `/cotacao` | Busca a cotação USD-BRL na AwesomeAPI, salva no banco de dados e retorna o valor `bid` em JSON. |

### Exemplo de resposta

```json
{
  "bid": "5.3716"
}
```
