<p align="center">
  <img src="docs/logo.svg" alt="MaSante" width="80" height="80">
  <br>
  <strong style="font-size: 1.5em;">MaSante</strong>
</p>

<p align="center">
  <strong>Free, open-source appointment scheduling platform for health clinics in Africa.</strong>
  <br><br>
  <a href="https://masante.africa">Website</a> · <a href="#quick-start">Quick Start</a> · <a href="#api-endpoints">API Docs</a> · <a href="#contributing">Contributing</a>
  <br><br>
  <img src="https://img.shields.io/badge/license-open--source-green" alt="License">
  <img src="https://img.shields.io/badge/go-1.22+-00ADD8?logo=go&logoColor=white" alt="Go">
  <img src="https://img.shields.io/badge/vue-3-4FC08D?logo=vuedotjs&logoColor=white" alt="Vue">
  <img src="https://img.shields.io/badge/database-SQLite-003B57?logo=sqlite&logoColor=white" alt="SQLite">
  <img src="https://img.shields.io/badge/platform-Windows%20%7C%20macOS%20%7C%20Linux-lightgrey" alt="Platform">
  <img src="https://img.shields.io/badge/tests-95+-brightgreen" alt="Tests">
</p>

---

MaSante helps health centers manage patient appointments, send SMS reminders, and track community health worker visits — all from a single binary with zero external dependencies.

Designed for low-resource settings: works offline, runs on any machine, never mentions patient conditions in the interface or messages.

## Features

- **Setup wizard** — 5-step guided first-launch configuration with live validation and back/forward navigation
- **Patient management** — register, search, filter by status, risk scoring, program exit (death, transfer, dropout, cured)
- **Appointment scheduling** — slot-based booking with patient search, type selection, notes, and confirmation
- **Calendar** — week view (hourly grid) and month view (day grid with RDV count), click any appointment to update status
- **Appointment lifecycle** — complete, missed, reschedule (+7 days), cancel — all from the calendar modal
- **User management** — role-based access control (Admin, Medecin, Infirmier, ASC) with account lockout after 5 failed attempts
- **Profile** — edit name, email, phone; change password with validation and session revocation
- **Exports** — 6 reports (monthly, all patients, active, monitored, lost, exited) in both Excel and PDF
- **Help center** — 8 expandable guides covering every feature + FAQ
- **Confidentiality** — no disease names anywhere in the UI, SMS templates, or exports
- **GPS detection** — automatic coordinates via browser geolocation during setup

## Quick start

### Download and run

