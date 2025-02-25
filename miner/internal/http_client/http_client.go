package httpclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
	"net/url"
)

type BaseResponse struct {
	Status       string
	ErrorMessage string `json:"ErrorMessage,omitempty"`
}

type RequestWork struct {
	RewardAddress string
}

type ResponseWork struct {
	BaseResponse
	Block string
}

type RequestCompletedWork struct {
	ResponseWork
	BlockPOW int
}

type HttpClient struct {
	client *http.Client
}

func NewHttpCleint() *HttpClient {
	return &HttpClient{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (hc *HttpClient) SendCompletedWorkRequest(response *ResponseWork, pow int) error {
	request := &RequestCompletedWork{
		ResponseWork: *response,
		BlockPOW: pow,
	}

	// Преобразуем данные в JSON
	jsonData, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("Ошибка при маршалинге JSON: %v", err)
	}

	// Отправляем POST-запрос
	resp, err := hc.client.Post("http://localhost:80/api/v1/work/completed", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("Ошибка при отправке запроса: %v", err)
	}
	defer resp.Body.Close()

	log.Println(resp)
	return nil
}

func (hc *HttpClient) GiveWorkRequest() (*ResponseWork, error) {
	params := url.Values{}
	params.Add("query", "reward_address")

	//URL := fmt.Sprintf("http://localhost:80/api/v1/work?%s", params.Encode())
	URL := "http://localhost:80/api/v1/work?reward_address=Minnya"

	// Отправляем GET-запрос
	resp, err := hc.client.Get(URL)
	if err != nil {
		return nil, fmt.Errorf("Ошибка при отправке запроса: %v", err)
	}
	defer resp.Body.Close()

	// Декодируем JSON-ответ
	var data ResponseWork
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("Ошибка декодирования JSON: %v", err)
	}

	return &data, nil
}
