package main

import (
	"fmt"
	"net/http"
	"time"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Serving: %s\n", r.URL.Path)
	fmt.Printf("Served: %s\n", r.Host)
}

func timeHandler(w http.ResponseWriter, r *http.Request) {
	 t := time.Now().Format(time.RFC1123)
	 Body := "The current time is:"
	 fmt.Fprintf(w, "<h1 align=\"center\">%s</h1>", Body)
	 fmt.Fprintf(w, "<h2 align=\"center\">%s</h2>", t)
	 fmt.Fprintf(w, "Serving: %s\n", r.URL.Path)
	 fmt.Printf("Served time for: %s\n", r.Host)
}

func main() {
	http.HandleFunc("/",handler)
	http.HandleFunc("/time", timeHandler)

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println(err)
		return
	}
}
