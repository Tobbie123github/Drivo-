# Drivo

A real-time ride-hailing backend built with Go, featuring WebSocket-based live tracking, intelligent driver matching, and a full trip lifecycle.

---

## Tech Stack

| Layer        | Technology                     |
| ------------ | ------------------------------ |
| Language     | Go                             |
| Framework    | Gin                            |
| Database     | PostgreSQL (GORM)              |
| Cache        | Redis                          |
| Real-time    | WebSockets (gorilla/websocket) |
| File Storage | Cloudinary                     |
| Email        | Mailjet                        |
| Auth         | JWT                            |

---

## Getting Started

### Prerequisites

- Go 1.21+
- PostgreSQL
- Redis
- Cloudinary account
- Mailjet account

### Environment Variables

Create a `.env` file in the root:

```env
# Server
PORT=5000

# Database

DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=yourpassword
DB_NAME=drivo

# Redis
REDIS_ADDR=
REDIS_PASSWORD=

# JWT
JWT_SECRET=your_jwt_secret

# Cloudinary
CLOUDINARY_CLOUD_NAME=your_cloud_name
CLOUDINARY_API_KEY=your_api_key
CLOUDINARY_API_SECRET=your_api_secret

# Mailjet
MAILJET_API_KEY=your_mailjet_api_key
MAILJET_SECRET_KEY=your_mailjet_secret_key
MAILJET_FROM_EMAIL=noreply@drivo.com
MAILJET_FROM_NAME=Drivo
```

### Run

```bash
# Install dependencies
go mod tidy

# Run the server
go run cmd/api/main.go
```

---

## API Reference

**Base URL:** `http://localhost:5000`

---

### Authentication

#### Driver Registration

```
POST /auth/driver/register
```

```json
{
  "name": "Tobi",
  "password": "12345678",
  "email": "tobi@gmail.com",
  "phone": "+2348012345678"
}
```

```json
{
  "message": "Please verify your account"
}
```

---

#### Verify Driver Email

```
POST /auth/driver/verify
```

```json
{
  "email": "tobi@gmail.com",
  "otp": "06431"
}
```

```json
{
  "message": "Registration Successful, Proceed to login",
  "data": { ...user }
}
```

---

#### Driver Login

```
POST /auth/driver/login
```

```json
{
  "email": "tobi@gmail.com",
  "password": "12345678"
}
```

```json
{
  "message": {
    "user": { ...user },
    "token": "eyJhbGci..."
  }
}
```

---

#### Rider Registration

```
POST /auth/user/register
```

```json
{
  "name": "John",
  "password": "12345678",
  "email": "john@gmail.com",
  "phone": "+2348087654321"
}
```

---

#### Verify Rider Email

```
POST /auth/user/verify
```

```json
{
  "email": "john@gmail.com",
  "otp": "12345"
}
```

---

#### Rider Login

```
POST /auth/user/login
```

```json
{
  "email": "john@gmail.com",
  "password": "12345678"
}
```

---

### Driver Onboarding

> All onboarding routes require `Authorization: Bearer <token>`

#### Step 1 — Complete Profile

```
PUT /driver/profile
Content-Type: multipart/form-data
```

| Field    | Type                | Required |
| -------- | ------------------- | -------- |
| fullname | string              | required |
| dob      | string (YYYY-MM-DD) | required |
| gender   | string              | required |
| address  | string              | required |
| city     | string              | required |
| state    | string              | required |
| country  | string              | required |
| avatar   | file                | required |

---

#### Step 2 — Upload License

```
PUT /driver/license
Content-Type: multipart/form-data
```

| Field         | Type   | Required |
| ------------- | ------ | -------- |
| licensenumber | string | required |
| licenseexpiry | string | required |
| licenseimage  | file   | required |

---

#### Step 3 — Add Vehicle

```
POST /driver/vehicle
Content-Type: multipart/form-data
```

| Field         | Type   | Required |
| ------------- | ------ | -------- |
| make          | string | required |
| model         | string | required |
| year          | int    | required |
| color         | string | required |
| plate_number  | string | required |
| category      | string | required |
| seats         | int    | required |
| vehicle_image | file   | required |

---

#### Step 4 — Upload Documents

```
POST /driver/documents
Content-Type: multipart/form-data
```

| Field             | Type | Required |
| ----------------- | ---- | -------- |
| national_id_image | file | required |
| selfie_image      | file | required |
| proof_of_address  | file | required |

---

#### Step 5 — Agree to Terms

```
POST /driver/onboarding/complete
```

```json
{
  "agree_terms": true
}
```

```json
{
  "message": "Onboarding Complete, pending account verification"
}
```

---

#### Get Driver Profile

```
GET /driver/profile
Authorization: Bearer <token>
```

