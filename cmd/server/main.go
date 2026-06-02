package main

import (
	"fmt"
	"net/http"
)

func main() {
	
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok")
	})

	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
