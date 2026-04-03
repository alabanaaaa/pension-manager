# Pension Manager

**Post Retirement Medical Fund & Pension Management System**

A production-grade, event-sourced pension and medical fund management platform built for Kenyan retirement schemes. Designed for regulatory compliance with immutable audit trails, maker-checker workflows, and real-time fraud detection.

---

## What This System Does

This system manages the complete lifecycle of a post-retirement medical fund and pension scheme:

- **Member Management** — Register members, track beneficiaries, manage contributions (DB & DC schemes)
- **Contribution Processing** — Monthly remittances via M-Pesa, bank transfer, or bulk CSV upload with automatic reconciliation
- **Claims & Withdrawals** — Full workflow from submission through maker-checker approval to payment, with document management
- **Online Voting** — USSD + web-based trustee elections with real-time dashboards, one-vote-per-member enforcement
- **Benefit Projections** — Actuarial calculations for both Defined Benefit and Defined Contribution schemes
- **Tax Computation** — KRA-compliant tax calculations with exemption handling
- **Member Self-Service Portal** — View balances, update beneficiaries, request changes, download statements
- **Fraud Detection** — Ghost Mode anomaly detection for duplicate claims, ghost members, contribution manipulation

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                         MEMBER PORTAL                               │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐           │
│  │  Web UI  │  │  Mobile  │  │   USSD   │  │  Admin   │           │
│  │  (HTMX)  │  │  (Flutter│  │ (Africa's│  │  Panel   │           │
│  │          │  │  future) │  │ Talking) │  │          │           │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘  └────┬─────┘           │
└───────┼─────────────┼─────────────┼─────────────┼──────────────────┘
        │             │             │             │
        ▼             ▼             ▼             ▼
┌─────────────────────────────────────────────────────────────────────┐
│                      API GATEWAY (Go / chi)                         │
│                                                                     │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ │
│  │  Members │ │Contrib-  │ │  Claims  │ │  Voting  │ │ Reports  │ │
│  │  CRUD    │ │  utions  │ │ Workflow │ │  Engine  │ │ & Export │ │
│  └────┬─────┘ └────┬─────┘ └────┬─────┘ └────┬─────┘ └────┬─────┘ │
│       │             │             │             │             │      │
│  ┌────┴─────────────┴─────────────┴─────────────┴─────────────┴──┐ │
│  │                  EVENT SOURCING ENGINE                         │ │
│  │                                                                │ │
│  │  ┌────────────┐  ┌────────────┐  ┌──────────────────────────┐ │ │
│  │  │ Hash-Chain │  │  Snapshot  │  │ Ghost Mode (Fraud Detect)│ │ │
│  │  │ Audit Log  │  │  Recovery  │  │ Anomaly Detection        │ │ │
│  │  └────────────┘  └────────────┘  └──────────────────────────┘ │ │
│  └────────────────────────┬───────────────────────────────────────┘ │
└───────────────────────────┼─────────────────────────────────────────┘
                            │
        ┌───────────────────┼───────────────────┐
        ▼                   ▼                   ▼
┌───────────────┐  ┌───────────────┐  ┌───────────────┐
│  PostgreSQL   │  │  M-Pesa API   │  │  SMS Gateway  │
│  (Primary DB) │  │  (Daraja)     │  │  (Africa's    │
│               │  │  STK Push     │  │   Talking)    │
│  • Members    │  │  Callbacks    │  │  OTP, Alerts  │
│  • Events     │  │  Status Check │  │  Bulk SMS     │
│  • Claims     │  └───────────────┘  └───────────────┘
│  • Votes      │
│  • Audit Log  │  ┌───────────────┐  ┌───────────────┐
└───────────────┘  │  S3/MinIO     │  │  Redis        │
                   │  Documents    │  │  (future)     │
                   │  Death Certs  │  │  Voting State │
                   │  IDs, Forms   │  │  Sessions     │
                   └───────────────┘  └───────────────┘
```

## Data Flow: Contribution Example

```
Sponsor uploads CSV ──→ API validates ──→ Event created (hash-chained)
                           │
                    ┌──────┴──────┐
                    ▼             ▼
            PostgreSQL       Engine KV
         (authoritative)   (fast replay)
                    │             │
                    ▼             ▼
           Member balance    Ghost Mode
            updated         checks for
                           anomalies
                    │
                    ▼
           M-Pesa callback
          confirms payment
                    │
                    ▼
        Quarterly statement
        generated & emailed
