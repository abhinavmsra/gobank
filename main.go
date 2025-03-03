package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
)

type Account struct {
	Name    string `json:"name"`
	Balance int    `json:"balance"`
	Mutex   sync.Mutex
}

type Bank struct {
	accounts map[string]*Account
	mutex    sync.RWMutex
}

type TransferRequest struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Amount int    `json:"amount"`
}

type TransferResponse struct {
	From   string `json:"from"`
	FromBalance int `json:"from_balance"`
	To     string `json:"to"`
	ToBalance int `json:"to_balance"`
	Message string `json:"message"`
}

var bank = Bank{accounts: make(map[string]*Account)}

const APP_PORT = 3000

func (b *Bank) LoadAccounts(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	var accounts []Account
	if err := json.NewDecoder(file).Decode(&accounts); err != nil {
		return err
	}

	b.mutex.Lock()
	defer b.mutex.Unlock()
	for _, acc := range accounts {
		b.accounts[strings.ToLower(acc.Name)] = &Account{Name: acc.Name, Balance: acc.Balance}
	}
	return nil
}

func (b *Bank) GetAccount(name string) (*Account, bool) {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	acc, exists := b.accounts[strings.ToLower(name)]
	return acc, exists
}

func validateTransferRequest(b *Bank, req *TransferRequest) error {
	req.From = strings.ToLower(req.From)
	req.To = strings.ToLower(req.To)

	if req.From == req.To {
		return errors.New("cannot transfer to the same account")
	}

	if req.Amount <= 0 {
		return errors.New("amount must be greater than zero")
	}

	if _, ok := b.GetAccount(req.From); !ok {
		return errors.New("sender account does not exist")
	}

	if _, ok := b.GetAccount(req.To); !ok {
		return errors.New("receiver account does not exist")
	}

	return nil
}

func (b *Bank) ProcessTransfer(req TransferRequest) (TransferResponse, error) {
	if err := validateTransferRequest(b, &req); err != nil {
		return TransferResponse{}, err
	}

	sender, _ := b.GetAccount(req.From)
	receiver, _ := b.GetAccount(req.To)

	// Lock accounts in consistent order to prevent deadlocks
	if req.From < req.To {
		sender.Mutex.Lock()
		receiver.Mutex.Lock()
	} else {
		receiver.Mutex.Lock()
		sender.Mutex.Lock()
	}
	defer sender.Mutex.Unlock()
	defer receiver.Mutex.Unlock()

	if sender.Balance < req.Amount {
		return TransferResponse{}, errors.New("insufficient funds")
	}

	sender.Balance -= req.Amount
	receiver.Balance += req.Amount

	response := TransferResponse{
		From:        sender.Name,
		FromBalance: sender.Balance,
		To:          receiver.Name,
		ToBalance:   receiver.Balance,
		Message:     "Transfer successful",
	}

	return response, nil
}

func (b *Bank) TransferMoneyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusNotFound)
		return
	}

	var req TransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	response, err := bank.ProcessTransfer(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func main() {
	if err := bank.LoadAccounts("accounts.json"); err != nil {
		fmt.Println("Error loading accounts:", err)
		return
	}

	http.HandleFunc("/transfer", bank.TransferMoneyHandler)
	fmt.Printf("Server running on port %d", APP_PORT)
	http.ListenAndServe(fmt.Sprintf(":%d", APP_PORT), nil)
}
