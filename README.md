# Agnos Health - Back-end Developer Technical Assessment
**Candidate:** Wanchai Pratoom  
**Position:** Back-end Developer  
**Tech Stack:** Go (Golang), Gin Framework, PostgreSQL, Docker, Nginx

---

## 🏥 Hospital Middleware System

A production-ready **Hospital Middleware System** built in **Go (Golang)** with the **Gin Framework**, **PostgreSQL**, **Nginx Reverse Proxy**, and **Docker Compose**.

This project implements all tasks, architecture specifications, database schemas, security isolations, and unit tests defined in the Agnos Back-end Developer Candidate Assignment.

---

## 📋 Assessment Deliverables & Compliance

| Requirement | Description | Status & Location |
|---|---|---|
| **Deliverable 1.a** | Project Structure Documentation | [`docs/PROJECT_STRUCTURE.md`](docs/PROJECT_STRUCTURE.md) |
| **Deliverable 1.b** | API Specification | [`docs/API_SPEC.md`](docs/API_SPEC.md) |
| **Deliverable 1.c** | ER Diagram & Schema Design | [`docs/ER_DIAGRAM.md`](docs/ER_DIAGRAM.md) |
| **Deliverable 2** | Docker Compose Server (Nginx + Go API + Postgres) | [`docker-compose.yml`](docker-compose.yml) |
| **Deliverable 3** | GitHub Repository & Clean Source Code | Repository ready with git commits |
| **Task 1** | Middleware system connecting to HIS (Hospital A API) | [`internal/service/his_client.go`](internal/service/his_client.go) |
| **Task 2** | Database schema for "Patient" model | [`internal/model/models.go`](internal/model/models.go) |
| **Task 3** | Database schema for "Staff" model with Hospital Isolation | [`internal/model/models.go`](internal/model/models.go) |
| **Task 4** | Implement `/staff/create`, `/staff/login`, `/patient/search` | [`internal/handler/`](internal/handler/) |
| **Task 5** | Unit tests covering Positive & Negative scenarios | [`internal/service/`](internal/service/) & [`internal/handler/`](internal/handler/) |

---

## 🏗️ Architecture & Technology Stack

