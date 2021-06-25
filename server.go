package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type postsStruct struct {
	ID string `json:"id"`
	Date string `json:"date"`
}


func main() {

	posts := []postsStruct{
		{"1", "25 Jun 21 19:06 MSK"},
		{"2", "26 Jun 21 18:01 MSK"},
		{"3", "27 Jun 21 20:16 MSK"},
	}

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



	http.HandleFunc("/post", func (w http.ResponseWriter, r *http.Request){
		//str := client.Client()
		//fmt.Fprintln(w, str)
		fmt.Println("request ", r.URL.Path)
		defer r.Body.Close()

		// Switch для разных типов запросов
		switch r.Method {
		// GET для получения данных
		case http.MethodGet:
			productsJson, _ := json.Marshal(posts)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(productsJson)
		// POST для добавление чего-то нового
		case http.MethodPost:
			decoder := json.NewDecoder(r.Body)
			news := postsStruct{}
			// преобразуем json в структуру
			err := decoder.Decode(&news)
			fmt.Println(news)
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			posts = append(posts, news)



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
