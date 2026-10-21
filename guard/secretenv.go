package guard

import "strings"

// secretNameMarkers flag environment variable names that carry credentials: LLM API
// keys, the Postgres DSN, the bootstrap token, cloud keys, and similar.
var secretNameMarkers = []string{"KEY", "TOKEN", "SECRET", "PASSWORD", "PASSWD", "CREDENTIAL", "DSN", "DATABASE_URL", "AUTH"}

// SecretEnvOverrides returns "NAME=" entries that blank every secret-looking variable in
// environ (os.Environ() form). Commands the model runs inherit the backend environment;
// appending these after it makes os/exec keep the empty value (last duplicate wins), so a
// prompt-injected command cannot read the backend's LLM keys or database DSN. Values a
// caller sets explicitly for a tool must be appended after these to take effect.
func SecretEnvOverrides(environ []string) []string {
	var out []string
	for _, kv := range environ {
		name, _, ok := strings.Cut(kv, "=")
		if !ok || name == "" {
			continue
		}
		upper := strings.ToUpper(name)
		for _, m := range secretNameMarkers {
			if strings.Contains(upper, m) {
				out = append(out, name+"=")
				break
			}
		}
	}
	return out
}
