package main

import (
	"bytes"
	"fmt"
	"net/http"
	"os/exec"
	"os"
	"log"
	"github.com/joho/godotenv"
	"github.com/gorilla/mux"
)

// Info handler
func infoHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, os.Getenv("NODE"))
}

type Response struct {
	Message string `json:"result"`
	Error   string `json:"error,omitempty"`
}

// Default query handler
// Example: { "user_input": "What is the capital of France?" }
func generateHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Query engine: ", os.Getenv("QUERY_ENGINE"))  // text, db, browsing

	// Input (user_input)
	input := r.URL.Query().Get("user_input")
	if input == "" {
		http.Error(w, "Input query parameter is required", http.StatusBadRequest)
		return
	}
	fmt.Println("Input:", input, "\n")

	// Run python script
	cmd := exec.Command(os.Getenv("PYTHON_PATH"), "py_exec.py", input)
	cmd.Env = append(os.Environ(), "PYTHONIOENCODING=utf-8")  // avoid encoding error

	// Debug python
	if (os.Getenv("DEBUG") == "true") {
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		err := cmd.Run()
		if err != nil {
			http.Error(w, "Failed to run python script: " + err.Error() + "\n\n" + stderr.String(), http.StatusInternalServerError)
			return
		}
	    fmt.Fprintf(w, "No error")
		return
	}

	// Output
	output, err := cmd.Output()
	if err != nil {
		http.Error(w, "Error: " + err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Println("Output: ", string(output))

	// Write response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(output)); err != nil {
        // Handle the error if there's one
        http.Error(w, "Internal Server Error while writing response", http.StatusInternalServerError)
    }
}

func main() {
	// load env
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// routes
	r := mux.NewRouter()
	r.HandleFunc("/", infoHandler).Methods("GET")
	r.HandleFunc("/generate", generateHandler).Methods("GET")

	port := os.Getenv("PORT")
	fmt.Println("Server started on port " + port + "\n")
	http.ListenAndServe(":" + port, r)  // for windows use 127.0.0.1:port
}
