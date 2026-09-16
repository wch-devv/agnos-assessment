# API Specification - Hospital Middleware System

This document outlines the complete RESTful API specification for the **Hospital Middleware System**, fulfilling **Deliverable 1.b** of the Agnos Candidate Assignment.

Base URL (Direct API): `http://localhost:8080/api/v1`  
Base URL (Via Nginx Reverse Proxy): `http://localhost/api/v1`  

---

## 1. Authentication Overview
Endpoints requiring authentication must include the JWT token in the HTTP `Authorization` request header:
```http
Authorization: Bearer <jwt_token>
```
Tokens are signed using `HMAC-SHA256` and encode `staff_id`, `username`, `hospital`, and expiry (`exp`).

---

## 2. API Endpoints

### 2.1 Create Hospital Staff
* **Endpoint:** `POST /api/v1/staff/create`
* **Public:** Yes (No authentication required)
* **Description:** Registers a new hospital staff member with login credentials.

#### Request Headers
| Header | Value |
|---|---|
| `Content-Type` | `application/json` |

#### Request Body
```json
{
  "username": "doctor_somchai",
  "password": "Password123!",
  "hospital": "hospital-a"
}
```

#### Fields Validation
| Field | Type | Required | Notes |
|---|---|---|---|
| `username` | string | Yes | 3-50 characters |
| `password` | string | Yes | Min 6 characters |
| `hospital` | string | Yes | Hospital identifier |

#### Responses
* **201 Created**
  ```json
  {
    "status": "success",
    "message": "Staff member created successfully",
    "data": {
      "id": "c7a8b4e2-6d2e-4cf7-bb4f-4d3b6f28a301",
      "username": "doctor_somchai",
      "hospital": "hospital-a",
      "created_at": "2026-09-16T22:00:00Z"
    }
  }
  ```
* **400 Bad Request** (Validation failure)
  ```json
  {
    "status": "error",
    "message": "Invalid input: username and hospital are required"
  }
  ```
* **409 Conflict** (Username already exists in this hospital)
  ```json
  {
    "status": "error",
    "message": "Username already exists in this hospital"
  }
  ```

---

### 2.2 Staff Login
* **Endpoint:** `POST /api/v1/staff/login`
* **Public:** Yes
* **Description:** Authenticates staff credentials and issues a JWT token.

#### Request Body
```json
{
  "username": "doctor_somchai",
  "password": "Password123!",
  "hospital": "hospital-a"
}
```

#### Responses
* **200 OK**
  ```json
  {
    "status": "success",
    "message": "Login successful",
    "data": {
      "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
      "token_type": "Bearer",
      "expires_in": 86400,
      "staff": {
        "id": "c7a8b4e2-6d2e-4cf7-bb4f-4d3b6f28a301",
        "username": "doctor_somchai",
        "hospital": "hospital-a"
      }
    }
  }
  ```
* **401 Unauthorized** (Invalid credentials or hospital mismatch)
  ```json
  {
    "status": "error",
    "message": "Invalid username, password, or hospital"
  }
  ```

---

### 2.3 Search Patients
* **Endpoint:** `GET /api/v1/patient/search`
* **Authentication:** **Required** (`Bearer <token>`)
* **Description:** Searches patients matching the criteria. Results are **strictly scoped to the hospital of the authenticated staff member**.

#### Query Parameters (All Optional)
| Parameter | Type | Example | Description |
|---|---|---|---|
| `national_id` | string | `1100501234567` | 13-digit Thai National ID |
| `passport_id` | string | `AA1234567` | Passport number |
| `first_name` | string | `Somchai` | Searches both Thai and English first names |
| `middle_name`| string | `John` | Searches both Thai and English middle names |
| `last_name` | string | `Jaidee` | Searches both Thai and English last names |
| `date_of_birth`| string | `1990-05-15` | Date of birth (YYYY-MM-DD) |
| `phone_number`| string | `0812345678` | Contact phone number |
| `email` | string | `somchai@example.com` | Email address |

#### Responses
* **200 OK**
  ```json
  {
    "status": "success",
    "count": 1,
    "data": [
      {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "hospital": "hospital-a",
        "patient_hn": "HN-100234",
        "national_id": "1100501234567",
        "passport_id": "AA1234567",
        "first_name_th": "สมชาย",
        "middle_name_th": "",
        "last_name_th": "ใจดี",
        "first_name_en": "Somchai",
        "middle_name_en": "",
        "last_name_en": "Jaidee",
        "date_of_birth": "1990-05-15",
        "phone_number": "0812345678",
        "email": "somchai@example.com",
        "gender": "M",
        "created_at": "2026-09-16T22:00:00Z",
        "updated_at": "2026-09-16T22:00:00Z"
      }
    ]
  }
  ```
* **401 Unauthorized** (Missing or invalid Bearer token)
  ```json
  {
    "status": "error",
    "message": "Authorization token is missing or invalid"
  }
  ```

---

## 3. External HIS API (Hospital A Reference)
* **Route:** `GET https://hospital-a.api.co.th/patient/search/{id}`
* **Parameter:** `id` can be either `national_id` or `passport_id`.
* The middleware system integrates with this endpoint via `HISClient` interface, with automatic fallback and caching to local PostgreSQL database.
