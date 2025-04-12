package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/kasataku0927/go-todo/database/db"
	"github.com/kasataku0927/go-todo/server/handlers"
	"github.com/rs/cors"
)

func main() {
	// データベースの初期化
	err := db.Init()
	if err != nil {
		log.Fatalf("データベース初期化エラー: %v", err)
	}

	// ServeMuxを作成
	mux := http.NewServeMux()

	// RESTfulなAPIエンドポイントを設定
	mux.HandleFunc("/todos", handlers.TodosHandler)

	// /todos/123 のような特定のIDに対するハンドラー
	mux.HandleFunc("/todos/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/todos/" {
			http.NotFound(w, r)
			return
		}
		handlers.TodoHandler(w, r)
	})

	// CORS設定
	handler := cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:3000"},
		AllowedMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
		AllowedHeaders: []string{"*"},
	}).Handler(mux)

	// サーバーの起動
	fmt.Println("Server is running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}