* **Language:** Go (Golang 1.24+)
* **HTTP Framework:** [Gin Web Framework](https://github.com/gin-gonic/gin)
* **ORM & Database:** [GORM](https://gorm.io/) with PostgreSQL 16
* **Security & Auth:** JWT (`golang-jwt/jwt/v5`), Password hashing with `bcrypt` (cost 12)
* **Web Server / Reverse Proxy:** Nginx (Alpine) routing external port `80` to Go service port `8080`
* **Containerization:** Docker & Docker Compose (Multi-stage build)

---

## 🚀 Quick Start (Running with Docker Compose)

### 1. Start all services
Run the following command in the project root:
```bash
docker compose up -d --build
```

This will automatically:
1. Start **PostgreSQL 16** with persistent volume.
2. Build and launch the **Golang API** service, which connects to Postgres and automatically creates and migrates the tables (`hospitals`, `staffs`, `patients`) via GORM AutoMigrate, and seeds initial demo data.
3. Start **Nginx** on port `80` as the reverse proxy.

### 2. Verify Container Health
```bash
docker compose ps
```
You should see:
* `agnos_postgres` (healthy)
* `agnos_api` (running on 8080)
* `agnos_nginx` (running on 80)

---

## 🧪 Running Unit Tests

Run the full test suite with coverage report:
```bash
go test -v -cover ./...
```

The test suite covers:
* ✅ **Positive Scenarios:** Staff creation, login, JWT issuance, patient search with optional filters, HIS Hospital A fallback.
* ❌ **Negative Scenarios:** Duplicate staff registration, wrong credentials, cross-hospital search denial (hospital data isolation), unauthenticated / malformed JWT access.

---

## 📡 API Endpoints & cURL Testing Guide

All endpoints can be called directly via Nginx on port `80` (`http://localhost/`) or directly on the Go API (`http://localhost:8080/`).

### Step 1: Create a Hospital Staff Member
```bash
curl -X POST http://localhost/staff/create \
  -H "Content-Type: application/json" \
  -d '{
    "username": "doctor_somchai",
    "password": "Password123!",
    "hospital": "hospital-a"
  }'
```

**Response (201 Created):**
```json
{
  "status": "success",
  "message": "Staff member created successfully",
  "data": {
    "id": "270dbcf6-a832-4ae0-b6aa-67c1e54adfa4",
    "username": "doctor_somchai",
    "hospital": "hospital-a",
    "created_at": "2026-09-16T22:00:00Z"
  }
}
```

---

### Step 2: Staff Login (Obtain JWT Token)
```bash
curl -X POST http://localhost/staff/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "doctor_somchai",
    "password": "Password123!",
    "hospital": "hospital-a"
  }'
```

**Response (200 OK):**
```json
{
  "status": "success",
  "message": "Login successful",
  "data": {
    "token": "<YOUR_JWT_TOKEN>",
    "token_type": "Bearer",
    "staff": {
      "id": "270dbcf6-a832-4ae0-b6aa-67c1e54adfa4",
      "username": "doctor_somchai",
      "hospital": "hospital-a"
    }
  }
}
```

---

### Step 3: Search Patients in Same Hospital (Authenticated)
> **Note:** Replace `<YOUR_JWT_TOKEN>` with the token received from Step 2.

```bash
curl -X GET "http://localhost/patient/search?national_id=1100501234567" \
  -H "Authorization: Bearer <YOUR_JWT_TOKEN>"
```

**Response (200 OK):**
```json
{
  "status": "success",
  "count": 1,
  "data": [
    {
      "id": "3bb6ff37-d7d8-4f81-8d0b-33758b29df99",
      "hospital": "hospital-a",
      "patient_hn": "HN-10001",
      "national_id": "1100501234567",
      "passport_id": "AA1234567",
      "first_name_th": "สมชาย",
      "last_name_th": "ใจดี",
      "first_name_en": "Somchai",
      "last_name_en": "Jaidee",
      "date_of_birth": "1990-05-15",
      "phone_number": "0812345678",
      "email": "somchai@example.com",
      "gender": "M"
    }
  ]
}
```

---

### Step 4: Verification of Data Isolation (Cross-Hospital Protection)
When a staff member from `hospital-b` logs in and searches for the same patient (`1100501234567`):
```bash
# Register staff in hospital-b
curl -X POST http://localhost/staff/create \
  -H "Content-Type: application/json" \
  -d '{"username":"doctor_b","password":"Password123!","hospital":"hospital-b"}'

# Login as hospital-b staff
# Then query patient 1100501234567
curl -X GET "http://localhost/patient/search?national_id=1100501234567" \
  -H "Authorization: Bearer <HOSPITAL_B_TOKEN>"
```
**Result:** Returns `{"status":"success","count":0,"data":[]}`.  
The staff member from `hospital-b` cannot access any patient records belonging to `hospital-a`.

---

### Step 5: Unauthenticated Access (Security Negative Test)
```bash
curl -X GET http://localhost/patient/search
```
**Response (401 Unauthorized):**
```json
{
  "status": "error",
  "message": "Authorization token is missing"
}
```

---

## 📂 Seeded Demo Data

For instant evaluation, the database auto-seeds the following on first launch:
* **Hospitals:** `hospital-a`, `hospital-b`
* **Patients:**
  * `HN-10001` (Hospital A): National ID `1100501234567` - "สมชาย ใจดี"
  * `HN-10002` (Hospital A): National ID `1100507654321` - "วิภาดา รักสงบ"
  * `HN-20001` (Hospital B): National ID `1200901234567` - "อนันต์ สุขเจริญ"
