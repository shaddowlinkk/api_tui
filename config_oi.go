package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type endpoint struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Auth        bool     `json:"auth"`
	Endpoint    string   `json:"endpoint"`
	Method      string   `json:"method"`
	Keys        []string `json:"keys"`
}
type apiConfig struct {
	BaseUrl         string   `json:"baseUrl"`
	TokenURL        string   `json:"tokenURL"`
	TokenField      string   `json:"tokenField"`
	TokenType       string   `json:"tokenType"`
	HeaderKeys      []string `json:"headerKeys"`
	HeaderValues    []string `json:"headerValues"`
	AuthParamKeys   []string `json:"authParamKeys"`
	AuthParamValues []string `json:"authParamValues"`
	AuthDataType    string   `json:"authDataType"`
	AuthKeys        []string `json:"authKeys"`
	AuthValues      []string
	Endpoints       []endpoint `json:"endpoints"`
}

func read_config(filename string) apiConfig {
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println("Failed to read json api file")
	}
	var api apiConfig
	if err := json.Unmarshal(data, &api); err != nil {
		fmt.Println("Error in unmarshalling json data")
	}
	return api
}
func auth_request(authUrl string, headerkeys []string, headervalues []string, authkeys []string, authvalues []string, dataType string) ([]byte, error) {
	bodymap := make(map[string]string)
	if len(authvalues) == 0 {
		return nil, fmt.Errorf("nil auth values")
	}
	var r *http.Request
	if "json" == strings.ToLower(dataType) {
		for i, key := range authkeys {
			bodymap[key] = authvalues[i]
		}
		bodyBytes, err := json.Marshal(bodymap)
		if err != nil {
			return nil, err
		}
		r, err = http.NewRequest("POST", authUrl, bytes.NewBuffer(bodyBytes))
		if err != nil {
			return nil, err
		}
	} else if "param" == strings.ToLower(dataType) {
		data := url.Values{}
		for i, key := range authkeys {
			data.Add(key, authvalues[i])
		}
		var err error
		r, err = http.NewRequest("POST", authUrl, strings.NewReader(data.Encode()))
		if err != nil {
			return nil, err
		}
	}
	if r == nil {
		return nil, fmt.Errorf("err in auth request")
	}
	r.Header.Add("Content-Type", "application/json")
	for i, key := range headerkeys {
		r.Header.Add(key, headervalues[i])
	}
	return exec_request(r)
}
func exec_http(config apiConfig, idx_endpoint int, values []string) ([]byte, error) {
	bodymap := make(map[string]string)
	for i, key := range config.Endpoints[idx_endpoint].Keys {
		bodymap[key] = values[i]
	}
	bodyBytes, err := json.Marshal(bodymap)
	if err != nil {
		return nil, err
	}
	r, err := http.NewRequest(config.Endpoints[idx_endpoint].Method, config.BaseUrl+config.Endpoints[idx_endpoint].Endpoint, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}
	if config.Endpoints[idx_endpoint].Auth {
		data, err := auth_request(config.TokenURL, config.HeaderKeys, config.HeaderValues, config.AuthKeys, config.AuthValues, config.AuthDataType)
		if err != nil {
			return nil, err
		}
		var authdata map[string]string
		err = json.Unmarshal(data, &authdata)
		if err != nil {
			return nil, err
		}
		r.Header.Add("Authorization", config.TokenType+" "+authdata[config.TokenField])
	}
	r.Header.Add("Content-Type", "application/json")
	return exec_request(r)
}
func exec_request(r *http.Request) ([]byte, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	res, err := client.Do(r)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println("Error closeing HTTP")
		}
	}(res.Body)
	resBody, err := io.ReadAll(res.Body)
	if err != nil || len(resBody) == 0 {
		if len(resBody) == 0 {
			fmt.Printf("Status code:%d\n", res.StatusCode)
			return nil, fmt.Errorf("responce was empty from http")
		}
		return nil, err
	}
	if res.StatusCode != 200 {
		return nil, fmt.Errorf(string(resBody))
	}
	return resBody, nil
}
