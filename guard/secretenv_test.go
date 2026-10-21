package guard

import (
	"slices"
	"testing"
)

// Commands the model runs inherit the backend environment. The LLM keys, Postgres DSN
// and bootstrap token must be blanked so prompt-injected commands cannot exfiltrate
// them, while routing and locale variables tools need must pass through untouched.
func TestSecretEnvOverridesBlanksCredentialsOnly(t *testing.T) {
	environ := []string{
		"ANTHROPIC_API_KEY=sk-ant-live",
		"OPENAI_API_KEY=sk-live",
		"SCOPEWEAVER_LLM_API_KEY=k",
		"ARTEX_PG_DSN=postgres://artex:pw@postgres:5432/artex",
		"SCOPEWEAVER_INIT_TOKEN=boot",
		"AWS_SECRET_ACCESS_KEY=aws",
		"GITHUB_TOKEN=ghp",
		"POSTGRES_PASSWORD=pw",
		"PATH=/usr/bin",
		"HOME=/root",
		"HTTPS_PROXY=http://127.0.0.1:8788",
		"SSL_CERT_FILE=/data/ca.pem",
		"LANG=C.UTF-8",
		"malformed",
	}
	got := SecretEnvOverrides(environ)
	for _, want := range []string{
		"ANTHROPIC_API_KEY=", "OPENAI_API_KEY=", "SCOPEWEAVER_LLM_API_KEY=", "ARTEX_PG_DSN=",
		"SCOPEWEAVER_INIT_TOKEN=", "AWS_SECRET_ACCESS_KEY=", "GITHUB_TOKEN=", "POSTGRES_PASSWORD=",
	} {
		if !slices.Contains(got, want) {
			t.Errorf("secret not blanked: missing %q in %v", want, got)
		}
	}
	for _, kept := range []string{"PATH", "HOME", "HTTPS_PROXY", "SSL_CERT_FILE", "LANG"} {
		for _, kv := range got {
			if kv == kept+"=" {
				t.Errorf("non-secret %s blanked; tools need it", kept)
			}
		}
	}
}
