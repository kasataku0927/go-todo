package models

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

type TodoModel struct {
	db *sql.DB
}

func NewTodoModel() *TodoModel {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	connectionString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", user, password, host, port, dbName)

	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		panic(err)
	}
	return &TodoModel{db: db}
}

type Todo struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Completed   bool   `json:"completed"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

func (t *TodoModel) GetTodos() ([]Todo, error) {
	// Get all todos from the database
	query := "SELECT id, title, description, completed, created_at, updated_at FROM todos"
	rows, err := t.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var todos []Todo
	for rows.Next() {
		var todo Todo
		err := rows.Scan(&todo.ID, &todo.Title, &todo.Description, &todo.Completed, &todo.CreatedAt, &todo.UpdatedAt)
		if err != nil {
			return nil, err
		}
		todos = append(todos, todo)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return todos, nil
}

func (t *TodoModel) CreateTodo(todo *Todo) error {
	// Insert a new todo into the database
	currentTime := time.Now().Format("2006-01-02 15:04:05")
	todo.CreatedAt = currentTime
	todo.UpdatedAt = currentTime
	query := "INSERT INTO todos (title, description, completed, created_at, updated_at) VALUES (?, ?, ?, ?, ?)"
	result, err := t.db.Exec(query, todo.Title, todo.Description, todo.Completed, todo.CreatedAt, todo.UpdatedAt)
	if err != nil {
		return err
	}

	// Retrieve the ID of the newly inserted todo
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	todo.ID = int(id)

	return nil
}

func (t *TodoModel) UpdateTodo(id int, todo *Todo) error {
	// Update an existing todo in the database
	currentTime := time.Now().Format("2006-01-02 15:04:05")
	todo.UpdatedAt = currentTime
	query := "UPDATE todos SET title = ?, description = ?, completed = ?, updated_at = ? WHERE id = ?"
	_, err := t.db.Exec(query, todo.Title, todo.Description, todo.Completed, todo.UpdatedAt, id)
	if err != nil {
		return err
	}
	return nil
}

func (t *TodoModel) DeleteTodo(id int) error {
	// Delete an existing todo from the database
	query := "DELETE FROM todos WHERE id = ?"
	_, err := t.db.Exec(query, id)
	if err != nil {
		return err
	}
	return nil
}

func (t *TodoModel) GetTodo(id int) (*Todo, error) {
	// Get a todo from the database by id
	query := "SELECT id, title, description, completed, created_at, updated_at FROM todos WHERE id = ?"
	row := t.db.QueryRow(query, id)

	var todo Todo
	err := row.Scan(&todo.ID, &todo.Title, &todo.Description, &todo.Completed, &todo.CreatedAt, &todo.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &todo, nil
}
