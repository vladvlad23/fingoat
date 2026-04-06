package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestParseNumeric(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		input  string
		wantOK bool
	}{
		{"valid integer", "100", true},
		{"valid decimal", "100.50", true},
		{"valid negative", "-50.25", true},
		{"valid zero", "0", true},
		{"valid large", "999999999.99", true},
		{"empty string", "", false},
		{"letters", "abc", false},
		{"mixed", "12abc", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, ok := parseNumeric(tt.input)
			if ok != tt.wantOK {
				t.Errorf("parseNumeric(%q) ok = %v, want %v", tt.input, ok, tt.wantOK)
			}
		})
	}
}

func TestParseDate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		input  string
		wantOK bool
	}{
		{"valid date", "2025-01-15", true},
		{"valid leap day", "2024-02-29", true},
		{"empty", "", false},
		{"wrong format", "01/15/2025", false},
		{"partial", "2025-01", false},
		{"garbage", "not-a-date", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, ok := parseDate(tt.input)
			if ok != tt.wantOK {
				t.Errorf("parseDate(%q) ok = %v, want %v", tt.input, ok, tt.wantOK)
			}
		})
	}
}

func TestParseID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		param   string
		want    int64
		wantErr bool
	}{
		{"valid", "42", 42, false},
		{"zero", "0", 0, false},
		{"negative", "-1", -1, false},
		{"not a number value", "xyz", 0, true},
		{"not a number", "abc", 0, true},
		{"float", "1.5", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create a chi router to set URL params
			r := chi.NewRouter()
			var gotID int64
			var gotErr error
			r.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
				gotID, gotErr = parseID(r)
			})

			req := httptest.NewRequest(http.MethodGet, "/"+tt.param, nil)
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			if tt.wantErr {
				if gotErr == nil {
					t.Errorf("parseID(%q) should return error", tt.param)
				}
			} else {
				if gotErr != nil {
					t.Errorf("parseID(%q) error: %v", tt.param, gotErr)
				}
				if gotID != tt.want {
					t.Errorf("parseID(%q) = %d, want %d", tt.param, gotID, tt.want)
				}
			}
		})
	}
}

func TestParseInt64(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input   string
		want    int64
		wantErr bool
	}{
		{"123", 123, false},
		{"0", 0, false},
		{"-5", -5, false},
		{"abc", 0, true},
		{"", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			got, err := parseInt64(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("parseInt64(%q) should return error", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("parseInt64(%q) error: %v", tt.input, err)
				}
				if got != tt.want {
					t.Errorf("parseInt64(%q) = %d, want %d", tt.input, got, tt.want)
				}
			}
		})
	}
}

func TestWriteJSON(t *testing.T) {
	t.Parallel()

	rr := httptest.NewRecorder()
	writeJSON(rr, http.StatusCreated, map[string]string{"key": "value"})

	if rr.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusCreated)
	}
	ct := rr.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("Content-Type = %q, want %q", ct, "application/json")
	}
	body := rr.Body.String()
	if body == "" {
		t.Error("body should not be empty")
	}
}

func TestWriteError(t *testing.T) {
	t.Parallel()

	rr := httptest.NewRecorder()
	writeError(rr, http.StatusBadRequest, "bad input")

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
	body := rr.Body.String()
	if body == "" {
		t.Error("body should not be empty")
	}
}
