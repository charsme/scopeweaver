# ScopeWeaver adoption — STATUS

Single source of truth for the security-hardening backlog before company adoption.
Fork: `charsme/scopeweaver` · upstream: `cskwork/scopeweaver` (derivative of `Autumn-27/ARTEX`).
Last updated: 2026-10-20 · Audited commit: `573344b` (fork == upstream main, 0 ahead / 0 behind).

## Gate

| Gate | Meaning | State |
|---|---|---|
| G0 — Pilot | One operator on `internal-tools` server, no company network reach | BLOCKED on P0 (4/7 coded on branch) |
| G1 — Team rollout | Multiple operators, real engagements | BLOCKED on P0 + P1 |
| G2 — Steady state | Repeatable update intake, governance in place | BLOCKED on P2 |

## Audit summary (2026-10-20)

- **Malicious code: none found** — Go backend, web/landing, scripts, CI, skills, and `github.com/Autumn-27/norma` v0.4.3 (agent core; hash matches `go.sum`).
- **Design risks found** — see backlog. Most important: every agent runs with `permission.ModeBypass`
  (`agent/mainagent.go:133`, `worker.go:431`, `planner.go:377`, `chat.go:148`, `goals.go:117`, `provider.go:572`),
  so norma's approval layer is off; the only gate is the DB intercept rules, which users can delete.
- **govulncheck**: 10 reachable vulns (net/http, x/net HTTP/2, crypto/tls, net/textproto). Fix: Go ≥ 1.26.9, `x/net` ≥ 0.60.0.
  Unreached: `x/crypto` 0.55.0 (ssh DoS, fixed 0.56.0), `klauspost/compress` 1.18.6 (fixed 1.18.7).
- **osv-scanner**: 10 npm vulns in `skills/api-recon/scripts/package-lock.json` (`ip-address`, `basic-ftp`, `extract-zip` — 4 High). `web/` lockfile clean.
- Not reviewed: other Go deps (`go-mitmproxy`, `projectdiscovery/*`), norma `noa/` internals line-by-line, no dynamic (runtime traffic) analysis.

## Progress (branch `security/adoption-hardening`, uncommitted)

Verified 2026-10-20 on go1.26.9: `go build`, `go vet` clean; full suite against throwaway Postgres 16 —
2016 pass / 0 fail / 4 skip (opt-in E2E/live suites); `govulncheck` 0 reachable (was 10).
Each new test was mutation-checked: it fails when its fix is reverted.

| ID | Change | Tests |
|---|---|---|
| SW-02 | `server/auth.go` `initAllowed`: init needs `SCOPEWEAVER_INIT_TOKEN` via `X-Init-Token`, else loopback only. Compose + `.env.example` pass the token. Docker users set the password with `curl` (UI init page cannot send the header). | `server/auth_init_gate_test.go`; existing DB init tests updated to loopback |
| SW-04 | `guard.SecretEnvOverrides` blanks `*KEY*/*TOKEN*/*SECRET*/*PASSWORD*/*DSN*/*AUTH*` vars for agent Bash (all 4 agents via `bashEnv`), custom command/script tools, stdio MCP (configured server env still wins). | `guard/secretenv_test.go`, `agent/proxyenv_test.go`, `server/customtool_test.go` (real child processes). stdio-MCP path: no test |
| SW-06 | `selfupdate.Repo` → `charsme/scopeweaver` (no releases yet ⇒ updater reports none) | `selfupdate/repo_test.go` guards against upstream merge reverting it |
| SW-07 | `go 1.26.9`, `Dockerfile` `golang:1.26.9-bookworm`, `x/net` 0.60.0, `x/crypto` 0.57.0, `klauspost/compress` 1.18.7 (+ `x/sys`, `x/text`, `x/sync`, `x/mod` minor bumps via tidy) | govulncheck |

**SW-04 limitation:** blanking child env does not stop a same-UID command reading the backend's
`/proc/<pid>/environ`. Full fix needs SW-03 (isolation) plus either a separate UID for agent commands
or secrets delivered outside the process environment (see SW-20).

## Backlog

### P0 — before G0 pilot