Download the latest binary for your platform from [masante.africa](https://masante.africa), then:

```bash
./masante
```

The browser opens automatically at `http://localhost:8080`. The setup wizard guides you through the initial configuration.

### Build from source

Requires Go 1.22+, Node.js 18+, and a C compiler (for SQLite).

```bash
git clone https://github.com/DrEYIKE-de/MaSante.git
cd MaSante

# Build frontend
cd web/frontend
npm install
npm run build
cd ../..

# Build binary
go build -o masante ./cmd/masante/
./masante
```

### CLI options

```
./masante [options]

Options:
  --port PORT        HTTP port (default: 8080)
  --host HOST        Bind address (default: 127.0.0.1, use 0.0.0.0 for network)
  --data-dir DIR     Data directory (default: ./masante-data)
  --version          Print version and exit
```

## Architecture

Hexagonal architecture (ports & adapters). The domain defines interfaces; adapters implement them; `main.go` wires everything.

```
masante/
├── domain/              Core business types and port interfaces (zero dependencies)
├── app/                 Application services (orchestration layer)
├── adapter/
│   ├── sqlite/          Driven adapter — persistence (SQLite)
│   ├── http/            Driving adapter — REST API (net/http)
│   ├── sms/             Driven adapter — SMS providers
│   ├── export/          Driven adapter — Excel and PDF generation
│   └── bcrypt.go        Driven adapter — password hashing
├── web/
│   ├── frontend/        Vue 3 + Vite source (SPA)
│   └── static/          Built frontend (embedded in binary via go:embed)
└── cmd/masante/         Composition root
```

### Domain ports

| Port | Description |
|---|---|
| `PatientRepository` | Patient CRUD, search, code generation, status counts |
| `AppointmentRepository` | Appointment CRUD, calendar queries, slot availability |
| `UserRepository` | User CRUD, username lookup, account lockout |
| `SessionRepository` | Auth session lifecycle |
| `ReminderRepository` | Reminder queue, delivery stats, message templates |
| `CenterRepository` | Center configuration, setup wizard state |
| `SMSConfigRepository` | SMS provider credentials and preferences |
| `ASCVisitRepository` | Community health worker visit reports |
| `AuditRepository` | Action audit trail |
| `PasswordHasher` | Password hashing and verification |
| `SMSProvider` | Send SMS messages |

### API endpoints

~40 REST endpoints grouped by domain:

| Group | Endpoints | Auth |
|---|---|---|
| Setup | `POST /api/v1/setup/{center,admin,schedule,sms,complete}` | Public (once) |
| Auth | `POST /api/v1/auth/{login,logout}`, `GET /api/v1/auth/me` | Public / Auth |
| Dashboard | `GET /api/v1/dashboard/{stats,today,overdue}` | Auth |
| Patients | `GET/POST /api/v1/patients`, `GET/PUT /api/v1/patients/{id}`, `PUT .../exit` | Auth (staff+) |
| Appointments | `POST /api/v1/appointments`, `PUT .../{complete,missed,reschedule}`, `DELETE ...` | Auth (staff+) |
| Calendar | `GET /api/v1/calendar/week`, `GET /api/v1/appointments/slots` | Auth |
| Reminders | `GET /api/v1/reminders`, `POST .../test`, `POST .../send-all` | Admin |
| Users | `GET/POST/PUT/DELETE /api/v1/users/{id}` | Admin |
| Profile | `GET/PUT /api/v1/profile`, `PUT .../password` | Auth |
| Exports | `GET /api/v1/export/{patients,monthly}/{excel,pdf}` | Admin/Medecin |

## Tech stack

| Component | Choice | Why |
|---|---|---|
| Backend | Go | Single binary, no runtime dependencies |
| Frontend | Vue 3 + Vite | Reactive forms, component-based, fast builds |
| Database | SQLite (WAL mode) | Zero config, embedded, works offline |
| HTTP | `net/http` (stdlib) | No framework dependency, Go 1.22+ routing |
| Password hashing | bcrypt (cost 12) | Industry standard |
| Auth | Server-side sessions in SQLite | Instant revocation, no JWT complexity |
| Excel | excelize | .xlsx generation |
| PDF | go-pdf/fpdf | PDF report generation |
| Tests | Go testing + Vitest | Backend + frontend coverage |

## Database

SQLite with WAL mode. Data stored in `./masante-data/masante.db` (permissions `0700`).

Designed for small to mid-size clinics: up to 2,000 patients and 5 concurrent users comfortably. For larger deployments, a migration path to PostgreSQL is planned.

### Tables

`center` · `users` · `sessions` · `patients` · `appointments` · `reminders` · `message_templates` · `asc_visits` · `sms_config` · `settings` · `audit_log` · `schema_migrations`

## Security

- **No disease names** anywhere — SMS messages say "health appointment", not the condition
- **Password policy** — minimum 8 characters with at least one digit, enforced on all endpoints
- **Account lockout** — 5 failed login attempts triggers a 15-minute lockout
- **Sessions** — server-side with 24h expiry, instant revocation on logout or password change
- **RBAC** — four roles with granular permissions (Admin > Medecin > Infirmier > ASC)
- **Security headers** — X-Content-Type-Options, X-Frame-Options, Cache-Control: no-store, Referrer-Policy
- **Request limits** — 1MB max body size to prevent DoS
- **Audit trail** — every sensitive action logged (patient create/exit, user management, login)
- **Localhost by default** — binds to 127.0.0.1 unless `--host 0.0.0.0` is specified
- **Soft delete** — users are disabled, never hard-deleted
- **No internal errors exposed** — generic error messages to clients, detailed logs server-side

## SMS providers

MaSante does not provide SMS service. You bring your own provider account. Supported:

| Provider | Coverage | Setup |
|---|---|---|
| [Africa's Talking](https://africastalking.com) | 25+ African countries | Recommended for most clinics |
| [MTN SMS API](https://developer.mtn.com) | Cameroon, Nigeria, Ghana, DRC, Uganda... | Best if patients are on MTN |
| [Orange SMS API](https://developer.orange.com) | Cameroon, Senegal, Ivory Coast, Mali... | Direct carrier integration |
| [Twilio](https://twilio.com) | Worldwide | Most reliable, slightly more expensive |

Configure during the setup wizard or later in Settings.

## Background processes

| Process | Interval | Purpose |
|---|---|---|
| Reminder scheduler | 5 minutes | Generate J-7/J-2/J-0 reminders, send pending, retry failed (max 3) |
| Session cleanup | 1 hour | Remove expired auth sessions |

## Testing

```bash
# Backend (Go)
go test ./... -v

# Frontend (Vitest)
cd web/frontend
npm test
```

78 backend tests (domain, services, repositories, handlers, SMS factory, exports, password validation) + 17 frontend tests (API client, store). No external mocking framework — manual mocks only.

## Roadmap

- [ ] Internationalization (French, English, Duala, Ewondo, Bamileke)
- [ ] PostgreSQL adapter for large deployments (5,000+ patients)
- [ ] WhatsApp Business API integration
- [ ] DHIS2 / OpenMRS data export
- [ ] Mobile-optimized progressive web app
- [ ] Landing page at masante.africa
- [ ] Community health worker (ASC) field module
- [ ] Patient risk score algorithm refinement

## License

Open source. See [LICENSE](LICENSE) for details.

## Contributing

Contributions welcome. Please open an issue before submitting a pull request.

Built for health workers in Africa, by people who understand the challenges they face every day.
