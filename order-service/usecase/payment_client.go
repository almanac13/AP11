package usecase

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

type PaymentClient struct {
	baseURL string
	client  *http.Client
}

func NewPaymentClient(url string) *PaymentClient {
	return &PaymentClient{
		baseURL: url,
		client: &http.Client{
			Timeout: 2 * time.Second,
		},
	}
}

var ErrPaymentServiceUnavailable = errors.New("payment service unavailable")

func (p *PaymentClient) Pay(orderID string, amount int64) (string, error) {
	body := map[string]interface{}{
		"order_id": orderID,
		"amount":   amount,
	}

	jsonBody, _ := json.Marshal(body)

	resp, err := p.client.Post(p.baseURL+"/payments", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		// Payment service is down or timed out
		return "", ErrPaymentServiceUnavailable
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		return "", ErrPaymentServiceUnavailable
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("invalid response from payment service")
	}

	statusVal, ok := result["Status"]
	if !ok {
		statusVal, ok = result["status"]
	}
	if !ok {
		return "", fmt.Errorf("missing status in payment response")
	}

	status, ok := statusVal.(string)
	if !ok {
		return "", fmt.Errorf("invalid status type in payment response")
	}

	return status, nil
}
