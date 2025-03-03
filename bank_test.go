package main

import (
	"strings"
	"sync"
	"testing"
)

// Helper function to create a test Bank instance
func addAccount(bank *Bank, name string, balance int) {
	bank.mutex.Lock()
	defer bank.mutex.Unlock()
	bank.accounts[strings.ToLower(name)] = &Account { Name: name, Balance: balance }
}

// Test successful transfer
func TestSuccessfulTransfer(t *testing.T) {
	b := Bank{accounts: make(map[string]*Account)}
	addAccount(&b, "Mark", 100)
	addAccount(&b, "Jane", 50)

	req := TransferRequest{From: "Mark", To: "Jane", Amount: 30}
	resp, err := b.ProcessTransfer(req)

	if err != nil {
    t.Fatalf("Unexpected error: %s", err)
	}

	if resp.FromBalance != 70 || resp.ToBalance != 80 {
		t.Fatalf("Incorrect balances: Mark=%d, Jane=%d", resp.FromBalance, resp.ToBalance)
	}
}

func TestConcurrentTransfers(t *testing.T) {
	b := Bank{accounts: make(map[string]*Account)}
	addAccount(&b, "Mark", 100)
	addAccount(&b, "Jane", 50)
	addAccount(&b, "Adam", 0)

	wg := sync.WaitGroup{}

	transferFunc := func(from, to string, amount int) {
		defer wg.Done()
		_, _ = b.ProcessTransfer(TransferRequest{From: from, To: to, Amount: amount})
	}

	wg.Add(3)
	// simulate cyclic transfers
	go transferFunc("Mark", "Jane", 20)
	go transferFunc("Jane", "Mark", 20)

	go transferFunc("Mark", "Adam", 20)
	wg.Wait()

	mark, _ := b.GetAccount("Mark")
	jane, _ := b.GetAccount("Jane")
	adam, _ := b.GetAccount("Adam")

	if mark.Balance != 80 || jane.Balance != 50 || adam.Balance != 20 {
		t.Fatalf("Incorrect balances after concurrent transfers: Mark=%d, Jane=%d, Adam=%d", mark.Balance, jane.Balance, adam.Balance)
	}
}

func TestInsufficientFunds(t *testing.T) {
	b := Bank{accounts: make(map[string]*Account)}
	addAccount(&b, "Jane", 50)
	addAccount(&b, "Adam", 0)

	req := TransferRequest{From: "Jane", To: "Adam", Amount: 100} // More than Jane has
	_, err := b.ProcessTransfer(req)

	if err == nil {
		t.Fatal("Expected error for insufficient funds, got none")
	}

	if err.Error() != "insufficient funds" {
		t.Fatalf("Unexpected error message: %s", err)
	}
}

func TestSelfTransfer(t *testing.T) {
	b := Bank{accounts: make(map[string]*Account)}
	addAccount(&b, "Mark", 100)

	req := TransferRequest{From: "Mark", To: "Mark", Amount: 10}
	_, err := b.ProcessTransfer(req)

	if err == nil {
		t.Fatal("Expected error for self-transfer, got none")
	}

	if err.Error() != "cannot transfer to the same account" {
		t.Fatalf("Unexpected error message: %s", err)
	}
}

func TestInvalidSenderAccount(t *testing.T) {
	b := Bank{accounts: make(map[string]*Account)}
	addAccount(&b, "Mark", 100)

	req := TransferRequest{From: "Unknown", To: "Mark", Amount: 10}
	_, err := b.ProcessTransfer(req)

	if err == nil {
		t.Fatal("Expected error for invalid account, got none")
	}

	if err.Error() != "sender account does not exist" {
		t.Fatalf("Unexpected error message: %s", err)
	}
}

func TestInvalidReceiverAccount(t *testing.T) {
	b := Bank{accounts: make(map[string]*Account)}
	addAccount(&b, "Mark", 100)

	req := TransferRequest{From: "Mark", To: "Unknown", Amount: 10}
	_, err := b.ProcessTransfer(req)

	if err == nil {
		t.Fatal("Expected error for invalid account, got none")
	}

	if err.Error() != "receiver account does not exist" {
		t.Fatalf("Unexpected error message: %s", err)
	}
}

func TestZeroAmountTransfer(t *testing.T) {
	b := Bank{accounts: make(map[string]*Account)}
	addAccount(&b, "Mark", 100)
	addAccount(&b, "Jane", 50)

	req := TransferRequest{From: "Mark", To: "Jane", Amount: 0}
	_, err := b.ProcessTransfer(req)

	if err == nil {
		t.Fatal("Expected error for zero amount transfer, got none")
	}
	if err.Error() != "amount must be greater than zero" {
		t.Fatalf("Unexpected error message: %s", err)
	}
}
