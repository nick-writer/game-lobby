package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

type Player struct {
	Name  string `json:"name"`
	Level int    `json:"level"`
}

var db *sql.DB

func main() {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://postgres:postgres@localhost:5432/lobby?sslmode=disable"
	}

	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Ошибка подключения к БД: %v", err)
	}
	defer db.Close()

	for i := 0; i < 10; i++ {
		err = db.Ping()
		if err == nil {
			break
		}
		log.Printf("Ожидание базы данных... (%d/10)", i+1)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatalf("База данных недоступна: %v", err)
	}

	initDB()

	// Раздача статики
	http.Handle("/", http.FileServer(http.Dir("./static")))
	// API
	http.HandleFunc("/players", playersHandler)

	fmt.Println("Сервер лобби запущен на http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func initDB() {
	query := `
	CREATE TABLE IF NOT EXISTS players (
		id SERIAL PRIMARY KEY,
		name TEXT NOT NULL,
		level INT NOT NULL
	);`
	if _, err := db.Exec(query); err != nil {
		log.Fatalf("Ошибка создания таблицы: %v", err)
	}

	var count int
	db.QueryRow("SELECT COUNT(*) FROM players").Scan(&count)
	if count == 0 {
		_, err := db.Exec("INSERT INTO players (name, level) VALUES ($1, $2), ($3, $4)", "Alice", 5, "Bob", 3)
		if err != nil {
			log.Printf("Ошибка заполнения БД: %v", err)
		}
	}
}

func playersHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {

	case http.MethodGet:
		// Чтение всех игроков
		rows, err := db.Query("SELECT name, level FROM players")
		if err != nil {
			http.Error(w, "Ошибка чтения из БД", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var players []Player
		for rows.Next() {
			var p Player
			if err := rows.Scan(&p.Name, &p.Level); err != nil {
				http.Error(w, "Ошибка парсинга данных", http.StatusInternalServerError)
				return
			}
			players = append(players, p)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(players)

	case http.MethodPost:
		// Добавление игрока
		var p Player
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, "некорректный JSON", http.StatusBadRequest)
			return
		}
		if p.Name == "" {
			http.Error(w, "поле name обязательно", http.StatusBadRequest)
			return
		}

		_, err := db.Exec("INSERT INTO players (name, level) VALUES ($1, $2)", p.Name, p.Level)
		if err != nil {
			http.Error(w, "Ошибка записи в БД", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(p)

	case http.MethodDelete:
		// Удаление игрока по имени
		name := r.URL.Query().Get("name")
		if name == "" {
			http.Error(w, "параметр name обязателен", http.StatusBadRequest)
			return
		}

		result, err := db.Exec("DELETE FROM players WHERE name = $1", name)
		if err != nil {
			http.Error(w, "Ошибка удаления из БД", http.StatusInternalServerError)
			return
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			http.Error(w, "Игрок не найден", http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "метод не поддерживается", http.StatusMethodNotAllowed)
	}
}