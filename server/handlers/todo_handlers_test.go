package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/kasataku0927/go-todo/server/models"
)

// モックTodoモデルの定義
type MockTodoModel struct {
	todos []*models.Todo
}

// 以下のメソッド実装によってMockTodoModelはTodoModelInterfaceを満たす
func NewMockTodoModel() *MockTodoModel {
	// テストデータの作成
	return &MockTodoModel{
		todos: []*models.Todo{
			{ID: 1, Title: "Test Todo 1", Description: "Test Description 1", Completed: false, CreatedAt: "2025-04-12 10:00:00", UpdatedAt: "2025-04-12 10:00:00"},
			{ID: 2, Title: "Test Todo 2", Description: "Test Description 2", Completed: true, CreatedAt: "2025-04-12 11:00:00", UpdatedAt: "2025-04-12 11:00:00"},
		},
	}
}

// GetTodos実装
func (m *MockTodoModel) GetTodos() ([]models.Todo, error) {
	result := make([]models.Todo, len(m.todos))
	for i, todo := range m.todos {
		result[i] = *todo
	}
	return result, nil
}

// GetTodo実装
func (m *MockTodoModel) GetTodo(id int) (*models.Todo, error) {
	for _, todo := range m.todos {
		if todo.ID == id {
			return todo, nil
		}
	}
	// 存在しないIDの場合はsql.ErrNoRowsを返す
	return nil, sql.ErrNoRows
}

// CreateTodo実装
func (m *MockTodoModel) CreateTodo(todo *models.Todo) error {
	todo.ID = len(m.todos) + 1
	todo.CreatedAt = "2025-04-12 12:00:00"
	todo.UpdatedAt = "2025-04-12 12:00:00"
	m.todos = append(m.todos, todo)
	return nil
}

// UpdateTodo実装
func (m *MockTodoModel) UpdateTodo(id int, todo *models.Todo) error {
	for i, t := range m.todos {
		if t.ID == id {
			todo.UpdatedAt = "2025-04-12 13:00:00"
			m.todos[i] = todo
			return nil
		}
	}
	return nil
}

// DeleteTodo実装
func (m *MockTodoModel) DeleteTodo(id int) error {
	for i, todo := range m.todos {
		if todo.ID == id {
			m.todos = append(m.todos[:i], m.todos[i+1:]...)
			return nil
		}
	}
	return nil
}

// テスト用のセットアップとクリーンアップ
func setupTest(t *testing.T) (*MockTodoModel, func()) {
	// 元のモデルを保持
	originalModel := todoModel

	// モックモデルを作成してセット
	mockModel := NewMockTodoModel()
	todoModel = mockModel

	// モデルがセットされているか確認
	if todoModel == nil {
		t.Fatalf("todoModel has not been initialized!")
	}

	// クリーンアップ関数を返す
	return mockModel, func() {
		todoModel = originalModel
	}
}

// GET /todos のテスト
func TestGetTodos(t *testing.T) {
	mockModel, cleanup := setupTest(t)
	defer cleanup()

	// リクエスト作成
	req, err := http.NewRequest("GET", "/todos", nil)
	if err != nil {
		t.Fatal(err)
	}

	// レスポンスレコーダー作成
	rr := httptest.NewRecorder()

	// ハンドラー実行
	handler := http.HandlerFunc(GetTodos)
	handler.ServeHTTP(rr, req)

	// ステータスコード確認
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("不正なステータスコードが返されました: got %v want %v",
			status, http.StatusOK)
	}

	// レスポンスボディ検証
	var todos []models.Todo
	err = json.Unmarshal(rr.Body.Bytes(), &todos)
	if err != nil {
		t.Errorf("JSONのデコードに失敗しました: %v", err)
	}

	// モックデータと一致するか検証
	if len(todos) != len(mockModel.todos) {
		t.Errorf("返されたTodoの数が一致しません: got %v want %v",
			len(todos), len(mockModel.todos))
	}
}

