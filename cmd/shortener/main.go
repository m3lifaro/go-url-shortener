package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
)

func main() {
	cache := map[string]string{}
	cache["ya"] = "https://ya.ru/"
	cache["go"] = "https://golang.org/"
	if err := run(cache); err != nil {
		log.Fatal(err)
	}
}

func redirect(w http.ResponseWriter, r *http.Request, cache map[string]string) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	redirectUrl := cache[r.PathValue("id")]
	if redirectUrl == "" {
		http.NotFound(w, r)
		return
	}
	println("Redirecting to: " + redirectUrl)
	w.Header().Set("Location", redirectUrl)
	w.WriteHeader(http.StatusTemporaryRedirect)
	//http.Redirect(w, r, "https://ya.rus", http.StatusTemporaryRedirect)
}

func shortenURL(url string) string {
	hash := md5.Sum([]byte(url))
	shortCode := hex.EncodeToString(hash[:])[:8]
	return shortCode
}

func urlConvertionHandler(w http.ResponseWriter, r *http.Request, cache map[string]string) {
	if r.Method != "POST" {
		println("Not a POST request")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Got error while reading request body", http.StatusInternalServerError)
		log.Printf("Error reading request body: %v", err)
		return
	}
	textUrl := string(body)
	println("URL: " + textUrl)
	shortedUrl := shortenURL(textUrl)
	cache[shortedUrl] = textUrl
	println("Shorten url: " + shortedUrl)
	for k, v := range cache {
		fmt.Println(k, v)
	}
	w.Header().Set("Content-Type", "text/plain")
	_, _ = w.Write([]byte("http://localhost:8080/" + shortedUrl))
}

func run(cache map[string]string) error {
	mux := http.NewServeMux()

	mux.HandleFunc("/{id}/", func(w http.ResponseWriter, r *http.Request) {
		redirect(w, r, cache)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		urlConvertionHandler(w, r, cache)
	})
	return http.ListenAndServe(":8080", mux)
}
