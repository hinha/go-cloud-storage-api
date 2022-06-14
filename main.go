package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"path"
	"time"
)

type (
	Response struct {
		Data interface{} `json:"data"`
	}
)

const (
	BucketName = "your-bucket"
)

func main() {
	client, err := NewGCPBucketClient(context.Background(), BucketName)
	if err != nil {
		panic(err)
	}
	r := mux.NewRouter()
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Hello World")
	})
	r.HandleFunc("/drive/files", func(w http.ResponseWriter, r *http.Request) {

		object, err := client.ListFiles(r.Context(), 30)
		if err != nil {
			writeError(w, err)
			return
		}

		response := Response{}
		response.Data = object
		files, _ := json.Marshal(response)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(files)
	})

	r.HandleFunc("/drive/upload", func(w http.ResponseWriter, r *http.Request) {
		file, handler, err := r.FormFile("file")
		if err != nil {
			panic(err)
		}
		defer file.Close()

		folder := r.FormValue("folder")

		err = client.UploadWriter(r.Context(), file, handler.Header, path.Join(folder, handler.Filename))
		if err != nil {
			writeError(w, err)
			return
		}
		response := Response{}
		response.Data = map[string]string{"file": path.Join(folder, handler.Filename)}
		files, _ := json.Marshal(response)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(files)
	})

	r.HandleFunc("/drive/download/{object}/{file}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)

		data, err := client.UploadReader(r.Context(), path.Join(vars["object"], vars["file"]))
		if err != nil {
			writeError(w, err)
			return
		}

		response := Response{}
		response.Data = map[string]string{"base64": base64.StdEncoding.EncodeToString(data)}
		files, _ := json.Marshal(response)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(files)
	})

	srv := &http.Server{
		Handler: r,
		Addr:    "127.0.0.1:9091",
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}

func writeError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%v", err)
}
