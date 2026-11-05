package service

import (
	"bytes"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type MidtransService interface {
	CreateSnapTransaction(orderID string, amount int64, customerName, customerEmail, itemName string) (string, string, error)
	VerifySignature(orderID, statusCode, grossAmount, signatureKey string) bool
}

type midtransService struct {
	client *http.Client
}

func NewMidtransService() MidtransService {
	return &midtransService{
		client: &http.Client{Timeout: 15 * time.Second},
	}
}

func (s *midtransService) getServerKey() string {
	return os.Getenv("MIDTRANS_SERVER_KEY")
}

func (s *midtransService) getSnapBaseURL() string {
	if os.Getenv("MIDTRANS_IS_PRODUCTION") == "true" {
		return "https://app.midtrans.com/snap/v1/transactions"
	}
	return "https://app.sandbox.midtrans.com/snap/v1/transactions"
}

func (s *midtransService) CreateSnapTransaction(orderID string, amount int64, customerName, customerEmail, itemName string) (string, string, error) {
	serverKey := s.getServerKey()

	safeItemName := strings.TrimSpace(itemName)
	if len(safeItemName) > 45 {
		safeItemName = safeItemName[:45]
	}
	if safeItemName == "" {
		safeItemName = "Layanan WargaPay"
	}

	safeCustomerName := strings.TrimSpace(customerName)
	if safeCustomerName == "" {
		safeCustomerName = "Warga"
	}
	if len(safeCustomerName) > 45 {
		safeCustomerName = safeCustomerName[:45]
	}

	safeCustomerEmail := strings.TrimSpace(customerEmail)
	if safeCustomerEmail == "" {
		safeCustomerEmail = "warga@wargapay.id"
	}

	itemID := orderID
	if len(itemID) > 45 {
		itemID = itemID[:45]
	}

	reqBody := map[string]interface{}{
		"transaction_details": map[string]interface{}{
			"order_id":     orderID,
			"gross_amount": amount,
		},
		"customer_details": map[string]interface{}{
			"first_name": safeCustomerName,
			"email":      safeCustomerEmail,
		},
		"item_details": []map[string]interface{}{
			{
				"id":       itemID,
				"price":    amount,
				"quantity": 1,
				"name":     safeItemName,
			},
		},
	}

	payloadBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", "", err
	}

	req, err := http.NewRequest(http.MethodPost, s.getSnapBaseURL(), bytes.NewBuffer(payloadBytes))
	if err != nil {
		return "", "", err
	}

	authString := base64.StdEncoding.EncodeToString([]byte(serverKey + ":"))
	req.Header.Set("Authorization", "Basic "+authString)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	res, err := s.client.Do(req)
	if err != nil {
		log.Printf("[MIDTRANS ERROR] Gagal menghubungi Midtrans Snap: %v", err)
		return "", "", fmt.Errorf("gagal menghubungi midtrans: %w", err)
	}
	defer res.Body.Close()

	var snapRes struct {
		Token         string   `json:"token"`
		RedirectURL   string   `json:"redirect_url"`
		ErrorMessages []string `json:"error_messages"`
	}

	if err := json.NewDecoder(res.Body).Decode(&snapRes); err != nil {
		return "", "", err
	}

	if res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusOK {
		log.Printf("[MIDTRANS ERROR] Response status %d: %v", res.StatusCode, snapRes.ErrorMessages)
		return "", "", fmt.Errorf("midtrans error: %v", snapRes.ErrorMessages)
	}

	if snapRes.Token == "" {
		return "", "", errors.New("midtrans snap token kosong")
	}

	return snapRes.Token, snapRes.RedirectURL, nil
}

func (s *midtransService) VerifySignature(orderID, statusCode, grossAmount, signatureKey string) bool {
	serverKey := s.getServerKey()
	rawString := orderID + statusCode + grossAmount + serverKey

	hash := sha512.Sum512([]byte(rawString))
	calculated := hex.EncodeToString(hash[:])

	return calculated == signatureKey
}

