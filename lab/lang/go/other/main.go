package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Age  int64  `json:"age"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

var users = []*User{
	{ID: "user_0001", Name: "Bob", Age: 28},
	{ID: "user_0003", Name: "Alice", Age: 18},
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /users", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(users); err != nil {
			log.Println("encode error:", err)
		}
	})
	mux.HandleFunc("GET /users/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		for _, u := range users {
			if u.ID == id {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				if err := json.NewEncoder(w).Encode(u); err != nil {
					log.Println("encode error:", err)
				}
				return
			}
		}

		w.WriteHeader(http.StatusNotFound)
		if err := json.NewEncoder(w).Encode(ErrorResponse{Error: "user not found"}); err != nil {
			log.Println("encode error:", err)
		}
	})

	fmt.Println("Start Server: 8080")

	http.ListenAndServe(":8080", mux)
}
