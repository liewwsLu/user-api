package main

import (
	"errors"
	"fmt"
	"user-api/internal/storage"
)

func printError(err error) {
	status := storage.StatusByError(err)

	if err == nil {
		fmt.Println("OK", status)
		return
	}

	if errors.Is(err, storage.ErrValidation) {
		fmt.Printf("Bad request | status=%d | error=%v\n", status, err)
		return
	}

	if errors.Is(err, storage.ErrNotFound) {
		fmt.Printf("Not found | status=%d | error=%v\n", status, err)
		return
	}

	if errors.Is(err, storage.ErrConflict) {
		fmt.Printf("Conflict | status=%d | error=%v\n", status, err)
		return
	}

	fmt.Printf("Unknown error | status=%d | error=%v\n", status, err)
}
