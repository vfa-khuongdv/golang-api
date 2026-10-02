#!/usr/bin/env bash
# Smoke test against a real MySQL database: start the server (which applies the
# migrations), seed it, and call the API end to end. The unit and e2e tests run
# on SQLite, so this is what catches MySQL-only problems (migration SQL, row
# locks, UPDATE semantics).
#
# Needs DB_*, JWT_KEY and SETTINGS_ENCRYPTION_KEY in the environment and an
# empty database. Usage: scripts/smoke-test.sh
set -euo pipefail

PORT="${PORT:-3000}"
BASE_URL="http://127.0.0.1:${PORT}/api/v1"
LOG_FILE="$(mktemp)"
BODY_FILE="$(mktemp)"

fail() {
	echo "::error::$1"
	echo "--- server log"
	cat "$LOG_FILE"
	exit 1
}

# expect STATUS LABEL CURL_ARGS... runs curl and fails unless it returns STATUS.
expect() {
	local want="$1" label="$2"
	shift 2
	local got
	got="$(curl -s -o "$BODY_FILE" -w '%{http_code}' "$@")"
	[ "$got" = "$want" ] || fail "$label: expected HTTP $want, got $got: $(cat "$BODY_FILE")"
	echo "ok: $label -> HTTP $got"
}

# login EMAIL PASSWORD STATUS logs in and, on 200, prints the access token.
login() {
	expect "$3" "login $1" -X POST "$BASE_URL/login" -H 'Content-Type: application/json' \
		-d "{\"email\":\"$1\",\"password\":\"$2\"}" >&2
	if [ "$3" = 200 ]; then jq -r '.access_token.token' "$BODY_FILE"; fi
}

go build -o bin/server ./cmd/server/main.go
RUN_MIGRATE=true ./bin/server >"$LOG_FILE" 2>&1 &
SERVER_PID=$!
trap 'kill "$SERVER_PID" 2>/dev/null || true' EXIT

for _ in $(seq 1 30); do
	curl -fs "http://127.0.0.1:${PORT}/healthz" >/dev/null && break
	kill -0 "$SERVER_PID" 2>/dev/null || fail "server exited during startup"
	sleep 1
done
curl -fs "http://127.0.0.1:${PORT}/healthz" >/dev/null || fail "server did not become healthy"

go run ./cmd/seeder

# Public endpoints allow 10 requests per minute per IP; this script makes 10.

# The first seeded user is the admin, the second has no role.
admin_token="$(login john@example.com password123 200)"
admin_refresh="$(jq -r '.refresh_token.token' "$BODY_FILE")"

# The refresh token is rotated: the new one works, the old one no longer does.
refresh_body() { echo "{\"refresh_token\":\"$1\",\"access_token\":\"$admin_token\"}"; }
expect 200 "refresh token" -X POST "$BASE_URL/refresh-token" -H 'Content-Type: application/json' -d "$(refresh_body "$admin_refresh")"
expect 401 "reuse rotated refresh token" -X POST "$BASE_URL/refresh-token" -H 'Content-Type: application/json' -d "$(refresh_body "$admin_refresh")"
expect 200 "admin GET /profile" "$BASE_URL/profile" -H "Authorization: Bearer $admin_token"
expect 200 "admin GET /roles" "$BASE_URL/roles" -H "Authorization: Bearer $admin_token"

user_token="$(login jane@example.com password123 200)"
expect 403 "user without role GET /roles" "$BASE_URL/roles" -H "Authorization: Bearer $user_token"

# Five wrong passwords lock the account: the right password is then refused
# too, with the same answer as a wrong one.
for _ in 1 2 3 4 5; do login jane@example.com wrong-password 400; done
wrong_answer="$(cat "$BODY_FILE")"
login jane@example.com password123 400
[ "$(cat "$BODY_FILE")" = "$wrong_answer" ] || fail "a locked account must answer like a wrong password, got: $(cat "$BODY_FILE")"

echo "Smoke test passed"
