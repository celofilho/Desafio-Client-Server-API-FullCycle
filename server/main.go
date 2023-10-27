package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	_ "strings"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

type CotacaoJson struct {
	Usdbrl struct {
		Code       string `json:"code,omitempty"`
		Codein     string `json:"codein,omitempty"`
		Name       string `json:"name,omitempty"`
		High       string `json:"high,omitempty"`
		Low        string `json:"low,omitempty"`
		VarBid     string `json:"varBid,omitempty"`
		PctChange  string `json:"pctChange,omitempty"`
		Bid        string `json:"bid,omitempty"`
		Ask        string `json:"ask,omitempty"`
		Timestamp  string `json:"timestamp,omitempty"`
		CreateDate string `json:"create_date,omitempty"`
	} `json:"USDBRL"`
}

type Cotacao struct {
	Id   string
	Bid  string
	Data string
}

func newCotacao(bid string, data string) *Cotacao {
	return &Cotacao{
		Id:   uuid.New().String(),
		Bid:  bid,
		Data: data,
	}
}

func main() {

	http.HandleFunc("/cotacao", CotacaoHandler)
	http.ListenAndServe(":8080", nil)

}

func insertProduct(db *sql.DB, cotacao *Cotacao) error {

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Millisecond*10)
	defer cancel()

	stmt, err := db.Prepare("insert into cotacao(id, bid, date) values(?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(cotacao.Id, cotacao.Bid, cotacao.Data)

	//Timeout de 10ms para o insert
	select {
	case <-time.After(5 * time.Millisecond):
		log.Println("Request Inserted")
	case <-ctx.Done():
		log.Println("Process of Database timed out")
	}

	if err != nil {
		return err
	}
	return nil
}

func CotacaoHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/cotacao" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	cotacao, error := getCotacao()
	if error != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(cotacao)

}

func getCotacao() (*CotacaoJson, error) {

	//Context

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/last/USD-BRL", nil)

	resp, error := http.DefaultClient.Do(req)

	defer resp.Body.Close()

	body, error := io.ReadAll(resp.Body)
	if error != nil {
		return nil, error
	}
	var cotacao CotacaoJson
	error = json.Unmarshal(body, &cotacao)
	if error != nil {
		return nil, error
	}
	jsonPuro := []byte(fmt.Sprintf(`{"USDBRL":{"code":"","codein":"","name":"","high":"","low":"","varBid":"","pctChange":"","bid":"%s","ask":"","timestamp":"","create_date":""}}`, cotacao.Usdbrl.Bid))
	var cotacaoPuro CotacaoJson

	error = json.Unmarshal(jsonPuro, &cotacaoPuro)
	if error != nil {
		return nil, error
	}

	//Timeout de 200ms para a requisição
	select {
	case <-time.After(100 * time.Millisecond):
		log.Println("Request Success")
	case <-ctx.Done():
		log.Println("Process timed out")
	}

	//Insert into database
	db, err := sql.Open("sqlite3", "./goexpertDesafioClientServer.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	err = insertProduct(db, newCotacao(cotacaoPuro.Usdbrl.Bid, time.Now().Format("2006-01-02 15:04:05")))
	if err != nil {
		panic(err)
	}

	return &cotacaoPuro, nil
}
