package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"io"
	"net/http"
	"os"
)

type OpenAIResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int      `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

func handler() (events.APIGatewayProxyResponse, error) {
	challenge, err := getPaintingChallenge()
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	return events.APIGatewayProxyResponse{
		Body:       challenge,
		StatusCode: 200,
	}, nil
}

func getPaintingChallenge() (string, error) {

	data := map[string]interface{}{
		"model":       "gpt-3.5-turbo",
		"temperature": 0.7,
		"messages": []map[string]string{
			{"role": "system", "content": "You are a poetic assistant, skilled in explaining complex programming concepts with creative flair."},
			{"role": "user", "content": "Compose a poem that explains the concept of recursion in programming."},
		},
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	request, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	request.Header.Set("Content-Type", "application/json")
	// Make sure to use your actual key here
	// You should not expose your secret keys in your code, use environment variables instead
	request.Header.Set("Authorization", "Bearer "+os.Getenv("OPENAI_KEY"))

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return "", err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(response.Body)

	body, _ := io.ReadAll(response.Body)

	// Unmarshalling the response
	var result OpenAIResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return "", err
	}

	if len(result.Choices) == 0 {
		return "", errors.New("no challenge generated")
	}

	return result.Choices[0].Message.Content, nil
}

func main() {
	lambda.Start(handler)
	//handler()
}
