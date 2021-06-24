package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Sources struct {
	Href string `json:"href"`
	Title string `json:"title"`
}

type ArcticlesArray struct {
	Link string `json:"link"`
	Published string `json:"published"`
	Source Sources `json:"source"`
	SubArticles []string `json:"-"`
	Title string `json:"title"`
}

type FeedParametr struct {
	Language string `json:"language"`
	Link string `json:"link"`
	Rights string `json:"rights"`
	Subtitle string `json:"subtitle"`
	Title string `json:"title"`
	Updated string `json:"updated"`
	
}
type News struct {
	Articles []ArcticlesArray `json:"articles"`
	Feed FeedParametr `json:"feed"`
}


func main() {
	// Данное API разрешает только 3 запроса в час
	url := "https://google-news.p.rapidapi.com/v1/source_search?source=nytimes.com&lang=en&country=US"

	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add("x-rapidapi-key", "72a1274a7dmshd2adfe34fd79578p12c05cjsn02c05332354d")
	req.Header.Add("x-rapidapi-host", "google-news.p.rapidapi.com")

	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	str := News{}
	err := json.Unmarshal(body, &str)
	if err != nil {
		fmt.Println(err)
		return
	} else {
		fmt.Println(str.Feed.Language)
	}
	// fmt.Println(res)
	fmt.Println(string(body))

}