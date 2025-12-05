# Insurance Broker Platform

AI-Powered Insurance Broker Platform for emerging markets. Multi-tenant system connecting users with insurance providers through intelligent recommendations.

## Tech Stack

### Backend
- **Go 1.22+** with Fiber framework
- **PostgreSQL 16+** with PostGIS extension
- **JWT** authentication with refresh tokens
- **bcrypt** password hashing
- **OpenAI/Anthropic** AI integration

### Frontend
- **Angular 18+** with standalone components
- **Signals** for reactive state management
- **SCSS** for styling
- **Progressive Web App (PWA)** capabilities

### Infrastructure
- **Docker & Docker Compose** for containerization
- **Nginx** for frontend serving
- **pgAdmin 4** for database management

## Project Structure

```
insurance/
├── backend/                 # Go backend API
│   ├── cmd/api/            # Entry point
│   ├── internal/           # Internal packages
│   │   ├── config/        # Configuration management
│   │   ├── database/      # Database connection
│   │   ├── middleware/    # HTTP middleware
│   │   ├── auth/          # Authentication logic
│   │   ├── handlers/      # HTTP handlers
│   │   ├── models/        # Domain models
│   │   ├── repository/    # Data access layer
│   │   └── services/      # Business logic
│   ├── pkg/               # Public packages
│   │   └── utils/         # Utilities (crypto, validation)
│   └── database/migrations/  # SQL migrations
│
├── frontend/              # Angular frontend
│   ├── src/app/
│   │   ├── core/         # Core services, guards, interceptors
│   │   ├── shared/       # Shared components, models
│   │   └── features/     # Feature modules
│   └── src/environments/ # Environment configs
│
└── docker-compose.yml    # Docker orchestration
```

## Getting Started

### Prerequisites
- Docker & Docker Compose
- Node.js 20+ (for local development)
- Go 1.22+ (for local development)

### Quick Start with Docker

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd insurance
   ```

2. **Copy environment file**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

3. **Start all services**
   ```bash
   docker-compose up -d
   ```

4. **Access the services**
   - Frontend: http://localhost:7100
   - Backend API: http://localhost:8096
   - pgAdmin: http://localhost:5050
   - PostgreSQL: localhost:5451

5. **Health check**
   ```bash
   curl http://localhost:8096/health
   ```

### Local Development (Without Docker)

#### Backend Setup

```bash
cd backend

# Install dependencies
go mod download

# Set up environment variables
cp .env.example .env
# Edit .env with your database credentials

# Run database migrations
psql -h localhost -p 5451 -U insurance_admin -d insurance_broker -f database/migrations/001_initial_schema.sql

# Run the server
go run cmd/api/main.go
```

#### Frontend Setup

```bash
cd frontend

# Install dependencies
npm install

# Run development server
npm start

# Build for production
npm run build
```

## Database Schema

The initial migration (`001_initial_schema.sql`) creates:

### Core Tables
- `users` - User authentication and basic info
- `user_profiles` - Extended user information
- `user_addresses` - Multiple user addresses
- `user_documents` - KYC documents
- `user_preferences` - User settings

### Authentication
- `refresh_tokens` - JWT refresh tokens
- `otp_codes` - One-time password codes
- `password_reset_tokens` - Password reset tokens

### Multi-Tenancy
- `insurance_providers` - Insurance companies
- `provider_api_configs` - Provider API integrations
- `user_provider_consents` - GDPR/KDPA compliant consent

### Compliance
- `audit_logs` - Comprehensive audit trail

## API Endpoints

### Public Endpoints
- `GET /` - API information
- `GET /health` - Health check
- `GET /api/v1/public/ping` - Ping test

### Authentication (Coming Soon)
- `POST /api/v1/auth/register` - User registration
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/refresh` - Token refresh

### Protected Endpoints (Coming Soon)
- `GET /api/v1/users/me` - Current user profile
- `GET /api/v1/admin/stats` - Admin statistics

## Security Features

