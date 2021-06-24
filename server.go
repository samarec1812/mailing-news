package main

import (
	"./client"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)




func main() {

	ourNews := client.Client()

	http.HandleFunc("/", func (w http.ResponseWriter, r *http.Request){
		// отдаём обычный HTML
		fileContents, err := ioutil.ReadFile("index.html")
		if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
		w.Write(fileContents)
	})
	http.HandleFunc("/time", func (w http.ResponseWriter, r *http.Request){
		//str := client.Client()
		//fmt.Fprintln(w, str)
		fmt.Println("request ", r.URL.Path)
		defer r.Body.Close()

		// Switch для разных типов запросов
		switch r.Method {
		// GET для получения данных
		case http.MethodGet:
			productsJson, _ := json.Marshal(ourNews)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(productsJson)
		// POST для добавление чего-то нового
		case http.MethodPost:
			decoder := json.NewDecoder(r.Body)
			news := client.NewsPost{}
			// преобразуем json в структуру
			err := decoder.Decode(&news)
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			news.UserInput.Lang = "UK"
			news.UserInput.Size += 1

		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})




	err := http.ListenAndServe(":8081",nil)
	if err != nil {
		fmt.Println(err)
		return
	}

}
