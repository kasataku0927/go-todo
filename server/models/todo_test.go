package models

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

var testDB *sql.DB
var testModel *TodoModel

// テスト前の準備
func setupTestDB(t *testing.T) func() {
	// テスト用の.envファイルを読み込む
	err := godotenv.Load("../../.env")
	if err != nil {
		t.Fatalf("テスト用の.envファイルを読み込めませんでした: %v", err)
	}

	// 環境変数からDB接続情報を取得
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")

	// テスト用のデータベース名を設定（本番DBとは別にする）
	dbName := "go_todo_test"

	// テスト用データベースに接続
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/", user, password, host, port)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatalf("データベース接続に失敗しました: %v", err)
	}

	// テスト用データベースを作成（存在していれば一旦削除）
	_, err = db.Exec("DROP DATABASE IF EXISTS " + dbName)
	if err != nil {
		t.Fatalf("テスト用データベースの削除に失敗しました: %v", err)
	}

	_, err = db.Exec("CREATE DATABASE " + dbName)
	if err != nil {
		t.Fatalf("テスト用データベースの作成に失敗しました: %v", err)
	}

	// 作成したテスト用データベースに接続を切り替え
	db.Close()
	testDsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", user, password, host, port, dbName)
	testDB, err = sql.Open("mysql", testDsn)
	if err != nil {
		t.Fatalf("テスト用データベースへの接続に失敗しました: %v", err)
	}

	// todosテーブルを作成
	_, err = testDB.Exec(`
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
		t.Fatalf("テーブル作成に失敗しました: %v", err)
	}

	// テスト用のモデルを作成
	testModel = &TodoModel{db: testDB}

	// クリーンアップ関数を返す
	return func() {
		testDB.Close()

		// テスト用のデータベースも削除
		tempDB, err := sql.Open("mysql", dsn)
		if err == nil {
			_, _ = tempDB.Exec("DROP DATABASE IF EXISTS " + dbName)
			tempDB.Close()
		}
	}
}

// テスト用のデータを挿入
func seedTestData(t *testing.T) {
	todos := []struct {
		title       string
		description string
		completed   bool
	}{
		{"Test Todo 1", "Description 1", false},
		{"Test Todo 2", "Description 2", true},
		{"Test Todo 3", "Description 3", false},
	}

	for _, todo := range todos {
		currentTime := time.Now().Format("2006-01-02 15:04:05")
		_, err := testDB.Exec(
			"INSERT INTO todos (title, description, completed, created_at, updated_at) VALUES (?, ?, ?, ?, ?)",
			todo.title, todo.description, todo.completed, currentTime, currentTime,
		)
		if err != nil {
			t.Fatalf("テストデータの挿入に失敗しました: %v", err)
		}
	}
}

func TestGetTodos(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	// テストデータを挿入
	seedTestData(t)

	// GetTodosをテスト
	todos, err := testModel.GetTodos()
	if err != nil {
		t.Fatalf("GetTodos実行中にエラーが発生しました: %v", err)
	}

	// 結果の検証
	if len(todos) != 3 {
		t.Errorf("期待されるTodoの数と異なります: got %v want %v", len(todos), 3)
	}

	// 特定のTodoの内容を検証
	if todos[0].Title != "Test Todo 1" {
		t.Errorf("1番目のTodoのタイトルが期待と異なります: got %v want %v", todos[0].Title, "Test Todo 1")
	}
	if todos[1].Completed != true {
		t.Errorf("2番目のTodoの完了ステータスが期待と異なります: got %v want %v", todos[1].Completed, true)
	}
}

func TestCreateTodo(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	// 新しいTodoを作成
	newTodo := &Todo{
		Title:       "New Test Todo",
		Description: "New Test Description",
		Completed:   false,
	}

	// CreateTodoをテスト
	err := testModel.CreateTodo(newTodo)
	if err != nil {
		t.Fatalf("CreateTodo実行中にエラーが発生しました: %v", err)
	}

	// IDが設定されていることを確認
	if newTodo.ID == 0 {
		t.Errorf("作成されたTodoにIDが設定されていません")
	}

	// データベースから取得して内容を検証
	todo, err := testModel.GetTodo(newTodo.ID)
	if err != nil {
		t.Fatalf("GetTodo実行中にエラーが発生しました: %v", err)
	}

	if todo.Title != "New Test Todo" {
		t.Errorf("作成されたTodoのタイトルが期待と異なります: got %v want %v", todo.Title, "New Test Todo")
	}
}

func TestUpdateTodo(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	// テストデータを挿入
	seedTestData(t)

	// 最初のTodoを取得
	todos, _ := testModel.GetTodos()
	if len(todos) == 0 {
		t.Fatalf("テストデータが挿入されていません")
	}

	todoID := todos[0].ID

	// 更新用のTodoを作成
	updatedTodo := &Todo{
		ID:          todoID,
		Title:       "Updated Todo",
		Description: "Updated Description",
		Completed:   true,
	}

	// UpdateTodoをテスト
	err := testModel.UpdateTodo(todoID, updatedTodo)
	if err != nil {
		t.Fatalf("UpdateTodo実行中にエラーが発生しました: %v", err)
	}

	// 更新されたTodoを取得して検証
	todo, err := testModel.GetTodo(todoID)
	if err != nil {
		t.Fatalf("GetTodo実行中にエラーが発生しました: %v", err)
	}

	if todo.Title != "Updated Todo" {
		t.Errorf("更新されたTodoのタイトルが期待と異なります: got %v want %v", todo.Title, "Updated Todo")
	}
	if !todo.Completed {
		t.Errorf("更新されたTodoの完了ステータスが期待と異なります: got %v want %v", todo.Completed, true)
	}
}

func TestDeleteTodo(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	// テストデータを挿入
	seedTestData(t)

	// 最初のTodoを取得
	todos, _ := testModel.GetTodos()
	if len(todos) == 0 {
		t.Fatalf("テストデータが挿入されていません")
	}

	todoID := todos[0].ID

	// DeleteTodoをテスト
	err := testModel.DeleteTodo(todoID)
	if err != nil {
		t.Fatalf("DeleteTodo実行中にエラーが発生しました: %v", err)
	}

	// 削除されたことを確認
	todo, err := testModel.GetTodo(todoID)
	if err != sql.ErrNoRows && todo != nil {
		t.Errorf("Todoが削除されていません")
	}

	// 残りのTodoの数を確認
	remainingTodos, _ := testModel.GetTodos()
	if len(remainingTodos) != 2 {
		t.Errorf("削除後の残りTodoの数が期待と異なります: got %v want %v", len(remainingTodos), 2)
	}
}
