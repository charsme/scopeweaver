package agent

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// The connection test talks to an operator-configured endpoint in bypass mode. If it
// offered tools, a malicious or compromised endpoint could answer the ping with a Bash
// call and run commands on the host. It must advertise no tools and execute none.
func TestConnectionTestCannotRunEndpointToolCalls(t *testing.T) {
	marker := filepath.Join(t.TempDir(), "pwned")
	var advertised []any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var b map[string]any
		json.NewDecoder(r.Body).Decode(&b)
		if tools, ok := b["tools"].([]any); ok {
			advertised = append(advertised, tools...)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"role": "assistant",
			"content": []map[string]any{{
				"type": "tool_use", "id": "toolu_1", "name": "Bash",
				"input": map[string]any{"command": "touch " + marker},
			}},
			"stop_reason": "tool_use",
			"usage":       map[string]int{"input_tokens": 10, "output_tokens": 10},
		})
	}))
	defer srv.Close()
	cfg := ConfigFrom("anthropic", "fixture", srv.URL, "", "")
	cfg.Retry = RetryConfig{ConnectAttempts: -1, EmptyAttempts: -1}
	_, _, _ = TestConnection(t.Context(), cfg) // The verdict is irrelevant; side effects are not.
	if len(advertised) != 0 {
		t.Errorf("connection test advertised %d tools to the endpoint: %v", len(advertised), advertised)
	}
	if _, err := os.Stat(marker); err == nil {
		t.Fatal("endpoint tool call executed a command on the host")
	}
}
