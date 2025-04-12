package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/kasataku0927/go-todo/database/db"
	"github.com/kasataku0927/go-todo/server/handlers"
	"github.com/kasataku0927/go-todo/server/models"
	"github.com/rs/cors"
)

var testServer *httptest.Server

// テスト環境のセットアップ
func setupTestServer(t *testing.T) func() {
	// 環境変数の設定
	os.Setenv("DB_NAME", "go_todo_integration_test")

	// データベースの初期化
	err := db.Init()
	if err != nil {
		t.Fatalf("データベース初期化エラー: %v", err)
	}

	// テスト用のデータベースを準備
	_, err = db.DB.Exec(`
	DROP TABLE IF EXISTS todos;
	CREATE TABLE todos (
		id INT AUTO_INCREMENT PRIMARY KEY,
		title VARCHAR(100) NOT NULL,
		description TEXT,
		completed BOOLEAN DEFAULT false,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)
	`)
	if err != nil {
		t.Fatalf("テーブル作成エラー: %v", err)
	}

	// テストサーバーを設定
	mux := http.NewServeMux()

	// APIエンドポイントの設定
	mux.HandleFunc("/todos", handlers.TodosHandler)
	mux.HandleFunc("/todos/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/todos/" {
			http.NotFound(w, r)
			return
		}
		handlers.TodoHandler(w, r)
	})

	// CORSの設定
	handler := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
		AllowedHeaders: []string{"*"},
	}).Handler(mux)

	// テストサーバーを起動
	testServer = httptest.NewServer(handler)

	// クリーンアップ関数を返す
	return func() {
		// テストサーバーを終了
		testServer.Close()

		// テスト用のテーブルを削除
		_, _ = db.DB.Exec("DROP TABLE IF EXISTS todos")
	}
}

// テストデータの挿入
func seedIntegrationTestData(t *testing.T) {
	// テストデータを挿入
	todos := []struct {
		title       string
		description string
		completed   bool
	}{
		{"Integration Test Todo 1", "Integration Description 1", false},
		{"Integration Test Todo 2", "Integration Description 2", true},
	}

	for _, todo := range todos {
		currentTime := time.Now().Format("2006-01-02 15:04:05")
		_, err := db.DB.Exec(
			"INSERT INTO todos (title, description, completed, created_at, updated_at) VALUES (?, ?, ?, ?, ?)",
			todo.title, todo.description, todo.completed, currentTime, currentTime,
		)
		if err != nil {
			t.Fatalf("テストデータの挿入に失敗しました: %v", err)
		}
	}
}

// GET /todos のテスト - 全Todoを取得
func TestIntegrationGetAllTodos(t *testing.T) {
	cleanup := setupTestServer(t)
	defer cleanup()

	// テストデータを挿入
	seedIntegrationTestData(t)

	// APIリクエスト
	resp, err := http.Get(testServer.URL + "/todos")
	if err != nil {
		t.Fatalf("リクエスト実行中にエラーが発生しました: %v", err)
	}
	defer resp.Body.Close()

	// ステータスコード確認
	if resp.StatusCode != http.StatusOK {
		t.Errorf("不正なステータスコード: got %v, want %v", resp.StatusCode, http.StatusOK)
	}

	// レスポンスボディのデコード
	var todos []models.Todo
	err = json.NewDecoder(resp.Body).Decode(&todos)
	if err != nil {
		t.Fatalf("レスポンスのデコードに失敗しました: %v", err)
	}

	// データ検証
	if len(todos) != 2 {
		t.Errorf("返されたTodoの数が一致しません: got %v, want %v", len(todos), 2)
	}

	if todos[0].Title != "Integration Test Todo 1" {
		t.Errorf("Todoのタイトルが一致しません: got %v, want %v", todos[0].Title, "Integration Test Todo 1")
	}
}

// POST /todos のテスト - 新しいTodoを作成
func TestIntegrationCreateTodo(t *testing.T) {
	cleanup := setupTestServer(t)
	defer cleanup()

	// 新しいTodo
	newTodo := models.Todo{
		Title:       "New Integration Todo",
		Description: "New Integration Description",
		Completed:   false,
	}

	// JSONに変換
	todoJSON, err := json.Marshal(newTodo)
	if err != nil {
		t.Fatalf("JSONへの変換に失敗しました: %v", err)
	}

	// POSTリクエスト
	resp, err := http.Post(
		testServer.URL+"/todos",
		"application/json",
		bytes.NewBuffer(todoJSON),
	)
	if err != nil {
		t.Fatalf("リクエスト実行中にエラーが発生しました: %v", err)
	}
	defer resp.Body.Close()

	// ステータスコード確認
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("不正なステータスコード: got %v, want %v", resp.StatusCode, http.StatusCreated)
	}

	// レスポンスボディのデコード
	var createdTodo models.Todo
	err = json.NewDecoder(resp.Body).Decode(&createdTodo)
	if err != nil {
		t.Fatalf("レスポンスのデコードに失敗しました: %v", err)
	}

	// データ検証
	if createdTodo.Title != "New Integration Todo" {
		t.Errorf("作成されたTodoのタイトルが一致しません: got %v, want %v", createdTodo.Title, "New Integration Todo")
	}

	// IDが設定されていることを確認
	if createdTodo.ID == 0 {
		t.Errorf("作成されたTodoにIDが設定されていません")
	}
}

