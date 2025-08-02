// Package challenge7 contains the solution for Challenge 7: Bank Account with Error Handling.
package challenge7

import (
	"fmt"
	"sync"
	// Add any other necessary imports
)

// BankAccount represents a bank account with balance management and minimum balance requirements.
type BankAccount struct {
	ID         string
	Owner      string
	Balance    float64
	MinBalance float64
	mu         sync.Mutex // For thread safety
}

// Constants for account operations
const (
	MaxTransactionAmount = 10000.0 // Example limit for deposits/withdrawals
)

// Custom error types

// AccountError is a general error type for bank account operations.
type AccountError struct {
	argName string
}

func (e *AccountError) Error() string {
	return fmt.Sprintf("%s can't be empty", e.argName)
}

// InsufficientFundsError occurs when a withdrawal or transfer would bring the balance below minimum.
type InsufficientFundsError struct {
	lesserArgName, greaterArgName string
	lesserArg, greaterArg         float64
}

func (e *InsufficientFundsError) Error() string {
	return fmt.Sprintf("InsufficientFundsError: %s must be greater or equal to %s, got %f < %f", e.lesserArgName, e.greaterArgName, e.lesserArg, e.greaterArg)
}

// NegativeAmountError occurs when an amount for deposit, withdrawal, or transfer is negative.
type NegativeAmountError struct {
	argName  string
	argValue float64
}

func (e *NegativeAmountError) Error() string {
	return fmt.Sprintf("NegativeAmountError: %s must be greater or equal to zero, got %f", e.argName, e.argValue)
}

// ExceedsLimitError occurs when a deposit or withdrawal amount exceeds the defined limit.
type ExceedsLimitError struct {
	arg float64
}

func (e *ExceedsLimitError) Error() string {
	return fmt.Sprintf("ExceedsLimitError: %f greater than MaxTransactionAmount", e.arg)
}

// NewBankAccount creates a new bank account with the given parameters.
// It returns an error if any of the parameters are invalid.
func NewBankAccount(id, owner string, initialBalance, minBalance float64) (*BankAccount, error) {
	if id == "" {
		return nil, &AccountError{argName: "id"}
	}

	if owner == "" {
		return nil, &AccountError{argName: "owner"}
	}

	if initialBalance < 0 {
		return nil, &NegativeAmountError{
			argName:  "initial balance",
			argValue: initialBalance,
		}
	}

	if minBalance < 0 {
		return nil, &NegativeAmountError{
			argName:  "minimal balance",
			argValue: minBalance,
		}
	}

	if initialBalance < minBalance {
		return nil, &InsufficientFundsError{
			lesserArgName:  "initial balance",
			greaterArgName: "minimal balance",
			lesserArg:      initialBalance,
			greaterArg:     minBalance,
		}
	}

	return &BankAccount{
		ID:         id,
		Owner:      owner,
		Balance:    initialBalance,
		MinBalance: minBalance,
	}, nil
}

// Deposit adds the specified amount to the account balance.
// It returns an error if the amount is invalid or exceeds the transaction limit.
func (a *BankAccount) Deposit(amount float64) error {
	if amount < 0 {
		return &NegativeAmountError{
			argName:  "amount",
			argValue: amount,
		}
	}

	if amount > MaxTransactionAmount {
		return &ExceedsLimitError{
			arg: amount,
		}
	}

	a.mu.Lock()
	a.Balance += amount
	a.mu.Unlock()

	return nil
}

// Withdraw removes the specified amount from the account balance.
// It returns an error if the amount is invalid, exceeds the transaction limit,
// or would bring the balance below the minimum required balance.
func (a *BankAccount) Withdraw(amount float64) error {
	if amount < 0 {
		return &NegativeAmountError{
			argName:  "amount",
			argValue: amount,
		}
	}

	if amount > MaxTransactionAmount {
		return &ExceedsLimitError{
			arg: amount,
		}
	}

	if a.Balance-amount < a.MinBalance {
		return &InsufficientFundsError{
			lesserArgName:  "balance - amount",
			greaterArgName: "minimal balance",
			lesserArg:      a.Balance - amount,
			greaterArg:     a.MinBalance,
		}
	}

	a.mu.Lock()
	a.Balance -= amount
	a.mu.Unlock()

	return nil
}

// Transfer moves the specified amount from this account to the target account.
// It returns an error if the amount is invalid, exceeds the transaction limit,
// or would bring the balance below the minimum required balance.
func (a *BankAccount) Transfer(amount float64, target *BankAccount) error {
	if amount < 0 {
		return &NegativeAmountError{
			argName:  "amount",
			argValue: amount,
		}
	}

	if amount > MaxTransactionAmount {
		return &ExceedsLimitError{
			arg: amount,
		}
	}

	if a.Balance-amount < a.MinBalance {
		return &InsufficientFundsError{
			lesserArgName:  "balance - amount",
			greaterArgName: "minimal balance",
			lesserArg:      a.Balance - amount,
			greaterArg:     a.MinBalance,
		}
	}

	a.mu.Lock()
	a.Balance -= amount
	a.mu.Unlock()

	target.mu.Lock()
	target.Balance+=amount
	target.mu.Unlock()

	return nil
}
