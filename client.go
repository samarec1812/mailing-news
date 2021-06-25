package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type Posts struct {
	ID string `json:"id"`
	Date string `json:"date"`
}
// Переделать application/json в другой

type Client struct {
	BaseURL *url.URL
	UserAgent string

	httpClient *http.Client
}

// Извлекаем методы newRequest(), do() чтобы можно было переиспользовать во всех вызовах API
func (c *Client) newRequest(method, path, typeRequest string, body interface{}) (*http.Request, error) {
	rel := &url.URL{Path: path}
	u := c.BaseURL.ResolveReference(rel)

	str := fmt.Sprintf("%v", body)
	request, err := http.NewRequest(method, u.String(), bytes.NewBufferString(str))
	if err != nil {
		return nil, err
	}
	if body != nil {
		request.Header.Set("Content-Type", typeRequest)
	}
	request.Header.Set("Accept",  typeRequest)
	request.Header.Set("User-Agent", c.UserAgent)

	return request, nil

}

func (c *Client) do(ctx context.Context, request *http.Request, v interface{}) (*http.Response, error) {
	request = request.WithContext(ctx)
	resp, err := c.httpClient.Do(request)
	if err != nil {
		select {
		case <- ctx.Done():
			return nil, ctx.Err()
		default:
		}
		return nil, err
	}
	defer resp.Body.Close()

	// body, err := ioutil.ReadAll(resp.Body)
	// log.Printf(string(body))
	// fmt.Println(resp.Body)

	err = json.NewDecoder(resp.Body).Decode(v)

	return resp, err
}

func (c *Client) ListNews(ctx context.Context) ([]Posts, error){
	request, err := c.newRequest("GET", "/post", "application/json", nil)
	if err != nil {
		return nil, err
	}
	var posts []Posts
	_, err = c.do(ctx, request, &posts)
	return posts, err
}
var ID uint64

// Получение последнего ID
func (c *Client) GetLastID(ctx context.Context) (string, error) {

	posts, err := c.ListNews(ctx)
	return posts[len(posts)-1].ID, err
}

// Добавление новой новости
func (c *Client) PostNews(ctx context.Context) ([]Posts, error) {
	lastID, err := c.GetLastID(ctx)
	strID := ""
	strDate := ""
	if err != nil {
		strID = "0"
		strDate = time.Now().Format(time.RFC822)
	} else {
		num, _ := strconv.ParseUint(lastID, 10, 64)
		strID = strconv.FormatUint(num + 1, 10)
		strDate = time.Now().Format(time.RFC822)

	}
	postBody, _ := json.Marshal(Posts{
		strID,
		strDate,
	})
	// fmt.Println(string(postBody))
	responseBody := bytes.NewBuffer((postBody))
	// fmt.Println(responseBody)

	request, err := c.newRequest("POST", "/post", "application/json", responseBody)

	if err != nil {
		return nil, err
	}
	var post []Posts
	resp, err := c.do(ctx, request, &post)
	if resp.StatusCode != 200 {
		log.Println("HTTP-Error", resp.StatusCode)
	}

	return post, err
}

func NewClient(urlNew string) *Client {
	newUrl, _ := url.Parse(urlNew)
	return &Client{
		httpClient: &http.Client{},
		BaseURL: newUrl,
		UserAgent: "my-user-agent",
	}
}
func main() {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 5 *time.Second)
	defer cancel()
	client := NewClient("http://localhost:8081/")
	id, err := client.GetLastID(ctx)
	// news, err := client.PostNews(ctx)
	if err == context.DeadlineExceeded {
		fmt.Println(err)
		return
	}
	fmt.Println(id)

	_, err = client.PostNews(ctx)
	if err == context.DeadlineExceeded {
		fmt.Println(err)
		return
	}

	id, err = client.GetLastID(ctx)
	// news, err := client.PostNews(ctx)
	if err == context.DeadlineExceeded {
		fmt.Println(err)
		return
	}

	fmt.Println(id)

}


