package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"log"
	"strconv"
	"sync"
	"time"
	"github.com/joho/godotenv"
	"github.com/gorilla/mux"
)

type Node struct {
	URL      string
	Timeout  time.Duration
	IsEnable bool
}

type Result struct {
	URL    string `json:"url"`
	Status string `json:"status"`
	Body   string `json:"body,omitempty"`
}

func infoHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, os.Getenv("HUB"))
}

func handleQuery(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Querying with hub...")

	input := r.URL.Query().Get("input")
	if input == "" {
		http.Error(w, "Input query parameter is required", http.StatusBadRequest)
		return
	}

	fmt.Println("Input: ", input, "\n")

	nodes, err := readNodes("node.csv")
	if err != nil {
		http.Error(w, "Failed to read node.csv", http.StatusInternalServerError)
		return
	}

	results, err := fetchResults(nodes, input)
	if err != nil {
		http.Error(w, "Failed to fetch results", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func fetchResults(nodes []Node, input string) ([]Result, error) {
	var wg sync.WaitGroup
	out := make([]Result, 0, len(nodes))
	results := make(chan Result, len(nodes))

	for _, node := range nodes {
		if node.IsEnable {
			wg.Add(1)
			go func(n Node) {
				defer wg.Done()
				client := &http.Client{
					Timeout: n.Timeout,
				}

				resp, err := client.Get(n.URL + "?input=" + input)
				if err != nil {
					results <- Result{n.URL, fmt.Sprintf("Error: %s", err.Error()), ""}
					return
				}
				defer resp.Body.Close()

				// Read the response body
				bodyBytes, err := io.ReadAll(resp.Body)
				if err != nil {
					results <- Result{n.URL, resp.Status, fmt.Sprintf("Error reading body: %s", err.Error())}
					return
				}

				results <- Result{n.URL, resp.Status, string(bodyBytes)}
			}(node)
		}
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	for r := range results {
		out = append(out, r)
	}

	return out, nil
}

func readNodes(filename string) ([]Node, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	nodes := []Node{}
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		timeout, _ := strconv.Atoi(record[1])
		isEnable, _ := strconv.ParseBool(record[2])

		node := Node{
			URL:      record[0],
			Timeout:  time.Duration(timeout) * time.Second,
			IsEnable: isEnable,
		}
		nodes = append(nodes, node)
	}
	return nodes, nil
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
	r.HandleFunc("/query", handleQuery).Methods("GET")  // query?input=...

	port := os.Getenv("PORT")
	fmt.Println("Server started on port " + port + "\n")
	http.ListenAndServe(":" + port, r)  // for windows use 127.0.0.1:port
}