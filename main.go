package main

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

type todo struct {
	ID string `json:"id"`
	UserID    int    `json:"user_id"`
	Item string `json:"item"`
	Completed bool `json:"completed"`
}

type user struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

var todos = []todo{}

// context contains info about the incoming http request

func getTodos(context *gin.Context, db *sql.DB) {
	rows, err := db.Query("SELECT id, item, completed FROM todos")
	if err != nil {
		context.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var todos []todo
	for rows.Next() {
		var t todo
		if err := rows.Scan(&t.ID, &t.Item, &t.Completed); err != nil {
			context.IndentedJSON(http.StatusInternalServerError, gin.H{"error" : err.Error()})
			return
		}
		todos = append(todos, t)
	}
	context.IndentedJSON(http.StatusOK, todos)
}

func addTodo(context *gin.Context, db *sql.DB) {
	var newTodo todo

	// we are binding the json from request body to newTodo of todo type. If its not of that type, it' give error, in which case, we don't wanna go forward with the execution, so we just return
	if err := context.BindJSON(&newTodo); err!=nil {
		context.IndentedJSON(http.StatusBadRequest, gin.H{"error" : err.Error()})
		return
	}

	// appending new todo to the todos array
	_, err := db.Exec("INSERT INTO todos (item, completed) VALUES ($1, $2)", newTodo.Item, newTodo.Completed)
	if err != nil {
		context.IndentedJSON(http.StatusBadRequest, gin.H{"error" : err.Error()})
		return
	}
	context.IndentedJSON(http.StatusCreated, newTodo)
}


func getTodo(context *gin.Context, db *sql.DB) {
	id := context.Param("id")
	todo, err := getTodoById(id, db)

	if err != nil{
		context.IndentedJSON(http.StatusNotFound, gin.H{"message":"Todo not found"})
		return
	}

	context.IndentedJSON(http.StatusOK, todo)
}

func toggleTodoStatus(context *gin.Context, db *sql.DB){
	id := context.Param("id")
	todo, err := getTodoById(id, db)

	if err!=nil {
		context.IndentedJSON(http.StatusNotFound, gin.H{"message": "Todo not found"})
		return
	}

	_, err = db.Exec("UPDATE todos SET completed = $1 WHERE id = $2", !todo.Completed, todo.ID)
	if err != nil {
		context.IndentedJSON(http.StatusInternalServerError, gin.H{"error" : err.Error()})
		return
	}

	todo.Completed = !todo.Completed
	context.IndentedJSON(http.StatusOK, todo)
}

func getTodoById(id string, db *sql.DB) (*todo, error){
	var t todo
	err := db.QueryRow("SELECT id, item, completed FROM todos WHERE id = $1", id).Scan(&t.ID, &t.Item, &t.Completed)
	if err != nil {
		if err == sql.ErrNoRows {
			// If the error is sql.ErrNoRows, it means that the query didn't find a matching row for the given ID. In this case, the function returns nil as the todo item and an error indicating that the todo was not found.
			return nil, errors.New("Todo not found")
		}
		return nil, err
	}

	return &t, nil
}

func createTable(db *sql.DB) error {
	// Define the SQL statement to create the table
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS todos (
		id SERIAL PRIMARY KEY,
		item TEXT NOT NULL,
		completed BOOLEAN NOT NULL
	);
	`

	// Execute the SQL statement to create the table
	_, err := db.Exec(createTableSQL)
	if err != nil {
		return err
	}
	return nil
}


func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	databaseURL := fmt.Sprintf("postgres://%s:%s@localhost/%s?sslmode=disable", dbUser, dbPassword, dbName)

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = createTable(db)
	if err != nil {
		log.Fatal(err)
	}

	router := gin.Default()
	router.GET("/todos", func(c *gin.Context) { getTodos(c, db) })
	router.GET("/todos/:id", func(c *gin.Context) { getTodo(c, db) })
	router.PATCH("/todos/:id", func(c *gin.Context) { toggleTodoStatus(c, db) })
	router.POST("/todos", func(c *gin.Context) { addTodo(c, db) })
	router.Run("localhost:9090")
}