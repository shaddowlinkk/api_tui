package api_tui

import (
	"encoding/json"
	"fmt"
	"os"
)

type endpoint struct {
	Fields   int      `json:"fields"`
	Endpoint string   `json:"endpoint"`
	Keys     []string `json:"keys"`
}

func read_config(filename string) []endpoint {
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println("Failed to read json api file")
	}
	var api []endpoint
	if err := json.Unmarshal(data, &api); err != nil {
		fmt.Println("Error in unmarshalling json data")
	}
	return api
}
