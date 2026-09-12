**Jump to:** [Requirements checklist](#requirements-checklist) · [Project structure](#project-structure) · [Run with Docker](#run-with-docker) · [Run locally](#run-locally) · [API endpoints](#api-endpoints) · [Environment variables](#environment-variables) · [Tests](#tests)

# Take-home Assignment: Auth with JWT (TypeScript)

Build a small application in **TypeScript/Go/C#** that supports user **registration** and **login** using **JWT**.
You can choose any stack or structure you want.
As long as the core flow works end-to-end, it’s accepted.
Please note User will be using the app in place with very bad connections, like jungle or caves.

---

## Requirements

### 1. Register

- Fields: `email`, `password`, `confirmPassword`

### 2. Login

- Input: `email`, `password`
- Return: **JWT**
- Token should contain at least:
  - `email`
  - `user id` or similar identifier

### 3. Authenticated View / Endpoint

After successful login, calling the protected route / loading the protected screen should show:

```
Hello [email], welcome back
```

user should be logged out after 15 minutes of inacitvity

---

### 4. Manager wallet

User should be able to see and transfer his money to other user.
fields are: recipient, amount, and notes

## What to deliver

- Fork this repository and then send the link
- A runnable project (any structure).
- README explaining:
  - How to build and run it (prepare docker compose)
  - Required environment variables

---

## Acceptance criteria

- Registration works with validation.
- Login returns a usable JWT.
- A protected route or screen shows the welcome message using JWT auth.
- User able to transfer funds

---

## Optional bonus

- Docker
- Backend built using Go (or their frameworks)
- Frontend built using React/Vue (or their frameworks)
- Tests (unit or integration)

This keeps the scope tight: just registration, login, and a protected “Hello [email]” flow.

---

# BTech Wallet

- Backend: Go (`./backend`), Postgres, migrations run automatically on startup
- Frontend: React + Vite (`./frontend`)

## Requirements checklist

- [x] Register with email, password, confirmPassword
- [x] Login returns a JWT containing email + user id
- [x] Protected route shows "Hello [email], welcome back"
- [x] Auto-logout after 15 minutes of inactivity (15-min JWT + separate refresh token, rotated on use)
- [x] Wallet: view balance, transfer to another user (recipient, amount, notes)
- [x] Registration validation
- [x] Protected route uses JWT auth
- [x] Transfer funds works end-to-end
- [x] Docker (bonus)
- [x] Go backend (bonus)
- [x] React frontend (bonus)
- [x] Tests, unit + integration (bonus)

## Project structure

```
backend/
  auth/          registration, login, JWT + refresh-token issuance/rotation
  user/          profile endpoint
  transaction/   wallet balance, transfer, top-up, transaction history
  middleware/    auth middleware, CORS, logging
  config/        env loading, DB pool + migration runner
  migrations/    SQL migrations (embedded, run automatically on startup)
  server/        shared JSON response helpers
  docs/          generated Swagger spec

frontend/
  src/auth/      login/register pages, forms, idle-logout hook
  src/wallet/    balance card, transfer/top-up modals, transaction table
  src/components/  shared UI (pagination, etc.)
  src/routes/    TanStack Router routes
  src/utils/     axios instance with auth interceptor/refresh flow
```

## Run with Docker

Requires [Docker](https://www.docker.com/) (with Compose) installed and running.

```bash
docker compose up --build
```

This builds and starts three containers: Postgres, the Go backend, and the frontend (served via Nginx). Migrations run automatically on backend startup.

- Frontend: http://localhost:80
- Backend: http://localhost:3333 (Swagger at `/swagger/index.html`)

Env vars are already set in [docker-compose.yaml](docker-compose.yaml). No `.env` files needed. Open the frontend URL, register a user, then log in.

If port 80, 3333, or 5432 is already in use on your machine, `docker compose up` will fail to bind. Stop whatever's using it, or change the port mapping in [docker-compose.yaml](docker-compose.yaml).

New accounts start with a balance of 0, so top up your account first before trying the transfer feature. Create another account to have someone to send money to.

## Run locally

Requires [Go](https://go.dev/) 1.26+ (see [backend/go.mod](backend/go.mod)), [Node.js](https://nodejs.org/) 22+, and Postgres.

**Postgres** (or use the one from Docker: `docker compose up -d postgres`)

**Backend**

```bash
cd backend
cp .env.example .env   # edit if your DB differs from the default
go run .
```

Runs on http://localhost:3333. Migrations run automatically on startup.

**Frontend**

```bash
cd frontend
cp .env.example .env
npm install
npm run dev
```

Runs on http://localhost:5173. Open it in the browser, register a user, then log in.

## API endpoints

Full interactive spec at `/swagger/index.html` (see [Run with Docker](#run-with-docker)/[Run locally](#run-locally)).

| Method | Path                       | Auth | Description                                   |
| ------ | -------------------------- | ---- | --------------------------------------------- |
| POST   | `/auth/register`           | -    | Register (email/password/confirmPassword)     |
| POST   | `/auth/login`              | -    | Login, returns access + refresh token         |
| POST   | `/auth/refresh-token`      | -    | Rotate a refresh token for a new access token |
| GET    | `/profile`                 | JWT  | Get logged-in user info + welcome message     |
| GET    | `/wallet/balance`          | JWT  | Current wallet balance                        |
| POST   | `/wallet/transfer`         | JWT  | Transfer funds (recipient, amount, notes)     |
| POST   | `/wallet/topup`            | JWT  | Top up own balance                            |
| GET    | `/wallet/transactions`     | JWT  | Paginated transaction history                 |
| GET    | `/wallet/transaction/{id}` | JWT  | Single transaction detail                     |

## Environment variables

**backend/.env**

| Var            | Default                                        | Description                                                        |
| -------------- | ---------------------------------------------- | ------------------------------------------------------------------ |
| `PORT`         | `3333`                                         | HTTP port                                                          |
| `DATABASE_URL` | `postgres://user:password@localhost:5432/mydb` | Postgres connection string                                         |
| `JWT_SECRET`   | `your_jwt_secret_key`                          | JWT signing secret. Change this for anything beyond local/demo use |

**frontend/.env**

| Var                 | Description                                                              |
| ------------------- | ------------------------------------------------------------------------ |
| `VITE_API_BASE_URL` | Backend base URL (e.g. `http://localhost:3333`). Baked in at build time. |

## Tests

```bash
cd backend && go test ./...
cd frontend && npm test
```