```json
{
  "driver": {
    "ID": "a20d3c5c-...",
    "UserID": "deb8af6d-...",
    "DOB": "2003-10-01T00:00:00Z",
    "gender": "male",
    "address": "Plot 5, Omolayo close",
    "city": "Ifo",
    "state": "Ogun State",
    "country": "Nigeria",
    "Status": "pending",
    "LicenseNumber": "DVN-47347",
    "LicenseVerified": false,
    "IsOnline": false,
    "Rating": 5.0,
    "TotalTrips": 0,
    "OnboardingStep": 6,
    "IsOnboardingCompleted": true,
    "Vehicles": []
  }
}
```

---

### Ride

#### Request a Ride

```
POST /ride/request
Authorization: Bearer <rider_token>
```

```json
{
  "pickup_lat": 6.525,
  "pickup_lng": 3.38,
  "dropoff_lat": 6.6,
  "dropoff_lng": 3.41,
  "pickup_address": "Victoria Island, Lagos",
  "dropoff_address": "Lekki Phase 1, Lagos"
}
```

```json
{
  "message": "Ride requested, finding your driver",
  "ride_id": "650ae3b7-...",
  "estimated_fare": 2200,
  "distance_km": 8.97
}
```

---

#### Rate Driver

```
POST /rating/driver
Authorization: Bearer <rider_token>
```

```json
{
  "ride_id": "650ae3b7-...",
  "score": 5,
  "comment": "Very smooth ride!"
}
```

---

#### Rate Rider

```
POST /rating/rider
Authorization: Bearer <driver_token>
```

```json
{
  "ride_id": "650ae3b7-...",
  "score": 4,
  "comment": "Good passenger"
}
```

---

## WebSocket Reference

### Driver WebSocket

```
ws://localhost:5000/ws/driver
Authorization: Bearer <driver_token>
```

The driver connects once when the app opens. The connection stays alive for the entire session. All driver actions are sent as messages with a `type` field.

---

#### Update Location

Sent every 3 seconds by the driver app:

```json
{
  "type": "location_update",
  "payload": {
    "latitude": 6.5244,
    "longitude": 3.3792
  }
}
```

---

#### Accept / Reject Ride

```json
{
  "type": "ride_response",
  "payload": {
    "ride_id": "650ae3b7-...",
    "action": "accept"
  }
}
```

```json
{
  "type": "ride_response",
  "payload": {
    "ride_id": "650ae3b7-...",
    "action": "reject"
  }
}
```

---

#### Driver Arrived at Pickup

```json
{
  "type": "driver_arrived",
  "payload": {
    "ride_id": "650ae3b7-..."
  }
}
```

---

#### Start Trip

```json
{
  "type": "start_trip",
  "payload": {
    "ride_id": "650ae3b7-..."
  }
}
```

---

#### End Trip

```json
{
  "type": "end_trip",
  "payload": {
    "ride_id": "650ae3b7-..."
  }
}
```

---

### Driver — Incoming Messages (Server → Driver)

These are pushed to the driver automatically:

#### Ride Request

```json
{
  "type": "ride_request",
  "payload": {
    "ride_id": "650ae3b7-...",
    "pickup_lat": 6.525,
    "pickup_lng": 3.38,
    "dropoff_lat": 6.6,
    "dropoff_lng": 3.41,
    "pickup_address": "Victoria Island, Lagos",
    "dropoff_address": "Lekki Phase 1, Lagos",
    "estimated_fare": 2200,
    "distance_km": 8.97,
    "rider_name": "John",
    "rider_rating": 4.8
  }
}
```

> Driver has **15 seconds** to accept. If no response, the request moves to the next nearest driver.

---

#### Rate Rider Prompt

```json
{
  "type": "rate_rider",
  "payload": {
    "ride_id": "650ae3b7-...",
    "message": "Rate your rider"
  }
}
```

---

### Rider WebSocket

```
ws://localhost:5000/ws/rider
Authorization: Bearer <rider_token>
```

The rider connects and only **receives** messages. All rider actions (request ride, rate driver) are HTTP endpoints.

---

#### Ride Accepted

```json
{
  "type": "ride_accepted",
  "payload": {
    "ride_id": "650ae3b7-...",
    "driver_name": "Tobi",
    "driver_phone": "+2348012345678",
    "vehicle_make": "Toyota",
    "vehicle_model": "Camry",
    "plate_number": "LAG-123-AB",
    "vehicle_color": "Black",
    "rating": 4.8,
    "eta_minutes": 3
  }
}
```

---

#### Driver Is Here

```json
{
  "type": "driver_is_here",
  "payload": {
    "ride_id": "650ae3b7-...",
    "message": "Your driver has arrived at the pickup point"
  }
}
```

---

#### Trip Started

```json
{
  "type": "ride_started",
  "payload": {
    "ride_id": "650ae3b7-...",
    "message": "Your trip has started"
  }
}
```

---

#### Trip Completed

```json
{
  "type": "ride_completed",
  "payload": {
    "ride_id": "650ae3b7-...",
    "actual_fare": 2200,
    "distance_km": 8.97
  }
}
```

---

#### Rate Driver Prompt

```json
{
  "type": "rate_driver",
  "payload": {
    "ride_id": "650ae3b7-...",
    "message": "How was your trip? Rate your driver"
  }
}
```

