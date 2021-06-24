package client

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

var StrOut string

func ClientFirst() {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	// url := "https://habr.com/ru"
	url := "https://google.com"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Request error", err)
	}

	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		fmt.Println("Response error", err)
		StrOut = "ERROR"
		return
	}
	defer resp.Body.Close()
	var outStr []byte
	resp.Body.Read(outStr)
	StrOut = string(outStr)

}