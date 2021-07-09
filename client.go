package main

import (
	_ "archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"time"
)

type Posts struct {
	ID   string `json:"id"`
	Date string `json:"date"`
}

type Client struct {
	BaseURL    *url.URL
	UserAgent  string
	httpClient *http.Client
}

// Структура с параметрами запросов
type OptionsURL struct {
	NumLastNews uint
	Hash        string
	ID          string
	Archive     string
}

type Hash struct {
	HashStr []byte `json:"hash_str"`
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
	optName4 := (reflect.Indirect(reflect.ValueOf(opt))).Type().Field(3).Name

	if opt.NumLastNews != 0 {
		q.Add(optName, fmt.Sprintf("%v", opt.NumLastNews))
	}
	if opt.Hash != "" {
		q.Add(optName2, opt.Hash)
	}
	if opt.ID != "" {
		q.Add(optName3, opt.ID)
	}
	if opt.Archive != "" {
		q.Add(optName4, opt.Archive)
	}

	request.URL.RawQuery = q.Encode()
	// fmt.Println(request.URL.String())

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
		if resp.Header.Get("Content-Hash") == "Hash-256" {

			byteArr, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Fatal(err)
			}
			err = ioutil.WriteFile("./serverhash.txt", byteArr, 0644)
			if err != nil {
				log.Println(err)
			}

			return resp, err
		} else if resp.Header.Get("Content-News") == "lastnews" {
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Fatal(err)
			}
			out, err := os.Create("./lastNewsClient.zip")
			if err != nil {
				log.Println(err)
				return resp, err
			}
			defer out.Close()
			_, err = io.Copy(out, bytes.NewReader(body))
			return resp, err
		} else {

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		out, err := os.Create("./client.zip")
		if err != nil {
			log.Println(err)
			return resp, err
		}
		defer out.Close()
		_, err = io.Copy(out, bytes.NewReader(body))
		return resp, err
		}
	}
	return resp, err

}

// Получение информации о всех новостях
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





// GetNews для получение новостей или update новостей
func (c *Client) GetNews(ctx context.Context) (error) {
	// Если у клиента отсутствует архив с новостями, то err != nil =>
	// делаем запрос на получение архива новостей
	file, err := os.Open("./resp.zip")
	if err != nil {
		err = c.GetArchiveNews(ctx)
		return err
	}

	err = c.GetHashSumOfNews(ctx)
	if err != nil {
		log.Fatal(err)
		return err
	}

	// иначе клиент считывает имеющиеся байты архива и вычисляет hash сумму
	bytesReadZIP, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println("Ошибка чтения")
		return err
	}
	hash := sha256.New()
	hashSum := hash.Sum(bytesReadZIP)
	// fmt.Println(string(hashSum))
	// открываем файл с hash суммой сервера


	fileHash, err := os.Open("./serverHash.txt")
	if err != nil {
		log.Fatal(err)
		return err
	}
	bytesServerHash, err := ioutil.ReadAll(fileHash)
	if err != nil {
		log.Fatal(err)
		return err
	}
	if string(hashSum) == string(bytesServerHash) {
		fmt.Println("Данные актуальны")
		return err
	} else {
		fmt.Println("Новости обновлены")
		err = c.GetArchiveNews(ctx)
	}

	return err

}
// GetHashSumOfNews отправляет GET запрос на получение hash суммы архива на сервере
func (c *Client) GetHashSumOfNews(ctx context.Context) error {
	opt := OptionsURL{Hash: "yes"}

	request, err := c.newRequest("GET", "/post", "application/octet-stream", opt, nil)
	if err != nil {
		return err
	}

	_, err = c.doImplementation(ctx, request, struct{}{})
	return err
}
//// Получение единого архива со всеми новостями
//func (c *Client) GetArchiveNews(ctx context.Context) ([]zip.File, error) {
//	opt := OptionsURL{Archive: "yes"}
//	request, err := c.newRequest("GET", "/post", "application/zip", opt, nil)
//	if err != nil {
//		return nil, err
//	}
//	var info []zip.File
//	_, err = c.doImplementation(ctx, request, info)
//	fmt.Println(info)
//	return info, err
//}

// GetArchiveNews отправляет GET запрос на получение единого архива со всеми новостями
func (c *Client) GetArchiveNews(ctx context.Context) (error) {
	var info string
	opt := OptionsURL{Archive: "yes"}
	request, err := c.newRequest("GET", "/post", "application/octet-stream", opt, nil)
	if err != nil {
		return err
	}
	_, err = c.doImplementation(ctx, request, &info)

	return err
}

// Получение ID последней новости
func (c *Client) GetLastID(ctx context.Context) (string, error) {
	posts, err := c.GetAllNews(ctx)
	if posts == nil {
		return "", err
	}
	return posts[len(posts)-1].ID, err
}

// Создание нового клиента
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

	fmt.Println()
	//
	//
	//startTime := time.Now()
	// Получение update по ID


	//		getNews, err := client.GetUpdateByID(ctx, post[ID-1])
	//		if err == context.DeadlineExceeded {
	//			fmt.Println(err)
	//			return
	//		} else if getNews == nil {
	//			fmt.Println("Already up to date")
	//		}
	//	})
	//
	//	secondTime := time.Now()
	//	if secondTime.Sub(startTime).Seconds() > 50 {
	//		timer.Stop()
	//		break
	//	}
	//
	//}


	//status, err := client.GetArchiveNews(ctx)
	//if err == context.DeadlineExceeded {
	//	fmt.Println(err)
	//	return
	//}
	//fmt.Println(status)


	err = client.GetNews(ctx)
	if err == context.DeadlineExceeded {
		fmt.Println(err)
		return
	}
	fmt.Println("Данные получены")
}