## Fare Calculation

Fares are calculated using the **Haversine formula** for straight-line distance between pickup and dropoff:

| Component              | Rate    |
| ---------------------- | ------- |
| Base fare              | ₦500    |
| Per km                 | ₦150    |
| Per minute (estimated) | ₦20     |
| Average speed assumed  | 30 km/h |

Final fare is rounded to the nearest ₦50.

---

## Admin Endpoints

> All admin routes require `Authorization: Bearer <token>` with role `admin`

### Create Admin

Manually update a user's role in the database:
```sql
UPDATE users SET role = 'admin' WHERE email = 'admin@drivo.com';
```

Then login normally via `POST /auth/user/login`

---

### Dashboard Stats
```
GET /admin/stats
```
```json
{
  "stats": {
    "total_drivers": 10,
    "pending_drivers": 3,
    "active_drivers": 6,
    "suspended_drivers": 1,
    "total_riders": 50,
    "total_rides": 120,
    "completed_rides": 100,
    "cancelled_rides": 20,
    "total_earnings": 250000.00,
    "online_drivers": 4
  }
}
```

---

### Drivers

#### Get All Drivers
```
GET /admin/drivers
GET /admin/drivers?status=pending
```
```json
{
  "drivers": [ ...driver objects ],
  "total": 10
}
```

#### Approve Driver
```
PUT /admin/drivers/:id/approve
```
```json
{ "message": "Driver approved" }
```

#### Reject Driver
```
PUT /admin/drivers/:id/reject
```
```json
{ "message": "Driver rejected" }
```

#### Suspend Driver
```
PUT /admin/drivers/:id/suspend
```
```json
{ "message": "Driver suspended" }
```

#### Ban Driver
```
PUT /admin/drivers/:id/ban
```
```json
{ "message": "Driver banned" }
```

#### Verify Driver Identity
```
PUT /admin/drivers/:id/verify-identity
```
```json
{ "message": "Driver identity verified" }
```

#### Verify Driver Vehicle
```
PUT /admin/drivers/:id/verify-vehicle
```
```json
{ "message": "Driver vehicle verified" }
```

#### Verify Driver License
```
PUT /admin/drivers/:id/verify-license
```
```json
{ "message": "Driver license verified" }
```

---

### Riders

#### Get All Riders
```
GET /admin/riders
```
```json
{
  "riders": [ ...rider objects ],
  "total": 50
}
```

---

### Rides

#### Get All Rides
```
GET /admin/rides
GET /admin/rides?status=completed
```
```json
{
  "rides": [ ...ride objects ],
  "total": 120
}
```

---

## Cancellation

### Rider Cancels
```
POST /ride/cancel
Authorization: Bearer <rider_token>
```
```json
{ "ride_id": "650ae3b7-..." }
```
```json
{ "message": "Ride cancelled" }
```

Rider WebSocket receives:
```json
{
  "type": "ride_cancelled_by_rider",
  "payload": {
    "ride_id": "650ae3b7-...",
    "message": "Your ride has been cancelled"
  }
}
```

If driver had already accepted, driver WebSocket receives:
```json
{
  "type": "ride_cancelled_by_rider",
  "payload": {
    "ride_id": "650ae3b7-...",
    "message": "Rider cancelled the ride"
  }
}
```

---

### Driver Cancels
```
POST /ride/driver/cancel
Authorization: Bearer <driver_token>
```
```json
{ "ride_id": "650ae3b7-..." }
```
```json
{ "message": "Ride cancelled" }
```

Rider WebSocket receives:
```json
{
  "type": "ride_cancelled_by_driver",
  "payload": {
    "ride_id": "650ae3b7-...",
    "message": "Your driver cancelled. Finding you a new driver..."
  }
}
```

> If another candidate driver exists in Redis, they automatically receive the ride request. If no candidates remain, the ride is fully cancelled.

---

## Ride History

### Rider History
```
GET /ride/history
Authorization: Bearer <rider_token>
```
```json
{
  "rides": [ ...ride objects ],
  "total": 12
}
```

### Driver History
```
GET /ride/driver/history
Authorization: Bearer <driver_token>
```
```json
{
  "rides": [ ...ride objects ],
  "total": 45
}
```

---

## Rating

### Rider Rates Driver
Triggered automatically after trip ends — rider receives a `rate_driver` WebSocket prompt, then submits via HTTP:
```
POST /rating/driver
Authorization: Bearer <rider_token>
```
```json
{
  "ride_id": "650ae3b7-...",
  "score": 5,
  "comment": "Very smooth ride!"
}
```
```json
{ "message": "Driver rated successfully" }
```

### Driver Rates Rider
```
POST /rating/rider
Authorization: Bearer <driver_token>
```
```json
{
  "ride_id": "650ae3b7-...",
  "score": 4,
  "comment": "Good passenger"
}
```
```json
{ "message": "Rider rated successfully" }
```

---

## Driver Status Flow
```



Recurring ride 








