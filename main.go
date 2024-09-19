package main

import (
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.Handle("/app/", http.StripPrefix("/app", http.FileServer(http.Dir("."))))
	mux.HandleFunc("/healthz", myHandler)

	server := &http.Server{
		Handler: mux,
		Addr:    ":8080",
	}

	server.ListenAndServe()
}

func myHandler(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("Content-Type", "text/plain; charset=utf-8")
	response.WriteHeader(200)
	response.Write([]byte("OK"))
}
