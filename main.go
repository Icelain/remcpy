package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"
)

func InitDir() error {
	if err := os.Mkdir("./store", os.ModePerm); !errors.Is(err, fs.ErrExist) {
		return err
	}
	return nil
}

type Router struct {
	mux    *http.ServeMux
	logger *slog.Logger
}

func NewRouter() *Router {
	return &Router{
		mux:    http.NewServeMux(),
		logger: slog.Default(),
	}
}

func IndexController(router *Router) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head><title>remcpy - Remote Copy Service</title></head>
<body>
	<h1>remcpy - Remote Copy Service</h1>
	<p>Upload: POST /@{identifier}</p>
	<p>Download: GET /@{identifier}</p>
</body>
</html>`)
	}
}

func DownloadController(router *Router) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if !strings.HasPrefix(path, "@") || len(path) <= 1 {
			http.Error(w, "Invalid file identifier format. Use /@{identifier}", http.StatusBadRequest)
			return
		}
		ident := path[1:]

		osFile, err := os.Open(fmt.Sprintf("./store/@%s", ident))
		if err != nil {
			if os.IsNotExist(err) {
				http.Error(w, "File not found", http.StatusNotFound)
			} else {
				http.Error(w, "Error reading provided file", http.StatusInternalServerError)
			}
			router.logger.Debug("Error reading file from disk: " + err.Error())
			return
		}
		defer osFile.Close()

		w.Header().Set("Content-Disposition", "attachment")
		w.Header().Set("Content-Type", "application/octet-stream")

		_, err = io.Copy(w, osFile)
		if err != nil {
			http.Error(w, "Error streaming file", http.StatusInternalServerError)
			router.logger.Debug("Error streaming file: " + err.Error())
			return
		}
	}
}

func UploadController(router *Router) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if !strings.HasPrefix(path, "@") || len(path) <= 1 {
			http.Error(w, "Invalid file identifier format. Use /@{identifier}", http.StatusBadRequest)
			return
		}
		ident := path[1:]

		fileReader, fileHeader, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "Error reading provided file", http.StatusBadRequest)
			router.logger.Debug("Error reading file from formdata: " + err.Error())
			return
		}
		defer fileReader.Close()

		osFile, err := os.Create(fmt.Sprintf("./store/@%s", ident))
		if err != nil {
			http.Error(w, "Internal error: file creation", http.StatusInternalServerError)
			router.logger.Debug("Error creating os file: " + err.Error())
			return
		}
		defer osFile.Close()

		n, err := io.Copy(osFile, fileReader)
		if err != nil {
			http.Error(w, "Internal error: file write failed", http.StatusInternalServerError)
			router.logger.Debug("Error writing file to disk: " + err.Error())
			return
		}

		downloadURL := fmt.Sprintf("/@%s", ident)

		response := fmt.Sprintf("Temporary remote copy made successfully.\nFile: %s\nSize Written: %d\nAccess at: GET %s",
			fileHeader.Filename, n, downloadURL)

		_, err = w.Write([]byte(response))
		if err != nil {
			router.logger.Debug("Error writing response to client: " + err.Error())
			return
		}
	}
}

func ApplyControllers(router *Router) {
	router.mux.HandleFunc("GET /", IndexController(router))

	router.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			return
		}

		if strings.HasPrefix(r.URL.Path, "/@") {
			switch r.Method {
			case "GET":
				DownloadController(router)(w, r)
			case "POST":
				UploadController(router)(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		} else {
			http.Error(w, "Invalid endpoint. Use /@{identifier}", http.StatusBadRequest)
		}
	})
}

func main() {
	port := flag.Uint("port", 5000, "Port to run remcpy on")
	flag.Parse()

	router := NewRouter()
	ApplyControllers(router)

	if err := InitDir(); err != nil {
		log.Fatal("Error creating store directory: " + err.Error())
	}

	log.Printf("remcpy listening on :%d", *port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), router.mux))
}
