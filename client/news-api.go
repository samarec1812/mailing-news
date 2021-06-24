package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type InfoArticles struct {
	ID string `json:"_id"`
	Score float64 `json:"_score"`
	Author string `json:"author"`
	Authors []string `json:"authors"`
	CleanURL string `json:"clean_url"`
	Country string `json:"country"`
	IsOpinion bool `json:"is_opinion"`
	Link string `json:"link"`
	Media string `json:"-"`
	PublishedDate string `json:"published_date"`
	Rank uint `json:"rank"`
	Rights string `json:"rights"`
	Summary string `json:"summary"`
	Title string `json:"title"`
	Topic string `json:"topic"`
	TwitterAccount string `json:"twitter_account,omitempty"`
}

type UserInputStruct struct {
	From string `json:"from"`
	Lang string `json:"lang"`
	Page uint `json:"page"`
	Q string `json:"q"`
	Size uint `json:"size"`
	SortBy string `json:"sort_by"`
}

type NewsPost struct {
	Articles []InfoArticles `json:"articles"`
	Page uint `json:"page"`
	PageSize uint `json:"page_size"`
	Status string `json:"status"`
	TotalHits uint `json:"total_hits"`
	TotalPages uint `json:"total_pages"`
	UserInput UserInputStruct `json:"user_input"`
	
}

func Client() NewsPost {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	url := "https://free-news.p.rapidapi.com/v1/search?q=Elon%20Musk&lang=en"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Request error", err)
		return NewsPost{}
	}


	req.Header.Add("x-rapidapi-key", "72a1274a7dmshd2adfe34fd79578p12c05cjsn02c05332354d")
	req.Header.Add("x-rapidapi-host", "free-news.p.rapidapi.com")
	res, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		fmt.Println("Response error", err)
		return NewsPost{}
	}


	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	//fmt.Println(res)
	//fmt.Println(string(body))
	news := NewsPost{}
	err = json.Unmarshal(body, &news)
	if err != nil {
		fmt.Println(err)
		return NewsPost{}
	}
	return news

}

