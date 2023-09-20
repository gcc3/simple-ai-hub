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
	"strings"
)

type Node struct {
	URL      string
	Timeout  time.Duration
	Type     string
	IsEnable bool
}

type QueryResponse struct {
	URL    string `json:"url"`
	Type   string `json:"type"`
	Status string `json:"status"`
	Result   string `json:"result,omitempty"`
}

type InnerQueryResponse struct {
	Result string `json:"result"`
}

func infoHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, os.Getenv("HUB"))
}

func queryHandler(w http.ResponseWriter, r *http.Request) {
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

	results, err := fetchSimpleResults(nodes, input)
	if err != nil {
		http.Error(w, "Failed to fetch results", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func fetchSimpleResults(nodes []Node, input string) (map[string]string, error) {
	var wg sync.WaitGroup
	results := make(chan QueryResponse, len(nodes))

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
					results <- QueryResponse{n.URL, n.Type, resp.Status, fmt.Sprintf("Error: %s", err.Error())}
					return
				}
				defer resp.Body.Close()

				// Read the response body
				bodyBytes, err := io.ReadAll(resp.Body)
				if err != nil {
					results <- QueryResponse{n.URL, n.Type, resp.Status, fmt.Sprintf("Error reading body: %s", err.Error())}
					return
				}

				// Parse the JSON response
				var parsedResponse InnerQueryResponse
				err = json.Unmarshal(bodyBytes, &parsedResponse)
				if err != nil {
					results <- QueryResponse{n.URL, n.Type, resp.Status, fmt.Sprintf("Error parsing JSON: %s", err.Error())}
					return
				}

				results <- QueryResponse{n.URL, n.Type, resp.Status, parsedResponse.Result}
			}(node)
		}
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	combinedResult := ""
	for r := range results {
		combinedResult += r.Result + "\n"
	}

	return map[string]string{"result": strings.TrimSpace(combinedResult)}, nil
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
		isEnable, _ := strconv.ParseBool(record[3])

		node := Node{
			URL:      record[0],
			Timeout:  time.Duration(timeout) * time.Second,
			Type:     record[2],
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
	r.HandleFunc("/query", queryHandler).Methods("GET")  // query?input=...

	port := os.Getenv("PORT")
	fmt.Println("Server started on port " + port + "\n")
	http.ListenAndServe(":" + port, r)  // for windows use 127.0.0.1:port
}