// GET /todos/:id のテスト - 特定のTodoを取得
func TestIntegrationGetTodo(t *testing.T) {
	cleanup := setupTestServer(t)
	defer cleanup()

	// テストデータを挿入
	seedIntegrationTestData(t)

	// 全Todoを取得してIDを得る
	resp, err := http.Get(testServer.URL + "/todos")
	if err != nil {
		t.Fatalf("リクエスト実行中にエラーが発生しました: %v", err)
	}
	var todos []models.Todo
	err = json.NewDecoder(resp.Body).Decode(&todos)
	if err != nil {
		t.Fatalf("レスポンスのデコードに失敗しました: %v", err)
	}
	resp.Body.Close()

	if len(todos) == 0 {
		t.Fatalf("テストデータが見つかりません")
	}

	// 最初のTodoのIDを使用
	todoID := todos[0].ID

	// 特定のTodoを取得
	resp, err = http.Get(fmt.Sprintf("%s/todos/%d", testServer.URL, todoID))
	if err != nil {
		t.Fatalf("リクエスト実行中にエラーが発生しました: %v", err)
	}
	defer resp.Body.Close()

	// ステータスコード確認
	if resp.StatusCode != http.StatusOK {
		t.Errorf("不正なステータスコード: got %v, want %v", resp.StatusCode, http.StatusOK)
	}

	// レスポンスボディのデコード
	var todo models.Todo
	err = json.NewDecoder(resp.Body).Decode(&todo)
	if err != nil {
		t.Fatalf("レスポンスのデコードに失敗しました: %v", err)
	}

	// データ検証
	if todo.ID != todoID {
		t.Errorf("取得されたTodoのIDが一致しません: got %v, want %v", todo.ID, todoID)
	}

	// 存在しないIDでテスト
	resp, err = http.Get(fmt.Sprintf("%s/todos/9999", testServer.URL))
	if err != nil {
		t.Fatalf("リクエスト実行中にエラーが発生しました: %v", err)
	}
	defer resp.Body.Close()

	// ステータスコード確認（404が期待される）
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("存在しないIDに対して不正なステータスコード: got %v, want %v", resp.StatusCode, http.StatusNotFound)
	}
}

// PUT /todos/:id のテスト - 特定のTodoを更新
func TestIntegrationUpdateTodo(t *testing.T) {
	cleanup := setupTestServer(t)
	defer cleanup()

	// テストデータを挿入
	seedIntegrationTestData(t)

	// 全Todoを取得してIDを得る
	resp, err := http.Get(testServer.URL + "/todos")
	if err != nil {
		t.Fatalf("リクエスト実行中にエラーが発生しました: %v", err)
	}
	var todos []models.Todo
	err = json.NewDecoder(resp.Body).Decode(&todos)
	resp.Body.Close()

	if len(todos) == 0 {
		t.Fatalf("テストデータが見つかりません")
	}

	// 最初のTodoを更新
	todoID := todos[0].ID
	updatedTodo := models.Todo{
		Title:       "Updated Integration Todo",
		Description: "Updated Integration Description",
		Completed:   true,
	}

	// JSONに変換
	todoJSON, err := json.Marshal(updatedTodo)
	if err != nil {
		t.Fatalf("JSONへの変換に失敗しました: %v", err)
	}

	// リクエスト作成
	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/todos/%d", testServer.URL, todoID), bytes.NewBuffer(todoJSON))
	if err != nil {
		t.Fatalf("リクエスト作成に失敗しました: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// リクエスト送信
	client := &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("リクエスト実行中にエラーが発生しました: %v", err)
	}
	defer resp.Body.Close()

	// ステータスコード確認
	if resp.StatusCode != http.StatusOK {
		t.Errorf("不正なステータスコード: got %v, want %v", resp.StatusCode, http.StatusOK)
	}

	// レスポンスボディのデコード
	var respTodo models.Todo
	err = json.NewDecoder(resp.Body).Decode(&respTodo)
	if err != nil {
		t.Fatalf("レスポンスのデコードに失敗しました: %v", err)
	}

	// データ検証
	if respTodo.Title != "Updated Integration Todo" {
		t.Errorf("更新されたTodoのタイトルが一致しません: got %v, want %v", respTodo.Title, "Updated Integration Todo")
	}
	if !respTodo.Completed {
		t.Errorf("更新されたTodoの完了ステータスが一致しません: got %v, want %v", respTodo.Completed, true)
	}
}

// DELETE /todos/:id のテスト - 特定のTodoを削除
func TestIntegrationDeleteTodo(t *testing.T) {
	cleanup := setupTestServer(t)
	defer cleanup()

	// テストデータを挿入
	seedIntegrationTestData(t)

	// 全Todoを取得してIDを得る
	resp, err := http.Get(testServer.URL + "/todos")
	if err != nil {
		t.Fatalf("リクエスト実行中にエラーが発生しました: %v", err)
	}
	var todos []models.Todo
	err = json.NewDecoder(resp.Body).Decode(&todos)
	resp.Body.Close()

	if len(todos) == 0 {
		t.Fatalf("テストデータが見つかりません")
	}

	// 最初のTodoを削除
	todoID := todos[0].ID

	// リクエスト作成
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/todos/%d", testServer.URL, todoID), nil)
	if err != nil {
		t.Fatalf("リクエスト作成に失敗しました: %v", err)
	}

	// リクエスト���信
	client := &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("リクエスト実行中にエラーが発生しました: %v", err)
	}
	defer resp.Body.Close()

	// ステータスコード確認
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("不正なステータスコード: got %v, want %v", resp.StatusCode, http.StatusNoContent)
	}

	// 削除されたことを確認
	resp, err = http.Get(fmt.Sprintf("%s/todos/%d", testServer.URL, todoID))
	if err != nil {
		t.Fatalf("リクエスト実行中にエラーが発生しました: %v", err)
	}
	defer resp.Body.Close()

	// ステータスコード確認（404が期待される）
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("削除されたTodoに対して不正なステータスコード: got %v, want %v", resp.StatusCode, http.StatusNotFound)
	}
}
