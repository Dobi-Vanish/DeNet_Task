package main

import (
	"database/sql"
	"embed"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/pressly/goose/v3"
	"log"
	"net/http"
	"os"
	"reward-service/data"
	"time"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
)

var counts int64

//go:embed migrations/*.sql
var EmbedMigrations embed.FS

type Config struct {
	Repo      data.Repository
	Client    *http.Client
	SecretKey string
}

// main starts the server and establishing connection to database
func main() {
	log.Println("Starting reward service")
	err := godotenv.Load("example.env")
	if err != nil {
		log.Panic("Error loading .env file", err)
	}
	var webPort = os.Getenv("PORT")

	// connect to DB
	conn := connectToDB()
	if conn == nil {
		log.Fatal("Can't connect to Postgres!")
	}

	goose.SetBaseFS(EmbedMigrations)

	migrationsDir := os.Getenv("GOOSE_MIGRATION_DIR")

	if err := goose.SetDialect("postgres"); err != nil {
		panic(err)
	}

	if err := goose.Up(conn, migrationsDir); err != nil {
		panic(err)
	}

	// set up config
	app := Config{
		Client: &http.Client{},
	}
	app.setupRepo(conn)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	err = srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}

// openDB establishes a connection to the PostgreSQL database using the provided Data Source Name (DSN)
func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx/v4", dsn)
	fmt.Println(db)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

// connectToDB connect to Postgres with provided dsn
func connectToDB() *sql.DB {
	dsn := os.Getenv("DSN")

	for {
		connection, err := openDB(dsn)
		if err != nil {
			log.Println("Postgres not yet ready ...")
			counts++
		} else {
			log.Println("Connected to Postgres!")
			return connection
		}

		if counts > 10 {
			log.Println(err)
			return nil
		}

		log.Println("Backing off for two seconds....")
		time.Sleep(2 * time.Second)
		continue
	}
}

// setupRepo sets new postgres repository
func (app *Config) setupRepo(conn *sql.DB) {
	if conn == nil {
		log.Fatal("Database connection is nil")
	}
	db := data.NewPostgresRepository(conn)
	app.Repo = db
}
