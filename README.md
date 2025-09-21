# Library Lending API

Golang REST API for a library book lending system.

## Features
- User registration & login (JWT bearer)
- Book search with filters (title, author, genre, availability), sorting & pagination
- Create idempotent loan requests (Idempotency-Key header) & return books
- View own loan history
- RFC 9457 Problem Details for errors (application/problem+json)
- Simple per-user/IP rate limiting (60 req/min window token bucket)
- OpenAPI 3.1 spec (`openapi/openapi.yaml`)
- Docker & docker-compose (API + Postgres) with auto migrations & seed data

## Tech Stack
- Go 1.22
- chi router
- pgx PostgreSQL driver
- JWT (HS256)

## Run Locally
Prerequisites: Go 1.22, Postgres running (or use docker-compose).

1. Start Postgres (if local): ensure a database `library` exists.
2. Set env vars (optional):
```
set DATABASE_URL=postgres://postgres:postgres@localhost:5432/library?sslmode=disable
set JWT_SECRET=devsecret
```
3. Run server:
```
go run ./cmd/server
```
4. Health check: `GET http://localhost:8080/v1/healthz`

## Docker
### Prerequisites
Install Docker Desktop (Windows):
1. Enable WSL2 (PowerShell as Administrator):
```
wsl --install
```
2. Reboot if prompted.
3. Download & install Docker Desktop: https://www.docker.com/products/docker-desktop/
4. In Docker Desktop settings ensure: "Use the WSL 2 based engine" is enabled.

Verify after opening a new PowerShell:
```
docker --version
docker compose version
```

### Start Stack (API + Postgres)
Option A (manual):
```
docker compose up --build
```

Option B (helper script):
```
pwsh .\scripts\start-docker.ps1
```

Flags:
```
pwsh .\scripts\start-docker.ps1 -Rebuild    # no-cache rebuild
pwsh .\scripts\start-docker.ps1 -ResetDb    # drop volumes & recreate
pwsh .\scripts\start-docker.ps1 -Logs       # tail API logs after start
```

API base: http://localhost:8080  (Health: /v1/healthz)

### Common Docker Tasks
| Action | Command |
|--------|---------|
| Stop stack | `docker compose down` |
| Stop & remove volumes (reset DB) | `docker compose down -v` |
| View logs | `docker compose logs -f api` |
| Exec into API container | `docker compose exec api sh` |
| Exec into Postgres | `docker compose exec db psql -U postgres -d library` |

### Troubleshooting
| Problem | Cause | Fix |
|---------|-------|-----|
| Port 8080 in use | Another service bound | Stop other service or change `PORT` env & expose mapping |
| Container restarts repeatedly | DB not healthy yet | Wait for healthcheck; logs: `docker compose logs db` |
| Migrations not applying | Permissions or crash | Check API logs; ensure migration files present in image |
| Slow startup on first run | Image build + downloading layers | Subsequent runs are faster |
| "no space left on device" | WSL2 disk full | Prune: `docker system prune -af` |

### Rebuild After Code Changes
Code changes require rebuild because of multi-stage build:
```
docker compose build api
docker compose up -d
```

### Environment Overrides
Edit `docker-compose.yml` or override at run-time:
```
setx JWT_SECRET "stronger-secret"   # Persist (needs new shell)
```

## Authentication
Register: `POST /v1/auth/register` JSON `{ "email": "a@b.com", "password": "secret123" }`
Login: `POST /v1/auth/login` -> `{ "token": "<JWT>" }`
Use: `Authorization: Bearer <JWT>`

## Idempotent Loan Creation
`POST /v1/loans` with header `Idempotency-Key: <uuid>` and body `{ "book_id": 1 }`.
Repeating the same request with identical key returns cached response.

## Endpoints Summary
| Method | Path | Description |
|--------|------|-------------|
| GET | /v1/healthz | Health check |
| POST | /v1/auth/register | Register user |
| POST | /v1/auth/login | Login user |
| GET | /v1/books | Search books |
| POST | /v1/loans | Create loan (auth) |
| PATCH | /v1/loans/{id}/return | Return loan (auth) |
| GET | /v1/me/loans | List own loans (auth) |

## OpenAPI
See `openapi/openapi.yaml`.

## Error Format
```
Content-Type: application/problem+json
{
	"type": "about:blank",
	"title": "validation error",
	"status": 400,
	"detail": "invalid input",
	"invalidParams": [ {"name":"password","reason":"length must be >="} ],
	"timestamp": "2025-09-21T12:00:00Z"
}
```

