package services

import (
	"context"
	"encoding/json"
	"gophermart/internal/models"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func accuralHandler(w http.ResponseWriter, r *http.Request) {
	orderNumber := r.URL.Path[len("/api/orders/"):]
	response := models.OrderStatusNew{
		Number:     orderNumber,
		Status:     "PROCESSED",
		UploadedAt: time.Time{},
		Accrual:    5,
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")

	encoder := json.NewEncoder(w)
	err := encoder.Encode(response)
	if err != nil {

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

func TestGetOrder(t *testing.T) {

	handler := http.HandlerFunc(accuralHandler)
	server := httptest.NewServer(handler)
	defer server.Close()

	ctx := context.Background()
	in := make(chan string)
	out := make(chan models.OrderStatusNew)
	var wg sync.WaitGroup

	// Create an instance of the accrual struct
	a := &accrual{
		accrualSysremAdress: server.URL,
		logger:              &zap.SugaredLogger{},
	}

	// Start the worker goroutine
	wg.Add(1)
	go a.worker(ctx, in, out, &wg)

	// Send test data to the worker goroutine
	in <- "123"
	close(in)
	// Wait for the worker goroutine to finish
	// wg.Wait()

	// Check the output channel for the expected result
	order := <-out
	expResponse := models.OrderStatusNew{
		Number:     "123",
		Status:     "PROCESSED",
		UploadedAt: time.Time{},
		Accrual:    5,
	}

	assert.Equal(t, expResponse, order)

	wg.Wait()

}
