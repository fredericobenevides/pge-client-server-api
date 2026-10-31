package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type EconomiaApiResponse struct {
	USDBRL BidResponse `json:"USDBRL"`
}

type BidResponse struct {
	Bid string `json:"bid"`
}

type EconomiaApi struct {
	url string
}

func NewEconomiaApi() *EconomiaApi {
	return &EconomiaApi{
		url: "https://economia.awesomeapi.com.br/json/last/USD-BRL",
	}
}

func (e *EconomiaApi) buscarCotacao(ctx context.Context) (*EconomiaApiResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
	defer cancel()

	log.Printf("Buscando cotação USD-BRL\n")

	req, err := http.NewRequestWithContext(ctx, "GET", e.url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			log.Println("Tempo limite atingido ao buscar a cotação")
		}
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var economiaApiResponse EconomiaApiResponse
	err = json.Unmarshal(body, &economiaApiResponse)
	if err != nil {
		return nil, err
	}

	log.Printf("Cotação USD-BRL obtida com sucesso\n")
	return &economiaApiResponse, nil
}

type ServerApiResponse struct {
	Bid string `json:"bid"`
}

func NewServerApiResponse(economiaApiResponse *EconomiaApiResponse) *ServerApiResponse {
	return &ServerApiResponse{
		Bid: economiaApiResponse.USDBRL.Bid,
	}
}

type ServerApi struct {
	db        *gorm.DB
	tableName string

	economiaApi *EconomiaApi
}

func NewServerApi(db *gorm.DB, tableName string, economiaApi *EconomiaApi) *ServerApi {
	return &ServerApi{db: db, tableName: tableName, economiaApi: economiaApi}
}

func (c *ServerApi) handler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	economiaApiResponse, err := c.economiaApi.buscarCotacao(ctx)
	if err != nil {
		log.Printf("Erro ao buscar a cotação: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	cotacao := NewCotacaoEntity(economiaApiResponse)

	err = c.salvarCotacao(ctx, cotacao)
	if err != nil {
		log.Printf("Erro ao salvar a cotação: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	serverApiResponse := NewServerApiResponse(economiaApiResponse)
	json.NewEncoder(w).Encode(serverApiResponse)
}

func (c *ServerApi) salvarCotacao(ctx context.Context, cotacao *Cotacao) error {
	log.Printf("Salvando a cotação no banco de dados\n")

	ctx, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
	defer cancel()

	err := c.db.WithContext(ctx).Table(c.tableName).Create(cotacao).Error
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			log.Println("Tempo limite atingido ao salvar a cotação")
		}
		return err
	}

	log.Printf("Cotação salva com sucesso\n")
	return nil
}

type Cotacao struct {
	gorm.Model
	Moeda string
	Bid   float64
}

func NewCotacaoEntity(economiaApiResponse *EconomiaApiResponse) *Cotacao {
	bidConv, err := strconv.ParseFloat(economiaApiResponse.USDBRL.Bid, 64)
	if err != nil {
		log.Printf("Erro ao converter o bid: %v\n", err)
		return nil
	}

	return &Cotacao{
		Moeda: "USD-BRL",
		Bid:   bidConv,
	}
}

func main() {
	db, err := gorm.Open(sqlite.Open("cotacao.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Falha ao conectar ao banco de dados:", err)
	}

	tableName := "cotacoes"

	db.Table(tableName).AutoMigrate(&Cotacao{})

	economiaApi := NewEconomiaApi()
	cotacaoApi := NewServerApi(db, tableName, economiaApi)

	http.HandleFunc("/cotacao", cotacaoApi.handler)

	log.Println("Servidor iniciado em :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