// POST /todos のテスト
func TestCreateTodo(t *testing.T) {
	_, cleanup := setupTest(t)
	defer cleanup()

	// テストケース定義
	testCases := []struct {
		name           string
		requestBody    string
		expectedStatus int
	}{
		{
			name:           "正常なTodo作成",
			requestBody:    `{"title":"New Todo","description":"New Description"}`,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "不正なJSON",
			requestBody:    `{"title":}`,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// リクエスト作成
			req, _ := http.NewRequest("POST", "/todos", bytes.NewBufferString(tc.requestBody))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			// ハンドラー実行
			handler := http.HandlerFunc(CreateTodo)
			handler.ServeHTTP(rr, req)

			// ステータスコード確認
			if status := rr.Code; status != tc.expectedStatus {
				t.Errorf("予期しないステータスコード: got %v want %v",
					status, tc.expectedStatus)
			}
		})
	}
}

// GET /todos/:id のテスト
func TestGetTodo(t *testing.T) {
	_, cleanup := setupTest(t)
	defer cleanup()

	// 存在するIDでテスト
	todoID := 1

	// リクエストパス作成（実際のハンドラーと一致するよう修正）
	req, err := http.NewRequest("GET", "/todos/"+strconv.Itoa(todoID), nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	// ハンドラー実行（ID引数付き）
	GetTodo(rr, req, todoID)

	// ステータスコード確認
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("不正なステータスコードが返されました: got %v want %v",
			status, http.StatusOK)
	}

	// レスポンスボディ検証
	var todo models.Todo
	err = json.Unmarshal(rr.Body.Bytes(), &todo)
	if err != nil {
		t.Errorf("JSONのデコードに失敗しました: %v", err)
	}

	// IDが一致するか検証
	if todo.ID != todoID {
		t.Errorf("不正なTodo IDが返されました: got %v want %v",
			todo.ID, todoID)
	}

	// 存在しないIDでテスト
	nonExistentID := 999
	req, _ = http.NewRequest("GET", "/todos/"+strconv.Itoa(nonExistentID), nil)
	rr = httptest.NewRecorder()

	// ハンドラー実行
	GetTodo(rr, req, nonExistentID)

	// ステータスコード確認（404が期待される）
	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("存在しないIDに対して不正なステータスコードが返されました: got %v want %v",
			status, http.StatusNotFound)
	}
}

// PUT /todos/:id のテスト
func TestUpdateTodo(t *testing.T) {
	_, cleanup := setupTest(t)
	defer cleanup()

	// 既存のTodoを更新
	todoID := 1
	updatedTodo := `{"title":"Updated Todo","description":"Updated Description","completed":true}`

	req, _ := http.NewRequest("PUT", "/todos/"+strconv.Itoa(todoID), bytes.NewBufferString(updatedTodo))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	// ハンドラー実行
	UpdateTodo(rr, req, todoID)

	// ステータスコード確認
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("不正なステータスコードが返されました: got %v want %v",
			status, http.StatusOK)
	}

	// レスポンスボディ検証
	var todo models.Todo
	err := json.Unmarshal(rr.Body.Bytes(), &todo)
	if err != nil {
		t.Errorf("JSONのデコードに失敗しました: %v", err)
	}

	// 更新が反映されているか検証
	if todo.Title != "Updated Todo" {
		t.Errorf("Todoタイトルが更新されていません: got %v want %v",
			todo.Title, "Updated Todo")
	}
	if !todo.Completed {
		t.Errorf("Todoの完了ステータスが更新されていません")
	}
}

// DELETE /todos/:id のテスト
func TestDeleteTodo(t *testing.T) {
	_, cleanup := setupTest(t)
	defer cleanup()

	// 既存のTodoを削除
	todoID := 1

	req, _ := http.NewRequest("DELETE", "/todos/"+strconv.Itoa(todoID), nil)
	rr := httptest.NewRecorder()

	// ハンドラー実行
	DeleteTodo(rr, req, todoID)

	// ステータスコード確認
	if status := rr.Code; status != http.StatusNoContent {
		t.Errorf("不正なステータスコードが返されました: got %v want %v",
			status, http.StatusNoContent)
	}
}
