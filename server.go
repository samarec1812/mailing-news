package main

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

type postsStruct struct {
	ID   string `json:"id"`
	Date string `json:"date"`
	Text string `json:"text"`
}

func main() {

	posts := []postsStruct{
		{"1", "25 Jun 21 19:06 MSK", "Статья 1"},
		{"2", "26 Jun 21 18:01 MSK", "Статья 2"},
		{"3", "27 Jun 21 20:16 MSK", "Статья 3"},
	}
	for i := 4; i < 20; i++ {
		posts = append(posts, postsStruct{fmt.Sprintf("%v", i),
			time.Now().Format(time.RFC822), fmt.Sprintf("Статья %v", i)})
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// отдаём обычный HTML
		fileContents, err := ioutil.ReadFile("index.html")
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Write(fileContents)
	})

	http.HandleFunc("/post", func(w http.ResponseWriter, r *http.Request) {
		//str := client.Client()
		//fmt.Fprintln(w, str)
		fmt.Println("request: ", r.URL.Path)
		fmt.Println("method request: ", r.Method)
		defer r.Body.Close()

		// Switch для разных типов запросов
		switch r.Method {
		// GET для получения данных
		case http.MethodGet:
			r.ParseForm()
			fmt.Println(r.Form)
			fmt.Println(r.FormValue("NumLastNews"))
			if r.FormValue("NumLastNews") == "" && r.FormValue("ID") == "" && r.FormValue("Hash") == "" {
				productsJson, _ := json.Marshal(posts)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write(productsJson)
			} else {
				if r.FormValue("NumLastNews") != "" {
					lastNumPost, err := strconv.Atoi(r.FormValue("NumLastNews"))
					if err != nil {
						log.Println(err)
						w.WriteHeader(http.StatusBadRequest)
						return
					}
					if lastNumPost < 0 || lastNumPost > len(posts) {
						log.Println(errors.New("number of requested posts is not allowed"))
						w.WriteHeader(http.StatusBadRequest)
						return
					}

					productsJson, _ := json.Marshal(posts[len(posts)-lastNumPost:])
					w.Header().Set("Content-Type", "application/json")

					w.WriteHeader(http.StatusOK)
					w.Write(productsJson)
				} else if r.FormValue("ID") != "" && r.FormValue("Hash") != "" {
					ID, err := strconv.Atoi(r.FormValue("ID"))
					if err != nil {
						log.Println(err)
						w.WriteHeader(http.StatusBadRequest)
						return
					}
					hash := sha256.New()
					hashSum := hash.Sum([]byte(posts[ID - 1].Date))
					if string(hashSum) != r.FormValue("Hash") {
						productsJson, _ := json.Marshal(posts[ID - 1])
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusOK)
						w.Write(productsJson)
					} else {
						w.Header().Set("Content-Type", "text/plain")
						w.WriteHeader(http.StatusOK)
						w.Write([]byte("Already up to date"))
					}
				}
			}
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

	err := http.ListenAndServe(":8081", nil)
	if err != nil {
		fmt.Println(err)
		return
	}

}
