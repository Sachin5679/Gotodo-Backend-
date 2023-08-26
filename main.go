package main

import (
	"net/http"
	"errors"
	"github.com/gin-gonic/gin"
)

type todo struct {
	ID string `json:"id"`
	Item string `json:"item"`
	Completed bool `json:"completed"`
}

var todos = []todo{
	{ID: "1", Item: "Clean Room", Completed: false},
	{ID: "2", Item: "Read book", Completed: false},
	{ID: "3", Item: "Record Video", Completed: false},
}

// context contains info about the incoming http request

func getTodos(context *gin.Context) {
	context.IndentedJSON(http.StatusOK, todos)
}

func addTodo(context *gin.Context) {
	var newTodo todo

	// we are binding the json from request body to newTodo of todo type. If its not of that type, it' give error, in which case, we don't wanna go forward with the execution, so we just return
	if err := context.BindJSON(&newTodo); err!=nil {
		return
	}

	// appending new todo to the todos array
	todos = append(todos, newTodo)
	context.IndentedJSON(http.StatusCreated, newTodo)
}


func getTodo(context *gin.Context) {
	id := context.Param("id")
	todo, err := getTodoById((id))

	if err != nil{
		context.IndentedJSON(http.StatusNotFound, gin.H{"message":"Todo not found"})
	}

	context.IndentedJSON(http.StatusOK, todo)
}

func toggleTodoStatus(context *gin.Context){
	id := context.Param("id")
	todo, err := getTodoById(id)

	if err!=nil {
		context.IndentedJSON(http.StatusNotFound, gin.H{"message": "Todo not found"})
	}

	todo.Completed = !todo.Completed

	context.IndentedJSON(http.StatusOK, todo)
}

func getTodoById(id string) (*todo, error){
	for i, t := range todos {
		if t.ID == id {
			return &todos[i], nil
		}
	}

	return nil, errors.New("Todo not found")
}

func main() {
	router := gin.Default()
	router.GET("/todos", getTodos)
	router.GET("todos/:id", getTodo)
	router.PATCH("todos/:id", toggleTodoStatus)
	router.POST("/todos", addTodo)
	router.Run("localhost:9090")
}