```

## Tech Stack

| Layer | Technology |
|---|---|
| Backend | Go 1.25+, chi router |
| Database | PostgreSQL 16 (primary), SQLite (offline) |
| Event Store | Custom append-only KV with CRC32 + SHA256 hash chain |
| Auth | JWT (HS256), bcrypt, OTP via SMS |
| Payments | Safaricom Daraja API (M-Pesa STK Push) |
| SMS | Africa's Talking (OTP, alerts, bulk) |
| USSD | Africa's Talking USSD gateway |
| Frontend | HTMX + TailwindCSS (server-rendered, no JS framework) |
| Documents | S3/MinIO (death certs, IDs, forms) |
| Deployment | Docker, docker-compose |

## Project Structure

```
pension-manager/
├── cmd/
│   ├── server/              # HTTP API server
│   └── admin/               # CLI admin tool (cobra)
├── core/
│   ├── domain/              # Domain models (member, contribution, claim, vote)
│   ├── errors.go            # Typed error system
│   └── db/                  # Custom KV store with snapshots
├── engine/
│   ├── engine.go            # Event sourcing core
│   ├── contributions.go     # Contribution processing
│   ├── claims.go            # Claims workflow
│   ├── voting.go            # Election engine
│   └── ghost_mode.go        # Fraud/anomaly detection
├── internal/
│   ├── api/                 # HTTP handlers, middleware, server setup
│   ├── auth/                # JWT, bcrypt, OTP
│   ├── mpesa/               # Daraja API client
│   ├── sms/                 # Africa's Talking client
│   ├── documents/           # S3 document management
│   ├── tax/                 # KRA tax computation
│   └── importcsv/           # Bulk CSV import/export
├── projection/              # Event → read model materialization
├── ledger/                  # Hash chain computation
├── migrations/              # PostgreSQL schema migrations
├── web/
│   ├── templates/           # HTMX pages (login, dashboard, members, claims, voting)
│   └── static/              # CSS, JS, service worker
└── monitoring/              # Prometheus metrics
```

## Compliance & Audit

Built for pension regulatory requirements:

- **Immutable audit trail** — Every change is a hash-chained event. Tampering breaks the chain.
- **Before/after values** — Event data captures full state transitions.
- **Maker-checker** — All member/beneficiary changes require approval workflow.
- **Timestamp + actor** — Every event records who did what and when.
- **Data retention** — Event replay allows full historical reconstruction.
- **Role-based access** — Fine-grained permissions (admin, officer, member, auditor).

## Getting Started

### Prerequisites
- Go 1.25+
- PostgreSQL 16
- Docker & docker-compose (optional)

### Quick Start

```bash
# Clone
git clone <repo-url>
cd pension-manager

# Copy environment
cp .env.example .env

# Start PostgreSQL
docker compose up -d postgres

# Run migrations
make migrate-up

# Run tests
make test

# Start server
make run
```

### Environment Variables

```env
DATABASE_URL=postgres://postgres:postgres@localhost:5432/pension?sslmode=disable
JWT_SECRET=your-secret-key-here
HTTP_PORT=8080
ENV=development

# M-Pesa (Daraja API)
MPESA_CONSUMER_KEY=...
MPESA_CONSUMER_SECRET=...
MPESA_SHORT_CODE=...
MPESA_PASSKEY=...
MPESA_ENVIRONMENT=sandbox

# SMS (Africa's Talking)
AT_API_KEY=...
AT_USERNAME=...
AT_SHORTCODE=...

# Document Storage (S3/MinIO)
S3_ENDPOINT=...
S3_BUCKET=...
S3_ACCESS_KEY=...
S3_SECRET_KEY=...
```

## API Reference

### Authentication
- `POST /api/auth/login` — Login with email/password
- `POST /api/auth/refresh` — Refresh access token
- `POST /api/auth/otp/request` — Request OTP via SMS
- `POST /api/auth/otp/verify` — Verify OTP

### Members
- `GET /api/members` — List members (paginated, filtered)
- `POST /api/members` — Register new member (maker-checker)
- `GET /api/members/{id}` — Get member details
- `PUT /api/members/{id}` — Update member (maker-checker)
- `GET /api/members/{id}/beneficiaries` — List beneficiaries
- `POST /api/members/{id}/beneficiaries` — Add beneficiary (maker-checker)

### Contributions
- `POST /api/contributions` — Record contribution
- `POST /api/contributions/mpesa` — Initiate M-Pesa STK Push
- `POST /api/contributions/bulk` — Bulk CSV upload
- `GET /api/contributions/{member_id}` — Member contribution history
- `POST /api/contributions/reconcile` — Reconcile sponsor remittance

### Claims
- `POST /api/claims` — Submit claim
- `GET /api/claims` — List claims (filtered by status)
- `GET /api/claims/{id}` — Claim details
- `PUT /api/claims/{id}/approve` — Approve claim (maker-checker)
- `PUT /api/claims/{id}/reject` — Reject claim with reason
- `PUT /api/claims/{id}/pay` — Mark claim as paid

### Voting
- `POST /api/voting/elections` — Create election
- `POST /api/voting/elections/{id}/vote` — Cast vote (USSD or web)
- `GET /api/voting/elections/{id}/results` — Real-time results
- `GET /api/voting/elections/{id}/voters` — Who has voted

### Reports
- `GET /api/reports/quarterly` — Quarterly contribution statement
- `GET /api/reports/member/{id}/statement` — Member benefit statement
- `GET /api/reports/contributions` — Contribution trends
- `GET /api/reports/claims` — Claims summary
- `GET /api/reports/export` — Export to CSV/PDF

## Development Roadmap

### Phase 1: Core Foundation (Weeks 1-3)
- [x] Project scaffold & architecture
- [x] Database schema (members, beneficiaries, contributions, claims, votes)
- [ ] Member CRUD with maker-checker
- [ ] Contribution recording + M-Pesa integration
- [ ] Basic reconciliation
- [ ] Admin dashboard

### Phase 2: Claims & Statements (Weeks 4-6)
- [ ] Claims workflow with state machine
- [ ] Document uploads (S3)
- [ ] Quarterly contribution statements (PDF)
- [ ] Member self-service portal

### Phase 3: Advanced Features (Weeks 7-9)
- [ ] Benefit projections (DB/DC actuarial)
- [ ] Tax computation (KRA rules)
- [ ] Bulk contribution processing with validation
- [ ] SMS alerts + OTP authentication

### Phase 4: Voting & Mobile (Weeks 10-12)
- [ ] Online voting system (web + USSD)
- [ ] Real-time voting dashboards
- [ ] Mobile-responsive portal
- [ ] Flutter mobile app (planned)

## License

Private — All rights reserved.
