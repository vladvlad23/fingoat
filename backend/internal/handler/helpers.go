package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func parseID(r *http.Request) (int64, error) {
	return strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
}

func parseInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

func parseNumeric(s string) (pgtype.Numeric, bool) {
	var n pgtype.Numeric
	if err := n.Scan(s); err != nil {
		return n, false
	}
	return n, true
}

func parseDate(s string) (pgtype.Date, bool) {
	var d pgtype.Date
	if err := d.Scan(s); err != nil {
		return d, false
	}
	return d, true
}

func isValidCurrency(c string) bool {
	return c == "USD" || c == "EUR" || c == "RON"
}