### Backend Security
- ✅ **Password Hashing**: bcrypt with cost 10
- ✅ **JWT Authentication**: Access & refresh tokens
- ✅ **Input Validation**: Email, phone, password validation
- ✅ **SQL Injection Prevention**: Parameterized queries
- ✅ **CORS Configuration**: Restricted origins
- ✅ **Rate Limiting**: Configured per environment
- ✅ **Audit Logging**: All user actions tracked

### Frontend Security
- ✅ **XSS Prevention**: Angular sanitization
- ✅ **CSRF Protection**: Token-based
- ✅ **Secure Storage**: JWT in localStorage
- ✅ **Auto Token Refresh**: Interceptor-based
- ✅ **Route Guards**: Authentication & role-based

### Infrastructure Security
- ✅ **Security Headers**: CSP, X-Frame-Options, etc.
- ✅ **Non-root Containers**: Both backend & frontend
- ✅ **Secrets Management**: Environment variables
- ✅ **Database Encryption**: pgcrypto extension

## Environment Variables

Key environment variables (see `.env.example`):

```env
# Database
DB_HOST=postgres
DB_PORT=5432
DB_NAME=insurance_broker
DB_USER=insurance_admin
DB_PASSWORD=<secure_password>

# JWT
JWT_SECRET=<min_32_chars_secret>
JWT_EXPIRY=24h

# AI Services
OPENAI_API_KEY=<your_key>
ANTHROPIC_API_KEY=<your_key>

# SMS (Africa's Talking)
SMS_API_KEY=<your_key>
SMS_USERNAME=<your_username>
SMS_SENDER_ID=<sender_id>

# Email (SendGrid)
EMAIL_PROVIDER=sendgrid
EMAIL_API_KEY=<your_key>
EMAIL_FROM_ADDRESS=no-reply@insurance.local
EMAIL_FROM_NAME="Insurance Broker AI"

# M-Pesa
MPESA_CONSUMER_KEY=<your_key>
```

## Docker Services

### PostgreSQL
- **Port**: 5451
- **Database**: insurance_broker
- **User**: insurance_admin
- **Health Check**: Automatic

### pgAdmin
- **Port**: 5050
- **Email**: admin@insurance.local
- **Password**: Change in production

### Backend API
- **Port**: 8096
- **Health**: http://localhost:8096/health
- **Auto-restart**: Yes

### Frontend
- **Port**: 7100
- **Live Reload**: Mounted volume
- **Server**: Nginx

## Default Credentials

### Database (Development)
- Username: `insurance_admin`
- Password: `secure_password_change_in_production`

### System Admin (After Migration)
- Email: `admin@insurance.local`
- Password: `admin123` (CHANGE IMMEDIATELY!)

### pgAdmin (Development)
- Email: `admin@insurance.local`
- Password: `admin_password_change_in_production`

## Testing

```bash
# Backend tests
cd backend
go test ./...

# Frontend tests
cd frontend
npm test

# E2E tests
npm run e2e
```

## Deployment

### Production Checklist
- [ ] Change all default passwords
- [ ] Set strong JWT_SECRET (min 32 chars)
- [ ] Configure production database credentials
- [ ] Set up SSL/TLS certificates
- [ ] Enable rate limiting
- [ ] Configure backup strategy
- [ ] Set up monitoring (Prometheus/Grafana)
- [ ] Configure error tracking (Sentry)
- [ ] Review security headers
- [ ] Set CORS origins to production domains
- [ ] Disable debug mode

## Monitoring

Health check endpoints:
- Backend: `GET /health`
- Database connectivity included in health check
- Returns 503 if any service is down

## Roadmap

### Phase 1: Foundation ✅
- [x] Project setup
- [x] Docker configuration
- [x] Database schema
- [x] Authentication scaffolding

### Phase 2: Authentication (In Progress)
- [ ] User registration
- [ ] Email/SMS OTP verification
- [ ] Login with JWT
- [ ] Password reset flow
- [ ] Multi-factor authentication

### Phase 3: User Management (Upcoming)
- [ ] User profile management
- [ ] Document upload (KYC)
- [ ] Address management
- [ ] Preferences & settings

### Phase 4+: See product-description.md

## License

Proprietary - All rights reserved

## Support

For issues and feature requests, please contact the development team.

---

**Built with security and privacy first** 🔒
