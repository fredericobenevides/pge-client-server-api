package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type Cotacao struct {
	Bid string `json:"bid"`
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	url := "http://localhost:8080/cotacao"

	cotacao, err := carregarCotacao(ctx, url)
	if err != nil {
		panic(err)
	}

	err = salvarCotacaoEmArquivo(cotacao, "cotacao.txt")
	if err != nil {
		panic(err)
	}
}

func carregarCotacao(ctx context.Context, url string) (*Cotacao, error) {
	log.Println("Buscando cotação USD-BRL no servidor localizado em:", url)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
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

	log.Println("Resposta recebida do servidor:", string(body))

	var cotacao Cotacao

	err = json.Unmarshal(body, &cotacao)
	if err != nil {
		return nil, err
	}

	log.Println("Cotação recebida do servidor:", cotacao.Bid)

	return &cotacao, nil
}

func salvarCotacaoEmArquivo(cotacao *Cotacao, caminho string) error {
	log.Println("Salvando cotação no arquivo localizado em:", caminho)

	f, err := os.Create(caminho)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write([]byte(fmt.Appendf(nil, "Dólar: %s\n", cotacao.Bid)))
	if err != nil {
		return err
	}

	log.Println("Cotação salva com sucesso")
	return nil
}
