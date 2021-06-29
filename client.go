package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"time"
)

type Posts struct {
	ID   string `json:"id"`
	Date string `json:"date"`
}

// Переделать application/json в другой

type Client struct {
	BaseURL    *url.URL
	UserAgent  string
	httpClient *http.Client
}

// Структура с параметрами запросов
type OptionsURL struct {
	NumLastNews uint
	Hash string
	ID string
}

// Извлекаем методы newRequest(), do() чтобы можно было переиспользовать во всех вызовах API
func (c *Client) newRequest(method, path, typeRequest string, opt OptionsURL, body interface{}) (*http.Request, error) {
	rel := &url.URL{Path: path}

	u := c.BaseURL.ResolveReference(rel)

	str := fmt.Sprintf("%v", body)
	request, err := http.NewRequest(method, u.String(), bytes.NewBufferString(str))

	// add request parameters
	q := url.Values{}
	// берём имя поля структуры как имя параметра запроса
	optName := (reflect.Indirect(reflect.ValueOf(opt))).Type().Field(0).Name
	optName2 := (reflect.Indirect(reflect.ValueOf(opt))).Type().Field(1).Name
	optName3 := (reflect.Indirect(reflect.ValueOf(opt))).Type().Field(2).Name

	if opt.NumLastNews != 0 {
		q.Add(optName, fmt.Sprintf("%v", opt.NumLastNews))
	}
	if opt.Hash != "" {
		q.Add(optName2, opt.Hash)
	}
	if opt.ID != "" {
		q.Add(optName3, opt.ID)
	}

	request.URL.RawQuery = q.Encode()
	fmt.Println(request.URL.String())

	if err != nil {
		return nil, err
	}
	if body != nil {
		request.Header.Set("Content-Type", typeRequest)
	}

	request.Header.Set("Accept", typeRequest)
	request.Header.Set("User-Agent", c.UserAgent)

	return request, nil

}

// Вынос метода Do структуры httpClient в отдельную функцию, для переиспользования в других функциях
func (c *Client) doImplementation(ctx context.Context, request *http.Request, v interface{}) (*http.Response, error) {
	request = request.WithContext(ctx)
	resp, err := c.httpClient.Do(request)
	if err != nil {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		return nil, err
	}
	defer resp.Body.Close()

	// body, err := ioutil.ReadAll(resp.Body)
	// log.Printf(string(body))
	// fmt.Println(resp.Body)
	if resp.Header.Get("Content-Type") == "application/json" {
		err = json.NewDecoder(resp.Body).Decode(v)
	} else {
		_, err = ioutil.ReadAll(resp.Body)
		// fmt.Println(string(text))
	}
	return resp, err
}





// Получение всех новостей
func (c *Client) GetAllNews(ctx context.Context) ([]Posts, error) {
	opt := OptionsURL{}
	request, err := c.newRequest("GET", "/post", "application/json", opt, nil)
	if err != nil {
		return nil, err
	}
	var posts []Posts
	_, err = c.doImplementation(ctx, request, &posts)
	return posts, err
}

// Получение последних N новостей
func (c *Client) GetNumLastNews(ctx context.Context, num uint) ([]Posts, error) {
	opt := OptionsURL{NumLastNews: num}
	// path := addOptionsURL("/post", OptionsURL{3})
	request, err := c.newRequest("GET", "/post", "application/json", opt, nil)
	if err != nil {
		return nil, err
	}
	var posts []Posts
	_, err = c.doImplementation(ctx, request, &posts)
	return posts, err
}

// Получение update новости по её ID
func (c *Client) GetUpdateByID(ctx context.Context, post Posts) (*Posts, error) {

	hash := sha256.New()
	hashSum := hash.Sum([]byte(post.Date))
	fmt.Println(string(hashSum))
	opt := OptionsURL{Hash: string(hashSum), ID: post.ID}
	request, err := c.newRequest("GET", "/post", "application/json", opt, nil)
	if err != nil {
		return nil, err
	}
	var posts *Posts
	_, err = c.doImplementation(ctx, request, &posts)
	return posts, err




}

var ID uint64

// Получение ID последней новости
func (c *Client) GetLastID(ctx context.Context) (string, error) {

	posts, err := c.GetAllNews(ctx)
	return posts[len(posts)-1].ID, err
}

func NewClient(urlNew string) *Client {
	newUrl, _ := url.Parse(urlNew)
	return &Client{
		httpClient: &http.Client{},
		BaseURL:    newUrl,
		UserAgent:  "my-user-agent",
	}
}


func main() {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	client := NewClient("http://localhost:8081/")

	post, err := client.GetAllNews(ctx)
	if err == context.DeadlineExceeded {
		fmt.Println(err)
		return
	}
	//id, err := client.GetLastID(ctx)
	//// news, err := client.PostNews(ctx)
	//if err == context.DeadlineExceeded {
	//	fmt.Println(err)
	//	return
	//}
	fmt.Println(post)
	var numLastNews uint
	_, err = fmt.Scanln(&numLastNews)
	if err != nil {
		fmt.Println(err)
		return
	}

	lastNews, err := client.GetNumLastNews(ctx, numLastNews)
	if err == context.DeadlineExceeded {
		fmt.Println(err)
		return
	}
	fmt.Println(lastNews)

	// Получение update по ID
	ID := 5
	getNews, err := client.GetUpdateByID(ctx, post[ID - 1])
	if err == context.DeadlineExceeded {
		fmt.Println(err)
		return
	} else if getNews == nil {
		fmt.Println("Already up to date")
		return
	}
	fmt.Println(getNews)

}
