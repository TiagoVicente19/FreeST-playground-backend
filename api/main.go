package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/google/uuid"
)

func main() {

	mux := http.NewServeMux()

	mux.HandleFunc("POST /runFreeST", postCode)

	log.Fatal(http.ListenAndServe(":8080", enableCORS(mux)))
}

// Created to avoid the Same Origin Policy disallowing the backend and frontend to connect. In later stages could possibly de deleted, if not, must be changed to connect to the correct address.
func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func postCode(w http.ResponseWriter, r *http.Request) {

	var newCode string

	if err := json.NewDecoder(r.Body).Decode(&newCode); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	out := handleFreest(newCode)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(out)
}

func handleFreest(newCode string) string {
	myData := []byte(newCode)

	fileName, err := createFile(myData)
	if err != nil {
		return "Error: could not create file."
	}

	ctx, cancel := handleTimeout()
	defer cancel()

	cmd := exec.CommandContext(ctx, "freest", fileName)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	var output string
	result, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			output = "Error: execution timed out."
		} else {
			output = cleanString(stderr)
		}
	} else {
		output = string(result)
	}

	os.Remove(fileName)
	return output
}

func handleTimeout() (context.Context, context.CancelFunc) {
	timeout := 10
	return context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
}

func cleanString(stderr bytes.Buffer) string {
	newOutput := stderr.String()
	idx := strings.Index(newOutput, "error")
	if idx != -1 {
		newOutput = newOutput[idx:]
	}
	idx = strings.Index(newOutput, "Failed, no modules loaded.")
	if idx != -1 {
		newOutput = newOutput[:idx]
	}
	newOutput = strings.TrimRight(newOutput, "\n")
	return newOutput
}

func createFile(myData []byte) (string, error) {

	fileName := (uuid.New().String()) + ".fst"

	erro := os.WriteFile(fileName, myData, 0644)
	if erro != nil {
		return "", erro
	}

	return fileName, nil

}
