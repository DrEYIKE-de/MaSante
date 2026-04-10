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
  <img src="https://img.shields.io/badge/database-SQLite-003B57?logo=sqlite&logoColor=white" alt="SQLite">
  <img src="https://img.shields.io/badge/platform-Windows%20%7C%20macOS%20%7C%20Linux-lightgrey" alt="Platform">
  <img src="https://img.shields.io/badge/tests-78%20passing-brightgreen" alt="Tests">
</p>

---

MaSante helps health centers manage patient appointments, send SMS reminders, and track community health worker visits — all from a single binary with zero external dependencies.

Designed for low-resource settings: works offline, runs on any machine, never mentions patient conditions in the interface or messages.

## Features

- **Setup wizard** — 5-step guided first-launch configuration (center info, admin account, schedule, SMS provider)
- **Patient management** — register, search, filter by status, risk scoring, program exit (death, transfer, dropout, cured)
- **Appointment scheduling** — slot-based booking, calendar view, post-visit workflow (complete, missed, reschedule, cancel)
- **SMS reminders** — automated J-7, J-2, day-of reminders via 5 providers (Africa's Talking, MTN, Orange, Twilio, Infobip)
- **Community health workers (ASC)** — field visit list, report form, zone map
- **User management** — role-based access control (Admin, Medecin, Infirmier, ASC)
- **Exports** — patient lists and monthly reports in Excel and PDF
- **Help center** — embedded guides and FAQ, works offline
- **Confidentiality** — no disease names anywhere in the UI, SMS templates, or exports

## Quick start

### Download and run

Download the latest binary for your platform from [masante.africa](https://masante.africa), then:

```bash
./masante
```

The browser opens automatically at `http://localhost:8080`. The setup wizard guides you through the initial configuration.

### Build from source

Requires Go 1.22+ and a C compiler (for SQLite).

```bash
git clone https://github.com/DrEYIKE-de/MaSante.git
cd MaSante
go build -o masante ./cmd/masante/
./masante
```

### CLI options

```
./masante [options]

Options:
  --port PORT        HTTP port (default: 8080)
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
├── web/                 Embedded frontend (served from the binary)
└── cmd/masante/         Composition root
```

### Domain ports

| Port | Description |
|---|---|
| `PatientRepository` | Patient CRUD, search, code generation, status counts |
| `AppointmentRepository` | Appointment CRUD, calendar queries, slot availability |
| `UserRepository` | User CRUD, username lookup |
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
| Patients | `GET/POST /api/v1/patients`, `GET/PUT /api/v1/patients/{id}`, `PUT .../exit` | Auth |
| Appointments | `POST /api/v1/appointments`, `PUT .../{complete,missed,reschedule}`, `DELETE ...` | Auth |
| Calendar | `GET /api/v1/calendar/week`, `GET /api/v1/appointments/slots` | Auth |
| Reminders | `GET /api/v1/reminders`, `POST .../test`, `POST .../send-all` | Auth |
| Users | `GET/POST/PUT/DELETE /api/v1/users/{id}` | Admin only |
| Profile | `GET/PUT /api/v1/profile`, `PUT .../password` | Auth |
| Exports | `GET /api/v1/export/{patients,monthly}/{excel,pdf}` | Auth |

## Tech stack

| Component | Choice | Why |
|---|---|---|
| Language | Go | Single binary, no runtime dependencies |
| Database | SQLite (WAL mode) | Zero config, embedded, works offline |
| HTTP | `net/http` (stdlib) | No framework dependency, Go 1.22 routing |
| Password hashing | bcrypt (cost 12) | Industry standard |
| Auth | Server-side sessions in SQLite | Instant revocation, no JWT complexity |
| SMS | Africa's Talking, MTN, Orange, Twilio, Infobip | Covers all major African providers |
| Excel | excelize | .xlsx generation |
| PDF | go-pdf/fpdf | PDF report generation |
| Frontend | Embedded HTML/CSS/JS via `go:embed` | No build step, works offline |

## Database

SQLite with WAL mode. Data stored in `./masante-data/masante.db`.

Designed for small to mid-size clinics: up to 2,000 patients and 5 concurrent users comfortably. For larger deployments, a migration path to PostgreSQL is planned.

### Tables

`center` · `users` · `sessions` · `patients` · `appointments` · `reminders` · `message_templates` · `asc_visits` · `sms_config` · `settings` · `audit_log`

## Security

- **No disease names** anywhere — SMS messages say "health appointment", not the condition
- **Password policy** — minimum 8 characters with at least one digit
- **Sessions** — server-side with 24h expiry, instant revocation on logout or password change
- **RBAC** — four roles with granular permissions (Admin > Medecin > Infirmier > ASC)
- **Audit trail** — every sensitive action logged (patient create/exit, user management, login)
- **Soft delete** — users are disabled, never hard-deleted

## SMS providers

MaSante does not provide SMS service. You bring your own provider account. Supported:

| Provider | Coverage | Setup |
|---|---|---|
| [Africa's Talking](https://africastalking.com) | 25+ African countries | Recommended for most clinics |
| [MTN SMS API](https://developer.mtn.com) | Cameroon, Nigeria, Ghana, DRC, Uganda... | Best if patients are on MTN |
| [Orange SMS API](https://developer.orange.com) | Cameroon, Senegal, Ivory Coast, Mali... | Direct carrier integration |
| [Twilio](https://twilio.com) | Worldwide | Most reliable, slightly more expensive |
| [Infobip](https://infobip.com) | Worldwide + WhatsApp | Good if you need WhatsApp support |

Configure during the setup wizard or later in Settings.

## Background processes

| Process | Interval | Purpose |
|---|---|---|
| Reminder scheduler | 5 minutes | Generate J-7/J-2/J-0 reminders, send pending, retry failed (max 3) |
| Session cleanup | 1 hour | Remove expired auth sessions |

## Testing

```bash
go test ./... -v
```

78 tests covering domain logic, application services, SQLite repositories, HTTP handlers, SMS factory, exports, and password validation. No external mocking framework — manual mocks only.

## Roadmap

- [ ] Internationalization (French, English, Duala, Ewondo, Bamileke)
- [ ] PostgreSQL adapter for large deployments
- [ ] WhatsApp Business API integration
- [ ] DHIS2 / OpenMRS data export
- [ ] Mobile-optimized progressive web app
- [ ] Landing page at masante.africa

## License

Open source. See [LICENSE](LICENSE) for details.

## Contributing

Contributions welcome. Please open an issue before submitting a pull request.

Built for health workers in Africa, by people who understand the challenges they face every day.
