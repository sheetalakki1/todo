package router

import (
	"go_todo/repository"
	"net/http"

	"github.com/gorilla/mux"
)

func Router() {
	r := mux.NewRouter()

	r.HandleFunc("/todos", repository.FetchTodos).Methods("GET", "OPTIONS")
	r.HandleFunc("/todos", repository.CreateTodo).Methods("POST", "OPTIONS")
	r.HandleFunc("/todos/{id}", repository.UpdateTodo).Methods("PUT", "OPTIONS")
	r.HandleFunc("/todos/{id}", repository.DeleteTodo).Methods("DELETE", "OPTIONS")
	r.HandleFunc("/todos/{id}", repository.GetTodoById).Methods("GET", "OPTIONS")
	http.Handle("/", r)
}
