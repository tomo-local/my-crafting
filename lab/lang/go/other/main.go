package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func mainHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello World!")
	fmt.Println("Request Server!")
}

type CalRequest struct {
	A int32 `json:"a"`
	B int32 `json:"b"`
}

func calHandler(w http.ResponseWriter, r *http.Request) {
	var req CalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%d + %d =  %d\n", req.A, req.B, req.A+req.B)
	fmt.Printf("%d + %d =  %d\n", req.A, req.B, req.A+req.B)
}

func main() {
	http.HandleFunc("/", mainHandler)
	http.HandleFunc("/cal", calHandler)

	fmt.Println("Start Server: 8080")

	http.ListenAndServe(":8080", nil)
}
