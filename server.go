package main

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type postsStruct struct {
	ID   string `json:"id"`
	Date string `json:"date"`
	Text string `json:"text"`
}

type files struct {
	FileName zip.File
	Date string
}

func createZip(post postsStruct) {
	fileTxt, err := os.Create(post.ID + ".txt")
	if err != nil {
		log.Println(err)
		return
	}
	// удаляем текстовый после создания
	defer os.Remove(fileTxt.Name())
	// закрываем текстовый
	defer fileTxt.Close()


	_, err = fileTxt.WriteString(post.Text)
	// fmt.Println(post.Text)
	if err != nil {
		log.Println(err)
		return
	}
	newZipFile, err := os.Create("./news/" + post.ID + ".zip")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer newZipFile.Close()
	zipWriter := zip.NewWriter(newZipFile)
	defer zipWriter.Close()

	info, err := fileTxt.Stat()
	if err != nil {
		log.Println(err)
		return
	}
	header, err := zip.FileInfoHeader(info)
	if err != nil {
		log.Println(err)
		return
	}
	header.Name = fileTxt.Name()
	header.Method = zip.Deflate
	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		log.Println(err)
		return
	}
	writer.Write([]byte(post.Text))
	_, err = io.Copy(writer, fileTxt)
	//if err != nil {
	//	log.Println(err)
	//	return
	//}

}


func zipit(source, target string) error {
	zipfile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	info, err := os.Stat(source)
	if err != nil {
		return nil
	}

	var baseDir string
	if info.IsDir() {
		baseDir = filepath.Base(source)
	}

	filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		if baseDir != "" {
			header.Name = filepath.Join(baseDir, strings.TrimPrefix(path, source))
		}

		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(writer, file)
		return err
	})

	return err
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

	for i := 0; i < len(posts); i++ {
		createZip(posts[i])
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
			if r.FormValue("NumLastNews") == "" && r.FormValue("ID") == "" && r.FormValue("Hash") == "" && r.FormValue("Archive") == ""  {
				productsJson, _ := json.Marshal(posts)

				w.Header().Set("Content-Type", "application/json")
			 	w.WriteHeader(http.StatusOK)

				w.Write(productsJson)

			} else if r.FormValue("Archive") != ""   {

				//w.Header().Set("Content-Type", "application/zip")
				//w.Header().Set("Content-Disposition", "attachment")
				////w.Header().Set("Content-Disposition", "form-data")
				//err := zipit("./news", "./send.zip")
				//if err != nil {
				//	log.Println(err)
				//	return
				//}
				//http.ServeFile(w, r, "send.zip")
				zipName := "./send.zip"
				err := zipit("./news", zipName)
				if err != nil {
						log.Println(err)
						return
					}
				defer os.Remove(zipName)
				file, err := os.Open(zipName)
				if err != nil {
					log.Fatal(err)
				}
				defer file.Close()
				mes, _ := ioutil.ReadAll(file)
				w.Header().Set("Content-Type", "application/octet-stream")
				w.Write(mes)


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


