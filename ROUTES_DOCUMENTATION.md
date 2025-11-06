# 🛣️ API Routes Documentation

Complete documentation of all API endpoints available in the IPTV platform.

**Last Updated:** January 6, 2025
**Version:** 1.0.0
**Base URL:** `http://localhost:8080/api/v1`

---

## 📋 Table of Contents

1. [Authentication Service](#authentication-service)
2. [User Management](#user-management)
3. [Streaming Service](#streaming-service)
4. [Category Management](#category-management)
5. [Package Management](#package-management)
6. [Billing & Invoicing](#billing--invoicing)
7. [Reseller Management](#reseller-management)
8. [Series & Episodes](#series--episodes)
9. [EPG (Electronic Program Guide)](#epg-electronic-program-guide)
10. [Device Management](#device-management)
11. [Analytics & Statistics](#analytics--statistics)

---

## 🔐 Authentication Service

**Base Path:** `/api/v1/auth`

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/auth/register` | Register new user | ❌ No |
| POST | `/auth/login` | User login | ❌ No |
| POST | `/auth/logout` | User logout | ✅ Yes |
| POST | `/auth/refresh` | Refresh access token | ✅ Yes |
| GET | `/auth/me` | Get current user profile | ✅ Yes |
| PUT | `/auth/profile` | Update user profile | ✅ Yes |
| POST | `/auth/change-password` | Change password | ✅ Yes |
| POST | `/auth/forgot-password` | Request password reset | ❌ No |
| POST | `/auth/reset-password` | Reset password with token | ❌ No |

### Request Examples

```bash
# Register
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "john_doe",
    "email": "john@example.com",
    "password": "SecurePass123!",
    "full_name": "John Doe"
  }'

# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "john_doe",
    "password": "SecurePass123!"
  }'

# Get Current User
curl -X GET http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

---

## 👥 User Management

**Base Path:** `/api/v1/admin/users`

| Method | Endpoint | Description | Auth Required | Role Required |
|--------|----------|-------------|---------------|---------------|
| GET | `/admin/users` | List all users (paginated) | ✅ Yes | Admin |
| GET | `/admin/users/:id` | Get user by ID | ✅ Yes | Admin |
| POST | `/admin/users` | Create new user | ✅ Yes | Admin |
| PUT | `/admin/users/:id` | Update user | ✅ Yes | Admin |
| DELETE | `/admin/users/:id` | Delete user | ✅ Yes | Admin |
| PUT | `/admin/users/:id/status` | Activate/deactivate user | ✅ Yes | Admin |
| PUT | `/admin/users/:id/subscription` | Update subscription | ✅ Yes | Admin |
| GET | `/admin/users/:id/activity` | Get user activity log | ✅ Yes | Admin |

### Query Parameters

**GET /admin/users**
- `page` (int): Page number (default: 1)
- `limit` (int): Items per page (default: 20, max: 100)
- `search` (string): Search by username, email, or full name
- `status` (string): Filter by status (active, inactive, suspended)
- `role` (string): Filter by role (user, reseller, admin)

### Request Examples

```bash
# List Users
curl -X GET "http://localhost:8080/api/v1/admin/users?page=1&limit=20&search=john" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# Create User
curl -X POST http://localhost:8080/api/v1/admin/users \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "new_user",
    "email": "user@example.com",
    "password": "SecurePass123!",
    "role": "user",
    "package_id": 1
  }'

# Update User Status
curl -X PUT http://localhost:8080/api/v1/admin/users/123/status \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"is_active": false}'
```

---

## 📺 Streaming Service

**Base Path:** `/api/v1`

### Live Streams

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/streams` | List all live streams | ✅ Yes |
| GET | `/streams/:id` | Get stream by ID | ✅ Yes |
| POST | `/admin/streams` | Create new stream | ✅ Yes (Admin) |
| PUT | `/admin/streams/:id` | Update stream | ✅ Yes (Admin) |
| DELETE | `/admin/streams/:id` | Delete stream | ✅ Yes (Admin) |
| GET | `/streams/:id/play` | Get stream playback URL | ✅ Yes |
| POST | `/streams/:id/view` | Increment view count | ✅ Yes |

### VOD (Video on Demand)

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/vod` | List all VOD content | ✅ Yes |
| GET | `/vod/:id` | Get VOD by ID | ✅ Yes |
| POST | `/admin/vod` | Create new VOD | ✅ Yes (Admin) |
| PUT | `/admin/vod/:id` | Update VOD | ✅ Yes (Admin) |
| DELETE | `/admin/vod/:id` | Delete VOD | ✅ Yes (Admin) |
| GET | `/vod/:id/play` | Get VOD playback URL | ✅ Yes |

### Query Parameters

**GET /streams**
- `page` (int): Page number
- `limit` (int): Items per page
- `category_id` (int): Filter by category
- `search` (string): Search by name
- `is_active` (bool): Filter by active status
- `featured` (bool): Show only featured streams

### Request Examples

```bash
# List Streams
curl -X GET "http://localhost:8080/api/v1/streams?category_id=1&featured=true" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# Create Stream
curl -X POST http://localhost:8080/api/v1/admin/streams \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "HBO Live",
    "category_id": 1,
    "stream_url": "https://cdn.example.com/hbo/playlist.m3u8",
    "icon_url": "https://cdn.example.com/icons/hbo.png",
    "is_featured": true
  }'

# Get Stream Playback
curl -X GET http://localhost:8080/api/v1/streams/123/play \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

---

## 📁 Category Management

**Base Path:** `/api/v1/admin/categories`

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/categories` | List all categories | ❌ No |
| GET | `/categories/:id` | Get category by ID | ❌ No |
| POST | `/admin/categories` | Create new category | ✅ Yes (Admin) |
| PUT | `/admin/categories/:id` | Update category | ✅ Yes (Admin) |
| DELETE | `/admin/categories/:id` | Delete category | ✅ Yes (Admin) |
| GET | `/categories/:id/streams` | Get streams in category | ✅ Yes |

### Request Examples

```bash
# List Categories
curl -X GET http://localhost:8080/api/v1/categories

# Create Category
curl -X POST http://localhost:8080/api/v1/admin/categories \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Sports",
    "description": "Live sports channels",
    "icon": "🏈"
  }'
```

---

## 💎 Package Management

**Base Path:** `/api/v1/admin/packages`

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/packages` | List all packages | ❌ No |
| GET | `/packages/:id` | Get package by ID | ❌ No |
| POST | `/admin/packages` | Create new package | ✅ Yes (Admin) |
| PUT | `/admin/packages/:id` | Update package | ✅ Yes (Admin) |
| DELETE | `/admin/packages/:id` | Delete package | ✅ Yes (Admin) |
| POST | `/admin/packages/:id/streams` | Add streams to package | ✅ Yes (Admin) |
| DELETE | `/admin/packages/:id/streams/:stream_id` | Remove stream from package | ✅ Yes (Admin) |

### Request Examples

```bash
# List Packages
curl -X GET http://localhost:8080/api/v1/packages

# Create Package
curl -X POST http://localhost:8080/api/v1/admin/packages \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Premium IPTV",
    "description": "Access to all channels",
    "price": 29.99,
    "duration_days": 30,
    "max_connections": 2,
    "is_active": true
  }'
```

---

## 💳 Billing & Invoicing

**Base Path:** `/api/v1/admin/billing`

### Invoices

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/admin/billing/invoices` | List all invoices | ✅ Yes (Admin) |
| GET | `/admin/billing/invoices/:id` | Get invoice by ID | ✅ Yes (Admin) |
| POST | `/admin/billing/invoices` | Create new invoice | ✅ Yes (Admin) |
| PUT | `/admin/billing/invoices/:id` | Update invoice | ✅ Yes (Admin) |
| DELETE | `/admin/billing/invoices/:id` | Delete invoice | ✅ Yes (Admin) |
| POST | `/admin/billing/invoices/:id/cancel` | Cancel invoice | ✅ Yes (Admin) |
| GET | `/billing/invoices` | Get user's invoices | ✅ Yes (User) |

### Payments

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/admin/billing/payments` | Process payment | ✅ Yes (Admin) |
| GET | `/admin/billing/transactions` | List transactions | ✅ Yes (Admin) |
| POST | `/admin/billing/transactions/:id/refund` | Refund payment | ✅ Yes (Admin) |

### Payment Methods

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/admin/billing/payment-methods` | List payment methods | ✅ Yes (Admin) |
| POST | `/admin/billing/payment-methods` | Add payment method | ✅ Yes (Admin) |
| DELETE | `/admin/billing/payment-methods/:id` | Delete payment method | ✅ Yes (Admin) |
| POST | `/admin/billing/payment-methods/:id/set-default` | Set default payment method | ✅ Yes (Admin) |

### Revenue Statistics

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/admin/billing/revenue/stats` | Get revenue statistics | ✅ Yes (Admin) |

### Query Parameters

**GET /admin/billing/invoices**
- `page` (int): Page number
- `limit` (int): Items per page
- `status` (string): Filter by status (pending, paid, overdue, cancelled, refunded)
- `user_id` (int): Filter by user
- `search` (string): Search by invoice number or username

### Request Examples

```bash
# List Invoices
curl -X GET "http://localhost:8080/api/v1/admin/billing/invoices?status=pending" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# Create Invoice
curl -X POST http://localhost:8080/api/v1/admin/billing/invoices \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 123,
    "package_id": 1,
    "description": "Monthly subscription",
    "subtotal": 29.99,
    "tax": 2.70,
    "discount": 0,
    "due_date": "2025-01-20"
  }'

# Process Payment
curl -X POST http://localhost:8080/api/v1/admin/billing/payments \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "invoice_id": 456,
    "payment_gateway": "stripe",
    "amount": 32.69,
    "payment_details": {
      "card_number": "4242424242424242",
      "exp_month": "12",
      "exp_year": "2025",
      "cvc": "123"
    }
  }'

# Get Revenue Stats
curl -X GET http://localhost:8080/api/v1/admin/billing/revenue/stats \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

---

## 👔 Reseller Management

**Base Path:** `/api/v1/admin/resellers`

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/admin/resellers` | List all resellers | ✅ Yes (Admin) |
| GET | `/admin/resellers/:id` | Get reseller by ID | ✅ Yes (Admin) |
| POST | `/admin/resellers` | Create new reseller | ✅ Yes (Admin) |
| PUT | `/admin/resellers/:id` | Update reseller | ✅ Yes (Admin) |
| DELETE | `/admin/resellers/:id` | Delete reseller | ✅ Yes (Admin) |
| POST | `/admin/resellers/:id/credits` | Add credits | ✅ Yes (Admin) |
| POST | `/admin/resellers/:id/credits/deduct` | Deduct credits | ✅ Yes (Admin) |
| GET | `/admin/resellers/:id/customers` | Get reseller's customers | ✅ Yes (Admin) |
| POST | `/admin/resellers/:id/customers` | Assign customer to reseller | ✅ Yes (Admin) |
| GET | `/admin/resellers/:id/stats` | Get reseller statistics | ✅ Yes (Admin) |
| GET | `/admin/resellers/:id/credits-log` | Get credits transaction log | ✅ Yes (Admin) |

### Query Parameters

**GET /admin/resellers**
- `page` (int): Page number
- `limit` (int): Items per page
- `search` (string): Search by username
- `parent_id` (int): Filter by parent reseller

### Request Examples

```bash
# List Resellers
curl -X GET "http://localhost:8080/api/v1/admin/resellers?page=1&limit=20" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# Create Reseller
curl -X POST http://localhost:8080/api/v1/admin/resellers \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 123,
    "parent_id": null,
    "credits": 1000.00,
    "commission_rate": 10.0,
    "can_create_resellers": true,
    "max_users": 100,
    "max_resellers": 10
  }'

# Add Credits
curl -X POST http://localhost:8080/api/v1/admin/resellers/456/credits \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "amount": 500.00,
    "description": "Monthly credit allocation"
  }'

# Get Reseller Stats
curl -X GET http://localhost:8080/api/v1/admin/resellers/456/stats \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

---

## 🎬 Series & Episodes

**Base Path:** `/api/v1/admin`

### Series

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/admin/series` | List all series | ✅ Yes (Admin) |
| GET | `/admin/series/:id` | Get series by ID | ✅ Yes (Admin) |
| POST | `/admin/series` | Create new series | ✅ Yes (Admin) |
| PUT | `/admin/series/:id` | Update series | ✅ Yes (Admin) |
| DELETE | `/admin/series/:id` | Delete series | ✅ Yes (Admin) |
| GET | `/admin/series/:id/seasons` | Get season info | ✅ Yes (Admin) |

### Episodes

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/admin/episodes` | List episodes | ✅ Yes (Admin) |
| GET | `/admin/episodes/:id` | Get episode by ID | ✅ Yes (Admin) |
| POST | `/admin/episodes` | Create new episode | ✅ Yes (Admin) |
| PUT | `/admin/episodes/:id` | Update episode | ✅ Yes (Admin) |
| DELETE | `/admin/episodes/:id` | Delete episode | ✅ Yes (Admin) |
| POST | `/episodes/:id/view` | Increment view count | ✅ Yes (User) |

### Query Parameters

**GET /admin/series**
- `page` (int): Page number
- `limit` (int): Items per page
- `category_id` (int): Filter by category
- `search` (string): Search by name or description
- `featured` (bool): Show only featured series

**GET /admin/episodes**
- `series_id` (int): Filter by series (required)
- `season` (int): Filter by season number

### Request Examples

```bash
# List Series
curl -X GET "http://localhost:8080/api/v1/admin/series?featured=true" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# Create Series
curl -X POST http://localhost:8080/api/v1/admin/series \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Breaking Bad",
    "description": "A high school chemistry teacher turned methamphetamine producer",
    "category_id": 1,
    "cover_url": "https://cdn.example.com/covers/breaking-bad.jpg",
    "rating": 9.5,
    "release_year": 2008,
    "genre": "Crime, Drama, Thriller",
    "cast": ["Bryan Cranston", "Aaron Paul"],
    "director": "Vince Gilligan",
    "is_featured": true
  }'

# List Episodes
curl -X GET "http://localhost:8080/api/v1/admin/episodes?series_id=123&season=1" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# Create Episode
curl -X POST http://localhost:8080/api/v1/admin/episodes \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "series_id": 123,
    "season": 1,
    "episode": 1,
    "title": "Pilot",
    "description": "Walter White is diagnosed with cancer",
    "stream_url": "https://cdn.example.com/bb/s01e01.m3u8",
    "duration": 3480,
    "air_date": "2008-01-20",
    "rating": 9.0
  }'
```

---

## 📅 EPG (Electronic Program Guide)

**Base Path:** `/api/v1/admin/epg`

### EPG Programs

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/admin/epg/programs` | List EPG programs | ✅ Yes (Admin) |
| GET | `/admin/epg/programs/:id` | Get program by ID | ✅ Yes (Admin) |
| POST | `/admin/epg/programs` | Create new program | ✅ Yes (Admin) |
| PUT | `/admin/epg/programs/:id` | Update program | ✅ Yes (Admin) |
| DELETE | `/admin/epg/programs/:id` | Delete program | ✅ Yes (Admin) |
| GET | `/epg/current` | Get current programs | ✅ Yes (User) |
| GET | `/epg/schedule` | Get program schedule | ✅ Yes (User) |

### EPG Sources

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/admin/epg/sources` | List EPG sources | ✅ Yes (Admin) |
| GET | `/admin/epg/sources/:id` | Get source by ID | ✅ Yes (Admin) |
| POST | `/admin/epg/sources` | Create new source | ✅ Yes (Admin) |
| PUT | `/admin/epg/sources/:id` | Update source | ✅ Yes (Admin) |
| DELETE | `/admin/epg/sources/:id` | Delete source | ✅ Yes (Admin) |
| POST | `/admin/epg/sources/:id/sync` | Trigger EPG import | ✅ Yes (Admin) |

### EPG Import

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/admin/epg/import/xmltv` | Import XMLTV file | ✅ Yes (Admin) |
| GET | `/admin/epg/import/history` | Get import history | ✅ Yes (Admin) |
| GET | `/admin/epg/import/:id` | Get import details | ✅ Yes (Admin) |

### User Reminders

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/epg/reminders` | Get user reminders | ✅ Yes (User) |
| POST | `/epg/reminders` | Set program reminder | ✅ Yes (User) |
| DELETE | `/epg/reminders/:id` | Delete reminder | ✅ Yes (User) |

### Query Parameters

**GET /admin/epg/programs**
- `stream_id` (int): Filter by stream
- `start_date` (string): Start date (YYYY-MM-DD)
- `end_date` (string): End date (YYYY-MM-DD)
- `category` (string): Filter by category

**GET /epg/schedule**
- `stream_id` (int): Filter by stream
- `date` (string): Date (YYYY-MM-DD)

---

## 📱 Device Management

**Base Path:** `/api/v1/admin/devices`

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/admin/devices` | List all devices | ✅ Yes (Admin) |
| GET | `/admin/devices/:id` | Get device by ID | ✅ Yes (Admin) |
| POST | `/admin/devices` | Register new device | ✅ Yes (Admin) |
| PUT | `/admin/devices/:id` | Update device | ✅ Yes (Admin) |
| DELETE | `/admin/devices/:id` | Delete device | ✅ Yes (Admin) |
| POST | `/admin/devices/:id/block` | Block device | ✅ Yes (Admin) |
| POST | `/admin/devices/:id/unblock` | Unblock device | ✅ Yes (Admin) |
| GET | `/admin/devices/:id/sessions` | Get device sessions | ✅ Yes (Admin) |
| POST | `/admin/devices/:id/sessions/terminate` | Terminate all sessions | ✅ Yes (Admin) |

### Query Parameters

**GET /admin/devices**
- `user_id` (int): Filter by user
- `device_type` (string): Filter by type (mag, enigma2, android, ios, web)
- `status` (string): Filter by status (active, blocked)

---

## 📊 Analytics & Statistics

**Base Path:** `/api/v1/admin/analytics`

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/admin/analytics/dashboard` | Get dashboard statistics | ✅ Yes (Admin) |
| GET | `/admin/analytics/users` | User analytics | ✅ Yes (Admin) |
| GET | `/admin/analytics/streams` | Stream analytics | ✅ Yes (Admin) |
| GET | `/admin/analytics/revenue` | Revenue analytics | ✅ Yes (Admin) |
| GET | `/admin/analytics/top-streams` | Most watched streams | ✅ Yes (Admin) |
| GET | `/admin/analytics/active-users` | Active users stats | ✅ Yes (Admin) |

---

## 📈 Route Summary

### Total Endpoints by Service

| Service | Admin Routes | User Routes | Public Routes | Total |
|---------|--------------|-------------|---------------|-------|
| Authentication | 0 | 6 | 3 | 9 |
| User Management | 8 | 0 | 0 | 8 |
| Streaming | 6 | 4 | 0 | 10 |
| Categories | 3 | 2 | 2 | 7 |
| Packages | 5 | 0 | 2 | 7 |
| Billing | 14 | 1 | 0 | 15 |
| Resellers | 11 | 0 | 0 | 11 |
| Series/Episodes | 11 | 1 | 0 | 12 |
| EPG | 12 | 4 | 0 | 16 |
| Devices | 9 | 0 | 0 | 9 |
| Analytics | 6 | 0 | 0 | 6 |
| **TOTAL** | **85** | **18** | **7** | **110** |

---

## 🔒 Authentication Headers

All authenticated requests must include:

```bash
Authorization: Bearer YOUR_JWT_TOKEN
```

---

## 📝 Response Format

### Success Response

```json
{
  "message": "Operation successful",
  "data": { ... }
}
```

### Error Response

```json
{
  "error": "Error description",
  "details": "Detailed error message"
}
```

### Paginated Response

```json
{
  "data": [ ... ],
  "pagination": {
    "total": 150,
    "page": 1,
    "limit": 20,
    "total_pages": 8
  }
}
```

---

## 🚀 Testing Routes

### Quick Test Commands

```bash
# Test Authentication
./scripts/test-auth.sh

# Test Billing System
./scripts/test-billing.sh

# Test Reseller System
./scripts/test-resellers.sh

# Test Streaming
./scripts/test-streaming.sh

# Test All Routes
./scripts/test-all-routes.sh
```

---

## 📚 Additional Resources

- [API Postman Collection](./postman/iptv-platform.json)
- [Authentication Guide](./docs/AUTH_GUIDE.md)
- [Error Codes Reference](./docs/ERROR_CODES.md)
- [Rate Limiting](./docs/RATE_LIMITS.md)

---

**Note:** This documentation is automatically generated and kept in sync with the codebase.
