# Backend — Banu Musa ERP

Go REST API server backed by SQLite.

## Prerequisites

- [Go 1.24+](https://go.dev/dl/)

Install on macOS:
```bash
brew install go
```

Verify:
```bash
go version
```

## Run

```bash
BANU_MUSA_FCM_ENABLED=false go run .
```

Server starts on **http://localhost:8080**

> `BANU_MUSA_FCM_ENABLED=false` disables Firebase push notifications — required locally unless you have the Firebase service account file.

## Environment Variables

All variables are optional with sensible defaults for local development.

| Variable | Default | Description |
|---|---|---|
| `BANU_MUSA_PORT` | `8080` | HTTP server port |
| `BANU_MUSA_DB_PATH` | `./data/banumusa.db` | SQLite database file path |
| `BANU_MUSA_AUTH_ENABLED` | `true` | Enable JWT authentication |
| `BANU_MUSA_DEV_OTP_BYPASS` | `true` | Allow fixed OTP code in dev |
| `BANU_MUSA_DEV_BYPASS_OTP` | `112233` | The OTP code to use when bypass is on |
| `BANU_MUSA_JWT_SECRET` | `dev-secret-change-in-production` | JWT signing secret |
| `BANU_MUSA_REFRESH_TOKEN_DAYS` | `90` | JWT refresh token lifetime |
| `BANU_MUSA_FCM_ENABLED` | `true` | Enable Firebase push notifications |
| `BANU_MUSA_FCM_SERVICE_ACCOUNT_PATH` | `./firebase/staging-banumusa-firebase-adminsdk-*.json` | Path to Firebase service account JSON |
| `BANU_MUSA_SCHEDULER_ENABLED` | `true` | Enable background job scheduler |
| `BANU_MUSA_SCHEDULER_INTERVAL_HOURS` | `1` | How often scheduler runs (hours) |
| `BANU_MUSA_EXPIRED_LEAVE_GRACE_DAYS` | `1` | Days before auto-rejecting expired leave requests |
| `BANU_MUSA_HOLIDAY_SYNC_ENDPOINT` | `https://calendarific.com/api/v2/holidays` | Calendarific holidays API endpoint |
| `BANU_MUSA_HOLIDAY_SYNC_API_KEY` | `` | Calendarific API key (required for holiday sync) |
| `BANU_MUSA_HOLIDAY_SYNC_COUNTRY` | `EG` | Country code for holiday sync |
| `BANU_MUSA_HOLIDAY_SYNC_TIMEZONE` | `Africa/Cairo` | Timezone used by holiday sync |

## Database

SQLite file lives at `data/banumusa.db`. Migrations run automatically on startup from `adapters/db/migrations/`.

Inspect the database with:
```bash
# CLI
sqlite3 data/banumusa.db

# GUI — open the file in TablePlus, DB Browser for SQLite, or any SQLite client
```

## View Logs

Logs are structured (key=value) and printed to stdout. To follow them while the server runs:
```bash
BANU_MUSA_FCM_ENABLED=false go run . 2>&1 | tee backend.log
```

Or if running via the background command:
```bash
tail -f /tmp/backend.log
```

## API Reference

Base URL: `http://localhost:8080`

### Authentication (Public — no token required)

| Method | Path | Description |
|---|---|---|
| `POST` | `/api/v1/auth/otp/request` | Request OTP sent to phone |
| `POST` | `/api/v1/auth/otp/verify` | Verify OTP and receive tokens |
| `POST` | `/api/v1/auth/login` | Login with phone + password |
| `POST` | `/api/v1/auth/refresh` | Refresh access token |

**Login flow (dev):**
```bash
# Step 1: request OTP
curl -X POST http://localhost:8080/api/v1/auth/otp/request \
  -H "Content-Type: application/json" \
  -d '{"phone": "01012345678"}'

# Step 2: verify with bypass OTP
curl -X POST http://localhost:8080/api/v1/auth/otp/verify \
  -H "Content-Type: application/json" \
  -d '{"phone": "01012345678", "otp": "112233"}'
```

### Authentication (Protected — requires `Authorization: Bearer <token>`)

| Method | Path | Description |
|---|---|---|
| `POST` | `/api/v1/auth/logout` | Logout and invalidate refresh token |
| `GET` | `/api/v1/auth/me` | Get current user info |

### Dashboard

| Method | Path | Permission |
|---|---|---|
| `GET` | `/api/v1/dashboard/stats` | — |

### Employees

| Method | Path | Permission |
|---|---|---|
| `GET` | `/api/v1/employees` | `employees:read` |
| `GET` | `/api/v1/employees/{uid}` | `employees:read` |
| `POST` | `/api/v1/employees/import` | `employees:import` |
| `GET` | `/api/v1/employees/export` | `employees:export` |
| `GET` | `/api/v1/employees/export/pdf` | `employees:export` |
| `GET` | `/api/v1/employees/import/template` | `employees:import` |
| `PUT` | `/api/v1/employees/{uid}/department` | `employees:write` |
| `DELETE` | `/api/v1/employees/{uid}/department` | `employees:write` |

### Users & Roles

| Method | Path | Permission |
|---|---|---|
| `GET` | `/api/v1/users` | `users:read` |
| `POST` | `/api/v1/users` | `users:write` |
| `PUT` | `/api/v1/users/{uid}` | `users:write` |
| `POST` | `/api/v1/users/{uid}/roles` | `users:write` |
| `DELETE` | `/api/v1/users/{uid}/roles/{roleUid}` | `users:write` |
| `GET` | `/api/v1/roles` | `roles:read` |
| `POST` | `/api/v1/roles` | `roles:write` |
| `POST` | `/api/v1/roles/{uid}/permissions` | `roles:write` |
| `GET` | `/api/v1/permissions` | `roles:read` |

### Leave Management

| Method | Path | Permission |
|---|---|---|
| `POST` | `/api/v1/leave` | `leave:record` |
| `GET` | `/api/v1/leave-records` | `leave:read` |
| `GET` | `/api/v1/employees/{uid}/balance` | `leave:read` |
| `GET` | `/api/v1/employees/{uid}/leave-records` | `leave:read` |

### Leave Requests (Employee)

| Method | Path | Description |
|---|---|---|
| `POST` | `/api/v1/leave-requests` | Submit a leave request |
| `GET` | `/api/v1/leave-requests` | List my leave requests |
| `GET` | `/api/v1/leave-requests/{uid}` | Get a specific request |
| `POST` | `/api/v1/leave-requests/{uid}/cancel` | Cancel a request |

### Approvals (Approver)

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/v1/approvals/pending` | List requests pending my approval |
| `POST` | `/api/v1/approvals/{uid}/approve` | Approve a request |
| `POST` | `/api/v1/approvals/{uid}/reject` | Reject a request |
| `GET` | `/api/v1/approvals/{uid}/history` | View approval history |

### Admin

| Method | Path | Permission |
|---|---|---|
| `GET` | `/api/v1/admin/departments` | `departments:read` |
| `GET` | `/api/v1/admin/departments/{uid}` | `departments:read` |
| `POST` | `/api/v1/admin/departments` | `departments:write` |
| `PATCH` | `/api/v1/admin/departments/{uid}` | `departments:write` |
| `POST` | `/api/v1/admin/departments/{uid}/manager` | `departments:write` |
| `DELETE` | `/api/v1/admin/departments/{uid}/manager` | `departments:write` |
| `GET` | `/api/v1/admin/approval-flows` | `approval-flows:read` |
| `POST` | `/api/v1/admin/approval-flows` | `approval-flows:write` |
| `PATCH` | `/api/v1/admin/approval-flows/{uid}` | `approval-flows:write` |
| `GET` | `/api/v1/admin/approval-flows/{uid}/steps` | `approval-flows:read` |
| `POST` | `/api/v1/admin/approval-flows/{uid}/steps` | `approval-flows:write` |
| `PATCH` | `/api/v1/admin/approval-flows/{uid}/steps/{stepUid}` | `approval-flows:write` |
| `DELETE` | `/api/v1/admin/approval-flows/{uid}/steps/{stepUid}` | `approval-flows:write` |
| `GET` | `/api/v1/admin/leave-types` | `leave-types:read` |
| `PATCH` | `/api/v1/admin/leave-types/{uid}/active` | `leave-types:write` |
