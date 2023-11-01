package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
	"os"
)

type Cotacao struct {
	Usdbrl struct {
		Bid string `json:"bid"`
	} `json:"USDBRL"`
}

func main() {
	// Para criar gerar o arquivo é necessário rodar o servidor primeiro e depois o cliente para que o arquivo seja criado
	// Caso contrário o arquivo não será criado
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		panic(err)
	}
	res, err := http.DefaultClient.Do(req)

	
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	// io.Copy(os.Stdout, res.Body)

	body, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	var cotacao Cotacao
	err = json.Unmarshal(body, &cotacao)

	//Create file
	file, err := os.Create("cotacao.txt")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao criar arquivo: %v\n", err)
	}
	defer file.Close()
    _, err = file.WriteString(fmt.Sprintf("Dólar:%s","{"+cotacao.Usdbrl.Bid+"}"))
	fmt.Println("Arquivo criado com sucesso!")

}
