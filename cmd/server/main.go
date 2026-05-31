package main

import (
	"fmt"
	"net/http"
)

func main() {
	// 動作確認用のヘルスチェックAPI
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok")
	})

	// Go APIサーバーを8080番で起動
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
