package db

import (
	"database/sql"
	"fmt"
	"os"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

var DB *sql.DB

func Init() error {
	//環境変数ファイルを読み込む
	err := godotenv.Load()
	if err != nil {
		return fmt.Errorf("環境変数ファイルの読み込みエラー: %w", err)
	}

	// 環境変数から接続情報を取得
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")
	
	// DSN (Data Source Name) を構築
	// MySQLプラグインの認証方法を指定
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?allowNativePasswords=true", user, password, host, port, dbName)
	
	// MySQLデータベースに接続
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("データベース接続設定エラー: %w", err)
	}

	// 接続を確認
	err = DB.Ping()
	if err != nil {
		return fmt.Errorf("データベース接続確認エラー: %w", err)
	}

	fmt.Println("データベースに正常に接続されました")
	return nil
}