| ID | Item | Evidence | Done when |
|---|---|---|---|
| SW-01 | Bind UI to loopback; reach it only via VPN or reverse proxy with SSO | `cmd/artex/main.go:55` (`:8787`), `docker-compose.yml:49` | `curl` to `:8787` from another host fails |
| SW-02 | Close first-visitor admin claim: set password before any exposure, or patch `authInit` to require a bootstrap token from env | `server/auth.go:116,148` | Unauthenticated `POST /api/auth/init` on fresh install rejected without token |
| SW-03 | Dedicated isolated host (VM/VPC), no company creds, egress firewall | — | Host runbook exists; no route to corp/prod networks |
| SW-04 | Strip secrets from agent-executed processes (Bash, custom script tools, MCP stdio inherit full env) | `server/customtool.go:421`; norma `tool/task.go:139`, `mcp/stdio.go:33` | Agent running `env` sees no `*_API_KEY`, `*_DSN`, passwords |
| SW-05 | LLM key hygiene: dedicated key with spend cap, or route via gateway holding the real key | — | Leaked key in agent env is low-value and capped |
| SW-06 | Self-update: disable, or repoint `selfupdate.Repo` to fork; updates only via reviewed `git pull` | `selfupdate/github.go:16` | No code path fetches from `cskwork/*` |
| SW-07 | Rebuild on patched toolchain + deps (Go ≥ 1.26.9, `x/net` ≥ 0.60.0, `x/crypto` ≥ 0.56.0, `klauspost/compress` ≥ 1.18.7); bump `go.mod` + `Dockerfile:11` | govulncheck output | `govulncheck ./...` exit 0 |

### P1 — before G1 team rollout

| ID | Item | Evidence | Done when |
|---|---|---|---|
| SW-08 | Scope enforcement: per-engagement egress allowlist at network layer (code RoE scope was removed) | `guard/guard.go:1-7` | Agent cannot reach a host outside the engagement allowlist |
| SW-09 | Lock built-in intercept rules (non-deletable or config-managed) since `ModeBypass` leaves them as sole gate | `guard/guard.go`, `agent/*.go` `ModeBypass` | Operator cannot silently disable destructive/exfil rules |
| SW-10 | Decide on agent self-tooling: disable `create_custom_tool` / `create_mcp` by default or require human approval | `server/platform_tools.go:19-36,198-220` | Agent cannot create host-executed tools unprompted |
| SW-11 | API hardening: restrict CORS origin, `SameSite=Strict` cookie, `?token=` only for SSE route | `server/server.go:3964`, `server/auth.go:100-110` | Cross-origin request from other site rejected |
| SW-12 | api-recon skill: regenerate lockfile against registry.npmjs.org, bump `ip-address`/`basic-ftp`; resolve `extract-zip` (no fix) or don't install skill | `skills/api-recon/scripts/package-lock.json` | osv-scanner clean; no `npmmirror.com` URLs |
| SW-13 | Build supply chain: digest-pin base images, drop `curl \| bash` NodeSource, pin `@playwright/*`, pin `ollama` tag, SHA-pin GitHub Actions | `Dockerfile:2,13,22,25,28-30`, `docker-compose.yml`, `.github/workflows/` | No mutable refs in build path |
| SW-14 | Bare-metal hosts: set `NORMA_RIPGREP_NO_INSTALL=1` (norma silently runs `npm install -g @vscode/ripgrep` if `rg` missing) | norma `tool/ripgrep.go:51-66` | Env set in install runbook (Docker image already ships `ripgrep`) |

### P2 — before G2 steady state

| ID | Item | Done when |
|---|---|---|
| SW-15 | Engagement governance: written authorization, target list, legal sign-off, audit-log retention | Template + approval flow exists |
| SW-16 | Upstream-sync intake: each upstream merge → diff review + govulncheck + osv-scanner in CI | CI job blocks on new vulns |
| SW-17 | Cover remaining audit gaps: `go-mitmproxy`, `projectdiscovery/*`, norma `noa/`; dynamic run in sandbox with egress capture | Gap list empty or risk-accepted |
| SW-18 | Report upstream: CORS `*`, first-visitor admin, env inheritance; norma `matchRule` ignores args (`permission.go:139-151`), read-only tools bypass `Ask` (`:111`), `ExecuteExtraTool` skips permission checks (`tool/deferred.go:221-260`) | Issues filed upstream |

### Found during fixes

| ID | Item | Evidence | Done when |
|---|---|---|---|
| SW-19 | LLM connection test runs agentcore with no `Tools` ⇒ norma `DefaultTools()` (incl. Bash) in bypass mode; a malicious or compromised LLM endpoint can answer the ping with a command | `agent/provider.go:572`, norma `agentcore/agentcore.go:250` | Connection test runs with an explicit empty tool set |
| SW-20 | Agent commands run as the backend's UID and can read `/proc/<pid>/environ` | — | Agent commands run under a separate UID, or backend secrets loaded from files/DB instead of env |
| SW-21 | Web UI init page cannot send `X-Init-Token`; Docker first-run needs `curl` | `web/src/lib/api.ts:232` | Optional token field on init page, or documented runbook only |

## Decisions

- 2026-10-20 — Pilot runs on the `internal-tools` server.
- 2026-10-20 — Fixes live on fork `charsme/scopeweaver`; self-update follows the fork.
