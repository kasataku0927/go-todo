package main
import 	(
	"fmt"
	"log"

	"github.com/kasataku0927/go-todo/internal/db"
)

func main() {
	// データベースの初期化
	db.Init()
	// データベースの接続確認
	fmt.Println("Hello, Go Todo App!")
	log.Println("Application started successfully")
}