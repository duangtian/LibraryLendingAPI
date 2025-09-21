package validation

import (
	"encoding/json"
	"net/http"
	"time"
)

// Problem per RFC 9457 (application/problem+json)
type Problem struct {
	Type   string `json:"type"`
	Title  string `json:"title"`
	Status int    `json:"status"`
	Detail string `json:"detail,omitempty"`
	Instance string `json:"instance,omitempty"`
	InvalidParams []InvalidParam `json:"invalidParams,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

type InvalidParam struct {
	Name string `json:"name"`
	Reason string `json:"reason"`
}

func Write(w http.ResponseWriter, status int, title, detail string, invalid []InvalidParam) {
	p := Problem{
		Type: "about:blank",
		Title: title,
		Status: status,
		Detail: detail,
		InvalidParams: invalid,
		Timestamp: time.Now(),
	}
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(p)
}
