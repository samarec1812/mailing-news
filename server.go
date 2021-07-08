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
	Text string `json:"-"`
	Description string `json:"description"`
}

type files struct {
	FileName zip.File
	Date     string
}



// createZip создаёт из структуры post соответствующий zip архив с содержанием Text и именем ID структуры
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

		// создание директории с именами ID
		err = os.MkdirAll("./news/" + post.ID, os.ModePerm)
		if err != nil {
			fmt.Println(err)
			return
		}
		//// создание xml файла c описанием
		//filenameJSON := "data.json"
		//fileJSON, _ := os.Create("./news/" + post.ID + "/" + filenameJSON)
		//jsonWriter := io.Writer(fileJSON)
		//enc := json.NewEncoder(jsonWriter).Encode(post)
		//
		//enc.Indent(" ", "    ")
		//if err := enc.Encode(post); err != nil {
		//	fmt.Printf("error: %v\n", err)
		//	return
		//}
		fileJSON, _ := json.MarshalIndent(post, "", " ")
	    filenameJSON := "data.json"
		_ = ioutil.WriteFile("./news/" + post.ID + "/" + filenameJSON, fileJSON, 0644)


		// создание zip архива с новостью
		newZipFile, err := os.Create("./news/" + post.ID + "/data.zip")
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
	// _, err = io.Copy(writer, fileTxt)
	//if err != nil {
	//	log.Println(err)
	//	return
	//}

}

// функция генерирует из source каталога/файла zip архив с путём target
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
		{"1", "25 Jun 21 19:06 MSK", "Статья 1", "Статья о чём-то 1"},
		{"2", "26 Jun 21 18:01 MSK", "Статья 2", "Статья о чём-то 2"},
		{"3", "27 Jun 21 20:16 MSK", "Статья 3ss", "Статья о чём-то 3"},
	}
	for i := 4; i < 20; i++ {
		posts = append(posts, postsStruct{fmt.Sprintf("%v", i),
			time.Now().Format(time.RFC822), fmt.Sprintf("Статья %v", i), fmt.Sprintf("Статья о чём-то %v", i)})
	}

	for i := 0; i < len(posts); i++ {
		createZip(posts[i])
	}
	zipName := "./send.zip"
	err := zipit("./news", zipName)
	if err != nil {
		log.Println(err)
		return
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
			if r.FormValue("NumLastNews") == "" && r.FormValue("ID") == "" && r.FormValue("Hash") == "" && r.FormValue("Archive") == "" {
				productsJson, _ := json.Marshal(posts)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)

				w.Write(productsJson)

			} else if r.FormValue("Archive") != "" {

				// defer os.Remove(zipName)
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
				} else if r.FormValue("Hash") != "" {
					file, _ := os.Open("./send.zip")
					defer file.Close()

					bytesReadZIP, err := ioutil.ReadAll(file)
					if err != nil {
						fmt.Println("Ошибка чтения")
						return
					}
					hash := sha256.New()
					hashSum := hash.Sum(bytesReadZIP)
					w.Header().Set("Content-Type", "application/octet-stream")
					w.Header().Add("Content-Hash", "Hash-256")
					w.WriteHeader(http.StatusOK)
					w.Write(hashSum)

				}
			}

		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	err = http.ListenAndServe(":8081", nil)
	if err != nil {
		fmt.Println(err)
		return
	}

}
