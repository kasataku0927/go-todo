package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/kasataku0927/go-todo/server/models"
)

var todoModel models.TodoModelInterface

func init() {
	// テストモードかどうかを検出
	// go test実行時にセットされる環境変数を検出
	if os.Getenv("GO_TEST") == "1" || strings.HasSuffix(os.Args[0], ".test") {
		// テストモードなので初期化をスキップ
		return
	}

	// 実行モードなのでモデルを初期化
	var err error
	m, err := models.NewTodoModel()
	if err != nil {
		log.Fatalf("TodoModelの初期化エラー: %v", err)
	}
	todoModel = m
}

// TodosHandler はTodoコレクション全体に対するリクエストを処理
func TodosHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		GetTodos(w, r)
	case http.MethodPost:
		CreateTodo(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// TodoHandler は特定のTodoに対するリクエストを処理
func TodoHandler(w http.ResponseWriter, r *http.Request) {
	id, err := ExtractIDFromPath(r.URL.Path, "/todos/")
	if err != nil {
		http.Error(w, "Invalid ID in path", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		GetTodo(w, r, id)
	case http.MethodPut:
		UpdateTodo(w, r, id)
	case http.MethodDelete:
		DeleteTodo(w, r, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// GetTodos は全てのTodoを取得するハンドラー
func GetTodos(w http.ResponseWriter, r *http.Request) {
	todos, err := todoModel.GetTodos()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todos)
}

// GetTodo は特定のTodoを取得するハンドラー
func GetTodo(w http.ResponseWriter, r *http.Request, id int) {
	todo, err := todoModel.GetTodo(id)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Todo not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todo)
}

// CreateTodo は新しいTodoを作成するハンドラー
func CreateTodo(w http.ResponseWriter, r *http.Request) {
	var todo models.Todo
	err := json.NewDecoder(r.Body).Decode(&todo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = todoModel.CreateTodo(&todo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(todo)
}

// UpdateTodo は特定のTodoを更新するハンドラー
func UpdateTodo(w http.ResponseWriter, r *http.Request, id int) {
	var todo models.Todo
	err := json.NewDecoder(r.Body).Decode(&todo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	existingTodo, err := todoModel.GetTodo(id)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Todo not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 更新するフィールドをマージ
	if todo.Title != "" {
		existingTodo.Title = todo.Title
	}
	if todo.Description != "" {
		existingTodo.Description = todo.Description
	}
	existingTodo.Completed = todo.Completed

	err = todoModel.UpdateTodo(id, existingTodo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(existingTodo)
}

// DeleteTodo は特定のTodoを削除するハンドラー
func DeleteTodo(w http.ResponseWriter, r *http.Request, id int) {
	err := todoModel.DeleteTodo(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ExtractIDFromPath はURLパスからIDを抽出するヘルパー関数
func ExtractIDFromPath(path string, prefix string) (int, error) {
	// /todos/ を除去してIDを取得
	idStr := strings.TrimPrefix(path, prefix)
	if idStr == "" {
		return 0, fmt.Errorf("ID not found in path")
	}
	return strconv.Atoi(idStr)
}
