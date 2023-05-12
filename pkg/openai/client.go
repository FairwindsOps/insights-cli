package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Request struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type Response struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int    `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Status string `json:"status"`
}

func sendRequest(ctx context.Context, apiKey string, request Request) (string, error) {
	apiURL := "https://api.openai.com/v1/chat/completions"
	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", err
	}

	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	reqDone := make(chan bool)

	go func() {
		for {
			select {
			case <-reqDone:
				return
			default:
				fmt.Print(".")
				time.Sleep(2 * time.Second)
			}
		}
	}()

	resp, err := client.Do(req)
	reqDone <- true
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if !isSuccessful(resp.StatusCode) {
		return "", fmt.Errorf("expected 200 - got %d - %s", resp.StatusCode, string(body))
	}

	var response Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", err
	}
	if len(response.Choices) == 0 {
		fmt.Println("Bad response:\n" + string(body))
		return "", errors.New("Bad response from OpenAI")
	}

	return response.Choices[0].Message.Content, nil
}

func isSuccessful(statusCode int) bool {
	return statusCode >= 200 && statusCode < 400
}