## Rate Limiting
Per user (authenticated) or IP (anonymous): 60 requests/minute window.

## End-to-End curl Test Suite
All commands assume API base `http://localhost:8080/v1` and you have `curl` (Git Bash / WSL) plus optional `jq` for JSON formatting.

Set a reusable environment variable (Bash / Git Bash / WSL):
```bash
BASE=http://localhost:8080/v1
EMAIL="tester$(date +%s)@example.com"
PASS='Secret123!'
```

### 1. Health Check
```bash
curl -i "$BASE/healthz"
```

### 2. Register User
```bash
curl -i -X POST "$BASE/auth/register" \
	-H 'Content-Type: application/json' \
	-d '{"email":"'$EMAIL'","password":"'$PASS'"}'
```
Expect: 200 + token (in JSON `{"token":"..."}`)

### 3. Login (gets JWT token)
```bash
TOKEN=$(curl -s -X POST "$BASE/auth/login" -H 'Content-Type: application/json' \
	-d '{"email":"'$EMAIL'","password":"'$PASS'"}' | jq -r .token)
echo "TOKEN=$TOKEN"
```

### 4. Book Search (filters, pagination)
```bash
curl -s "$BASE/books?limit=5&available=true" | jq
curl -s "$BASE/books?author=Rowling&sort=title" | jq
```

### 5. Create Loan (Idempotent)
```bash
IDEMP_KEY=$(uuidgen 2>/dev/null || cat /proc/sys/kernel/random/uuid)
curl -i -X POST "$BASE/loans" \
	-H "Authorization: Bearer $TOKEN" \
	-H "Idempotency-Key: $IDEMP_KEY" \
	-H 'Content-Type: application/json' \
	-d '{"book_id":1}'
```
Repeat same request (should return same loan / 201 response without duplicating):
```bash
curl -i -X POST "$BASE/loans" \
	-H "Authorization: Bearer $TOKEN" \
	-H "Idempotency-Key: $IDEMP_KEY" \
	-H 'Content-Type: application/json' \
	-d '{"book_id":1}'
```

### 6. List My Loans
```bash
curl -s -H "Authorization: Bearer $TOKEN" "$BASE/me/loans" | jq
```

### 7. Return a Loan
Assuming loan id = 1 (adjust if needed):
```bash
curl -i -X PATCH "$BASE/loans/1/return" -H "Authorization: Bearer $TOKEN"
```
Repeat (idempotent):
```bash
curl -i -X PATCH "$BASE/loans/1/return" -H "Authorization: Bearer $TOKEN"
```

### 8. Error / Edge Case Tests

Duplicate Email:
```bash
curl -i -X POST "$BASE/auth/register" -H 'Content-Type: application/json' \
	-d '{"email":"'$EMAIL'","password":"'$PASS'"}'
```

Missing Idempotency Key:
```bash
curl -i -X POST "$BASE/loans" -H "Authorization: Bearer $TOKEN" \
	-H 'Content-Type: application/json' -d '{"book_id":1}'
```

Invalid JSON:
```bash
curl -i -X POST "$BASE/auth/register" -H 'Content-Type: application/json' -d '{"email":123}'
```

Unauthorized Access (list loans without token):
```bash
curl -i "$BASE/me/loans"
```

Rate Limit (expect some 429 after ~60 requests):
```bash
for i in $(seq 1 70); do curl -s -o /dev/null -w "${i}:%{http_code} " "$BASE/healthz"; done; echo
```

Book Unavailable Scenario (borrow same book until exhausted):
```bash
for i in $(seq 1 10); do \
	KEY=$(uuidgen 2>/dev/null || cat /proc/sys/kernel/random/uuid); \
	curl -s -X POST "$BASE/loans" -H "Authorization: Bearer $TOKEN" -H "Idempotency-Key: $KEY" \
			 -H 'Content-Type: application/json' -d '{"book_id":1}' | jq '.error?, .id?'; \
done
```

### 9. Troubleshooting with curl
Add `-v` for verbose, or `-D -` to inspect headers:
```bash
curl -v "$BASE/healthz"
curl -D - -s -o /dev/null "$BASE/healthz"
```

### 10. Cleanup (Docker)
```bash
docker compose down -v
```

---
These commands cover success, idempotency, error responses, rate limiting, and resource state changes (loan lifecycle).

## Future Improvements
- Refresh tokens / logout
- Admin endpoints to add/update books
- More granular error types & logging
- Testing suite expansion

## License
MIT (add LICENSE file if distributing)