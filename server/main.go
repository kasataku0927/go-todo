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
	db.Init()

	// ServeMuxを作成
	mux := http.NewServeMux()

	// ルートハンドラを定義
	mux.HandleFunc("/todos", handlers.TodoHandler)

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
