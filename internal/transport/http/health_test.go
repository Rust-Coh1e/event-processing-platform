package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthHandler(t *testing.T) {
	req := httptest.NewRequest(
		http.MethodGet,
		"/health",
		nil,
	)

	recorder := httptest.NewRecorder()

	HealthHandler(recorder, req)

	resp := recorder.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf(
			"expected 200, got %d",
			resp.StatusCode,
		)
	}

	if recorder.Header().Get("Content-Type") != "application/json" {
		t.Errorf(
			"wrong content type %s",
			recorder.Header().Get("Content-Type"),
		)
	}

	var reqBody struct {
		Status string `json:"status"`
	}

	err := json.NewDecoder(resp.Body).Decode(&reqBody)
	if err != nil {
		t.Errorf(
			"Error decoding JSON: %v",
			err,
		)
	}

	if reqBody.Status != "ok" {
		t.Errorf(
			"Expected 'ok' got : %s",
			reqBody.Status,
		)
	}
}
