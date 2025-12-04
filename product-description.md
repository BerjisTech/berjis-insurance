# AI-Powered Insurance Broker Platform - Product Roadmap

## Product Vision
A multi-tenant AI-powered insurance broker platform that connects users with multiple insurance providers, enabling intelligent policy recommendations, comparison, and management. Designed for emerging markets with mobile-first approach, offline capability considerations, and support for diverse payment methods including mobile money.

## Technical Stack
- **Frontend**: Angular 18+ (standalone components, signals)
- **Backend**: Go 1.22+ (fiber/gin framework)
- **Database**: PostgreSQL 16+ with PostGIS
- **AI/ML**: OpenAI API / Anthropic Claude API for intelligent recommendations
- **Infrastructure**: Docker, Kubernetes-ready
- **File Structure**: Strict separation - .html, .ts, .css, .go files

## Core Principles
- **Privacy-First**: Zero-knowledge architecture for sensitive data
- **Consent-Driven**: Explicit user consent for all data sharing
- **Mobile-First**: Progressive Web App (PWA) capabilities
- **Emerging Market Optimized**: Low bandwidth mode, mobile money integration, multi-language support
- **Modular Architecture**: Small, single-responsibility files (<300 lines)
- **Extensive Documentation**: Every file heavily commented

---

## Phase 1: Foundation & Core Infrastructure

### 1.1 Project Setup & Architecture
**Goal**: Establish development environment and project structure

**Backend (Go)**
```
project-root/
├── cmd/
│   └── api/
│       └── main.go              # Entry point, server initialization
├── internal/
│   ├── config/
│   │   └── config.go           # Environment and configuration management
│   ├── database/
│   │   ├── connection.go       # PostgreSQL connection pool
│   │   └── migrations.go       # Migration runner
│   └── middleware/
│       ├── auth.go             # JWT authentication middleware
│       ├── cors.go             # CORS configuration
│       └── logger.go           # Request logging middleware
├── pkg/
│   └── utils/
│       ├── crypto.go           # Encryption utilities
│       └── validator.go        # Input validation
└── go.mod
```

**Frontend (Angular)**
```
src/
├── app/
│   ├── core/
│   │   ├── services/
│   │   │   ├── auth.service.ts       # Authentication service
│   │   │   ├── api.service.ts        # HTTP client wrapper
│   │   │   └── storage.service.ts    # LocalStorage/IndexedDB wrapper
│   │   ├── guards/
│   │   │   └── auth.guard.ts         # Route protection
│   │   └── interceptors/
│   │       └── auth.interceptor.ts   # JWT token injection
│   ├── shared/
│   │   ├── components/              # Reusable UI components
│   │   └── models/                  # TypeScript interfaces
│   └── features/                    # Feature modules (created per phase)
├── styles/
│   ├── _variables.scss              # Design tokens
│   ├── _mixins.scss                 # Reusable style patterns
│   └── main.scss                    # Global styles
└── environments/
```

**Comments Standard**:
- File header: Purpose, dependencies, and usage examples
- Function/method: Parameters, return values, side effects
- Complex logic: Step-by-step explanation
- API endpoints: Request/response examples

### 1.2 Database Schema Design
**Goal**: Design multi-tenant database with strict data isolation

**Tables to Create**:
```sql
-- Core entities with row-level security
users (id, email, phone, password_hash, role, created_at, updated_at)
insurance_providers (id, name, license_number, country_code, status, api_config)
user_provider_consents (user_id, provider_id, consent_type, granted_at, expires_at)

-- Encryption keys table (for field-level encryption)
encryption_keys (entity_id, entity_type, key_hash, created_at)

-- Audit log for compliance
audit_logs (id, user_id, action, entity_type, entity_id, timestamp, ip_address)
```

**Implementation**:
- PostgreSQL Row-Level Security (RLS) policies
- Separate schemas per tenant for hard isolation
- Encrypted columns for PII (using pgcrypto)
- Composite indexes on foreign keys
- Database migration files (numbered: 001_initial_schema.sql)

### 1.3 Authentication & Authorization System
**Goal**: Secure multi-tenant authentication

**Features**:
- JWT-based authentication with refresh tokens
- Role-based access control (RBAC): User, Broker, Provider Admin, System Admin
- Multi-factor authentication (SMS/Email OTP)
- Phone number verification (Kenya +254 support)
- Password reset flow
- Session management

**Backend Files**:
- `internal/auth/jwt.go` - Token generation/validation
- `internal/auth/otp.go` - OTP generation/verification
- `internal/auth/password.go` - Password hashing (bcrypt)
- `internal/handlers/auth_handler.go` - Login/register/refresh endpoints

**Frontend Files**:
- `features/auth/login/login.component.ts|html|css`
- `features/auth/register/register.component.ts|html|css`
- `features/auth/verify-otp/verify-otp.component.ts|html|css`

---

## Phase 2: User Management & Profiles

### 2.1 User Profile System
**Goal**: Comprehensive user profile management

**Features**:
- Personal information management
- KYC document upload (ID, proof of address)
- Profile photo
- Communication preferences
- Data export functionality (GDPR compliance)

**Database Tables**:
```sql
user_profiles (user_id, full_name, date_of_birth, gender, occupation, id_number)
user_addresses (user_id, type, street, city, postal_code, country, is_primary)
user_documents (user_id, document_type, file_url, verification_status, uploaded_at)
user_preferences (user_id, language, currency, notification_channels, timezone)
```

**Backend Files**:
- `internal/models/user.go` - User domain models
- `internal/repository/user_repository.go` - Database operations
- `internal/services/user_service.go` - Business logic
- `internal/handlers/user_handler.go` - HTTP endpoints

**Frontend Files**:
- `features/profile/view-profile/` - Profile display
- `features/profile/edit-profile/` - Profile editing
- `features/profile/upload-documents/` - Document management
- `shared/components/avatar-upload/` - Reusable avatar component

### 2.2 Multi-Tenancy & Provider Onboarding
**Goal**: Allow insurance providers to join and manage their presence

**Features**:
- Provider registration workflow
- License verification
- Provider profile (logo, description, supported products)
- API integration setup (for automated quote fetching)
- Provider admin dashboard

**Database Tables**:
```sql
provider_profiles (provider_id, logo_url, description, website, support_phone, support_email)
provider_licenses (provider_id, license_type, license_number, issuing_authority, expiry_date)
provider_products (id, provider_id, product_type, name, description, is_active)
provider_api_configs (provider_id, api_type, endpoint_url, auth_method, credentials_encrypted)
```

**Backend Files**:
- `internal/models/provider.go`
- `internal/services/provider_service.go`
- `internal/services/license_verification_service.go` - External API integration
- `internal/handlers/provider_handler.go`

**Frontend Files**:
- `features/provider/onboarding/` - Multi-step registration wizard
- `features/provider/dashboard/` - Provider admin panel
- `features/provider/product-catalog/` - Manage insurance products

---

## Phase 3: Insurance Product Catalog

### 3.1 Product Management System
**Goal**: Comprehensive insurance product catalog

**Features**:
- Product categories (Auto, Health, Life, Property, Travel, Agriculture)
- Product variants and riders
- Coverage details and exclusions
- Premium calculation rules
- Product comparison matrix

**Database Tables**:
```sql
product_categories (id, name, icon, description, sort_order)
insurance_products (id, provider_id, category_id, name, description, min_age, max_age, coverage_amount_min, coverage_amount_max)
product_features (product_id, feature_name, feature_value, is_highlight)
product_exclusions (product_id, exclusion_text)
product_riders (id, product_id, name, description, additional_premium_percent)
premium_calculation_rules (product_id, rule_type, rule_config_json)
```

**Backend Files**:
- `internal/models/product.go`
- `internal/repository/product_repository.go`
- `internal/services/product_service.go`
- `internal/services/premium_calculator.go` - Premium calculation engine
- `internal/handlers/product_handler.go`

**Frontend Files**:
- `features/products/product-list/` - Browse products with filters
- `features/products/product-detail/` - Detailed product view
- `features/products/product-comparison/` - Side-by-side comparison
- `shared/components/product-card/` - Reusable product display

### 3.2 Search & Filtering Engine
**Goal**: Intelligent product discovery

**Features**:
- Full-text search (PostgreSQL tsvector)
- Advanced filtering (price range, coverage, provider)
- Sorting (relevance, price, rating)
- Recently viewed products
- Saved searches

**Backend Files**:
- `internal/services/search_service.go` - Search implementation
- `internal/repository/search_repository.go` - Optimized queries
- Database migration for full-text search indexes

**Frontend Files**:
- `features/search/search-bar/` - Autocomplete search
- `features/search/search-results/` - Results display
- `features/search/filters/` - Dynamic filter sidebar

---

## Phase 4: AI-Powered Recommendation Engine

### 4.1 User Profiling & Risk Assessment
**Goal**: Understand user needs and risk profile

**Features**:
- Interactive questionnaire (lifestyle, health, assets)
- Risk tolerance assessment
- Life event detection (marriage, childbirth, home purchase)
- Financial profile (income, dependents, existing coverage)

**Database Tables**:
```sql
user_risk_profiles (user_id, risk_score, health_score, lifestyle_score, calculated_at)
user_questionnaire_responses (user_id, question_id, answer, answered_at)
user_life_events (user_id, event_type, event_date, impact_score)
user_financial_profiles (user_id, monthly_income, dependents, existing_coverage_total, debt_amount)
```

**Backend Files**:
- `internal/models/risk_profile.go`
- `internal/services/risk_assessment_service.go`
- `internal/services/questionnaire_service.go`
- `pkg/scoring/risk_calculator.go` - Risk scoring algorithms

**Frontend Files**:
- `features/onboarding/questionnaire/` - Step-by-step questionnaire
- `features/onboarding/life-events/` - Life event tracker
- `features/profile/risk-profile/` - Risk profile visualization

### 4.2 AI Recommendation System
**Goal**: Intelligent insurance recommendations

**Features**:
- Personalized product recommendations
- Gap analysis (coverage gaps in user's portfolio)
- Budget optimization
- Natural language explanation of recommendations
- Recommendation confidence scores

**Backend Files**:
- `internal/services/ai_recommendation_service.go` - AI integration
- `internal/services/recommendation_engine.go` - Core recommendation logic
- `internal/clients/openai_client.go` - OpenAI API wrapper
- `pkg/ai/prompt_templates.go` - AI prompt management
- `pkg/ai/recommendation_scorer.go` - Confidence scoring

**AI Integration**:
```go
// Example prompt structure
recommendationPrompt := `
User Profile:
- Age: {age}
- Occupation: {occupation}
- Dependents: {dependents}
- Monthly Income: {income}
- Risk Tolerance: {risk_tolerance}
- Existing Coverage: {existing_policies}

Available Products:
{product_catalog_json}

Task: Recommend top 5 insurance products with detailed reasoning.
Output format: JSON array with product_id, recommendation_score, reasoning.
`
```

**Frontend Files**:
- `features/recommendations/recommendations-dashboard/` - AI recommendations view
- `features/recommendations/gap-analysis/` - Coverage gap visualization
- `shared/components/recommendation-card/` - Recommendation display

### 4.3 Conversational AI Assistant
**Goal**: Natural language insurance advice

**Features**:
- Chat interface for insurance queries
- Policy explanation in simple terms
- Quote comparison assistance
- Claims guidance
- Conversation history

**Database Tables**:
```sql
chat_conversations (id, user_id, started_at, last_message_at)
chat_messages (id, conversation_id, sender_type, message_text, ai_context_json, timestamp)
```

**Backend Files**:
- `internal/services/chatbot_service.go`
- `internal/clients/claude_client.go` - Anthropic API wrapper
- `pkg/ai/context_manager.go` - Conversation context management

**Frontend Files**:
- `features/assistant/chat-widget/` - Floating chat interface
- `features/assistant/chat-history/` - Conversation history

---

## Phase 5: Quote & Policy Management

### 5.1 Quote Generation System
**Goal**: Get insurance quotes from multiple providers

**Features**:
- Multi-provider quote requests
- Real-time quote comparison
- Quote validity tracking
- Quote history
- Save quotes for later

**Database Tables**:
```sql
quote_requests (id, user_id, product_category, coverage_details_json, requested_at, status)
quotes (id, quote_request_id, provider_id, product_id, premium_amount, coverage_amount, validity_period, quote_document_url)
quote_comparisons (id, user_id, quote_ids_array, created_at)
```

**Backend Files**:
- `internal/models/quote.go`
- `internal/services/quote_service.go`
- `internal/services/quote_aggregator.go` - Multi-provider quote fetching
- `internal/services/quote_comparison_service.go`
- `internal/adapters/provider_adapter.go` - Provider API adapters

**Frontend Files**:
- `features/quotes/request-quote/` - Quote request form
- `features/quotes/quote-list/` - View all quotes
- `features/quotes/quote-comparison/` - Compare quotes side-by-side
- `shared/components/quote-card/` - Quote display

### 5.2 Policy Purchase & Onboarding
**Goal**: Seamless policy purchase experience

**Features**:
- Policy application workflow
- Document upload (medical reports, vehicle photos)
- E-signature integration
- Payment gateway integration (M-Pesa, card payments)
- Policy document generation
- Policy activation

**Database Tables**:
```sql
policy_applications (id, user_id, quote_id, application_status, submitted_at)
policy_documents (application_id, document_type, file_url, verification_status)
policies (id, user_id, provider_id, product_id, policy_number, start_date, end_date, premium_amount, coverage_amount, status)
policy_payments (id, policy_id, payment_amount, payment_method, transaction_id, payment_date, status)
```

**Backend Files**:
- `internal/models/policy.go`
- `internal/services/policy_service.go`
- `internal/services/payment_service.go`
- `internal/clients/mpesa_client.go` - M-Pesa STK Push integration
- `internal/clients/payment_gateway.go` - Card payment integration
- `pkg/documents/pdf_generator.go` - Policy document generation

**Frontend Files**:
- `features/policies/apply/` - Application wizard
- `features/policies/payment/` - Payment options
- `features/policies/confirmation/` - Purchase confirmation
- `shared/components/payment-methods/` - Payment selection

### 5.3 Policy Portfolio Management
**Goal**: Central hub for all user policies

**Features**:
- Policy dashboard (all policies in one view)
- Policy details and documents
- Renewal reminders
- Payment history
- Beneficiary management
- Policy cancellation

**Database Tables**:
```sql
policy_beneficiaries (id, policy_id, beneficiary_name, relationship, percentage_share, id_number)
policy_renewals (policy_id, renewal_date, renewal_premium, reminder_sent_at, status)
policy_cancellations (policy_id, cancellation_date, cancellation_reason, refund_amount)
```

**Backend Files**:
- `internal/services/policy_portfolio_service.go`
- `internal/services/renewal_service.go`
- `internal/jobs/renewal_reminder_job.go` - Cron job for reminders

**Frontend Files**:
- `features/policies/dashboard/` - Policy overview
- `features/policies/policy-detail/` - Individual policy view
- `features/policies/beneficiaries/` - Manage beneficiaries
- `features/policies/renewals/` - Renewal management

---

## Phase 6: Claims Management

### 6.1 Claims Submission System
**Goal**: Easy and transparent claims process

**Features**:
- Claims filing workflow
- Document upload (photos, receipts, reports)
- Claims tracking
- Communication with provider
- Claims history

**Database Tables**:
```sql
claims (id, policy_id, claim_type, claim_amount, incident_date, incident_description, claim_date, status)
claim_documents (claim_id, document_type, file_url, uploaded_at)
claim_status_history (claim_id, status, updated_at, updated_by, notes)
claim_communications (id, claim_id, sender_id, sender_type, message, timestamp)
```

**Backend Files**:
- `internal/models/claim.go`
- `internal/services/claim_service.go`
- `internal/services/claim_notification_service.go`
- `internal/handlers/claim_handler.go`

**Frontend Files**:
- `features/claims/file-claim/` - Claims submission form
- `features/claims/claim-list/` - All claims view
- `features/claims/claim-detail/` - Track claim status
- `features/claims/claim-chat/` - Messaging with provider

### 6.2 AI-Powered Claims Assistant
**Goal**: Intelligent claims guidance

**Features**:
- Claims eligibility check
- Document checklist generation
- Fraud detection alerts (backend)
- Estimated settlement amount
- Next steps recommendations

**Backend Files**:
- `internal/services/claims_ai_service.go`
- `pkg/ai/claims_analyzer.go`
- `pkg/fraud/fraud_detection.go` - Basic fraud scoring

**Frontend Files**:
- `features/claims/claims-wizard/` - Guided claims process
- `features/claims/eligibility-check/` - Pre-submission check

---

## Phase 7: Consent & Data Sharing

### 7.1 Granular Consent Management
**Goal**: User control over data sharing

**Features**:
- Consent dashboard (see all consents)
- Grant/revoke consent per provider
- Consent expiry management
- Data sharing audit log
- Consent templates (regulatory compliance)

**Database Tables**:
```sql
consent_types (id, name, description, data_categories_json, regulatory_basis)
user_consents (id, user_id, provider_id, consent_type_id, granted_at, expires_at, revoked_at, consent_proof_signature)
data_sharing_logs (id, user_id, provider_id, data_shared_json, shared_at, consent_id, purpose)
```

**Backend Files**:
- `internal/models/consent.go`
- `internal/services/consent_service.go`
- `internal/services/data_sharing_service.go`
- `internal/middleware/consent_check.go` - Verify consent before data access

**Frontend Files**:
- `features/privacy/consent-dashboard/` - Manage all consents
- `features/privacy/consent-request/` - Provider consent request
- `features/privacy/data-sharing-log/` - Audit trail view
- `shared/components/consent-modal/` - Consent granting UI

### 7.2 Data Portability
**Goal**: User data ownership

**Features**:
- Export all user data (JSON/PDF)
- Import data from other platforms
- Delete account (right to be forgotten)
- Data anonymization

**Backend Files**:
- `internal/services/data_export_service.go`
- `internal/services/data_deletion_service.go`
- `pkg/anonymization/anonymizer.go`

**Frontend Files**:
- `features/privacy/data-export/`
- `features/privacy/delete-account/`

---

## Phase 8: Provider Features

### 8.1 Provider Dashboard
**Goal**: Tools for insurance providers

**Features**:
- Customer overview (with consent)
- Quote request management
- Policy issuance workflow
- Claims processing dashboard
- Commission tracking
- Analytics (policies sold, revenue, retention)

**Database Tables**:
```sql
provider_customers (provider_id, user_id, relationship_started, total_policies, total_premium)
provider_commissions (id, provider_id, policy_id, commission_amount, commission_rate, payment_date, status)
provider_analytics (provider_id, metric_name, metric_value, period_start, period_end)
```

**Backend Files**:
- `internal/services/provider_dashboard_service.go`
- `internal/services/provider_analytics_service.go`
- `internal/jobs/analytics_aggregation_job.go`

**Frontend Files**:
- `features/provider/dashboard/` - Provider home
- `features/provider/customers/` - Customer management
- `features/provider/quote-requests/` - Handle incoming quotes
- `features/provider/analytics/` - Business intelligence

### 8.2 Provider Communication Tools
**Goal**: Enable provider-customer interaction

**Features**:
- In-app messaging with customers
- Broadcast notifications
- Email campaigns
- SMS notifications (Kenya mobile numbers)

**Database Tables**:
```sql
provider_messages (id, provider_id, user_id, subject, message_body, sent_at, read_at)
provider_campaigns (id, provider_id, campaign_name, message_template, target_criteria_json, scheduled_at, status)
```

**Backend Files**:
- `internal/services/messaging_service.go`
- `internal/clients/sms_client.go` - Africa's Talking API integration
- `internal/services/campaign_service.go`

**Frontend Files**:
- `features/provider/messaging/`
- `features/provider/campaigns/`

---

## Phase 9: Mobile & Accessibility

### 9.1 Progressive Web App (PWA)
**Goal**: Mobile-first experience

**Features**:
- Offline mode (IndexedDB caching)
- Push notifications
- Add to home screen
- Responsive design (320px to 4K)
- Touch gestures

**Frontend Files**:
- `ngsw-config.json` - Service worker configuration
- `src/manifest.webmanifest` - PWA manifest
- `core/services/offline.service.ts` - Offline data sync
- `shared/components/install-prompt/` - PWA install prompt

### 9.2 Emerging Market Optimization
**Goal**: Performance in low-bandwidth environments

**Features**:
- Image lazy loading and compression
- Data-saver mode
- Low-bandwidth indicator
- Prefetching critical resources
- USSD integration (future consideration)

**Backend Files**:
- `internal/middleware/compression.go` - Gzip compression
- `internal/services/image_optimization_service.go`

**Frontend Files**:
- `core/services/network-status.service.ts` - Bandwidth detection
- `shared/components/data-saver-toggle/`

### 9.3 Multi-Language & Localization
**Goal**: Support local languages

**Features**:
- English, Swahili support
- Currency formatting (KES, USD)
- Date/time localization
- Right-to-left (RTL) support preparation

**Frontend Files**:
- `assets/i18n/en.json` - English translations
- `assets/i18n/sw.json` - Swahili translations
- `core/services/translation.service.ts`

---

## Phase 10: Security & Compliance

### 10.1 Security Hardening
**Goal**: Enterprise-grade security

**Features**:
- Rate limiting (per IP, per user)
- CSRF protection
- XSS prevention
- SQL injection prevention (parameterized queries)
- Security headers (CSP, HSTS)
- Penetration testing preparation

**Backend Files**:
- `internal/middleware/rate_limiter.go`
- `internal/middleware/security_headers.go`
- `internal/middleware/csrf.go`
- `pkg/validation/sanitizer.go`

### 10.2 Compliance & Audit
**Goal**: Regulatory compliance

**Features**:
- GDPR compliance tools
- Kenya Data Protection Act compliance
- Audit trail for all data access
- Data retention policies
- Incident response workflow

**Database Tables**:
```sql
audit_trail (id, user_id, action, entity_type, entity_id, ip_address, user_agent, timestamp, changes_json)
compliance_reports (id, report_type, generated_at, report_data_json, generated_by)
data_retention_policies (entity_type, retention_days, deletion_method)
```

**Backend Files**:
- `internal/services/audit_service.go`
- `internal/services/compliance_service.go`
- `internal/jobs/data_retention_job.go`

**Frontend Files**:
- `features/admin/audit-logs/`
- `features/admin/compliance-dashboard/`

---

## Phase 11: Analytics & Reporting

### 11.1 User Analytics
**Goal**: Insights for users

**Features**:
- Insurance portfolio summary
- Premium payment history
- Coverage visualization
- Financial health score
- Spending insights

**Frontend Files**:
- `features/analytics/user-dashboard/`
- `features/analytics/premium-tracker/`
- `shared/components/charts/` - Reusable chart components (Chart.js)

### 11.2 Platform Analytics
**Goal**: Business intelligence for platform owners

**Features**:
- User acquisition metrics
- Conversion funnels
- Provider performance
- Revenue tracking
- Churn analysis

**Backend Files**:
- `internal/services/platform_analytics_service.go`
- `internal/jobs/metrics_aggregation_job.go`

**Frontend Files**:
- `features/admin/platform-analytics/`

---

## Phase 12: Advanced Features

### 12.1 Smart Notifications
**Goal**: Timely and relevant alerts

**Features**:
- Policy renewal reminders (30, 14, 7 days before)
- Payment due notifications
- Claims status updates
- Personalized recommendations alerts
- Life event-triggered notifications

**Database Tables**:
```sql
notification_preferences (user_id, notification_type, channel, is_enabled)
notifications (id, user_id, notification_type, title, message, action_url, is_read, sent_at)
scheduled_notifications (id, user_id, notification_type, scheduled_for, payload_json, status)
```

**Backend Files**:
- `internal/services/notification_service.go`
- `internal/jobs/notification_scheduler_job.go`
- `internal/clients/push_notification_client.go` - Firebase Cloud Messaging

**Frontend Files**:
- `shared/components/notification-center/`
- `features/settings/notification-preferences/`

### 12.2 Referral & Rewards Program
**Goal**: User growth through referrals

**Features**:
- Referral code generation
- Referral tracking
- Rewards (discounts, cashback)
- Leaderboards

**Database Tables**:
```sql
referral_codes (id, user_id, code, created_at, uses_count, max_uses)
referrals (id, referrer_id, referred_user_id, referral_code, signup_date, policy_purchased, reward_earned)
user_rewards (id, user_id, reward_type, reward_amount, earned_date, redeemed_date, expiry_date)
```

**Backend Files**:
- `internal/services/referral_service.go`
- `internal/services/rewards_service.go`

**Frontend Files**:
- `features/referrals/referral-dashboard/`
- `features/referrals/rewards/`

### 12.3 Integration Ecosystem
**Goal**: Third-party integrations

**Features**:
- Hospital network API integration
- Garage/repair shop integration (auto insurance)
- Wearable device integration (health insurance)
- Bank account linking (payment automation)
- Government ID verification (IPRS in Kenya)

**Backend Files**:
- `internal/clients/hospital_api_client.go`
- `internal/clients/bank_integration.go`
- `internal/clients/iprs_client.go` - Kenya IPRS integration

---

## File Organization Standards

### Backend (Go)
```
internal/
├── models/          # Domain models (structs)
├── repository/      # Database layer (SQL queries)
├── services/        # Business logic
├── handlers/        # HTTP request handlers
├── middleware/      # HTTP middleware
├── clients/         # External API clients
└── jobs/            # Background jobs/cron tasks

pkg/
├── utils/           # Generic utilities
├── ai/              # AI-related packages
├── validation/      # Input validation
└── documents/       # Document generation
```

### Frontend (Angular)
```
features/
├── [feature-name]/
│   ├── [component-name]/
│   │   ├── component.ts        # TypeScript logic
│   │   ├── component.html      # HTML template
│   │   ├── component.css       # Component styles
│   │   └── component.spec.ts   # Unit tests

shared/
├── components/      # Reusable UI components
├── models/          # TypeScript interfaces
├── pipes/           # Custom pipes
└── directives/      # Custom directives

core/
├── services/        # Singleton services
├── guards/          # Route guards
└── interceptors/    # HTTP interceptors
```

---

## UI/UX Design Principles

### Visual Design
- **Color Palette**: 
  - Primary: Deep blue (#1E40AF) - Trust & stability
  - Secondary: Emerald green (#059669) - Growth & prosperity
  - Accent: Amber (#F59E0B) - Attention & warmth
  - Neutral: Gray scale for backgrounds
  
- **Typography**:
  - Headings: Inter (bold, 24-48px)
  - Body: Inter (regular, 14-16px)
  - Monospace: JetBrains Mono (for numbers/codes)

- **Layout**:
  - Card-based design with subtle shadows
  - 8px grid system
  - Maximum content width: 1440px
  - Generous white space (minimum 16px padding)

### UX Patterns
- **Progressive Disclosure**: Show basic info first, details on demand
- **Inline Validation**: Real-time form feedback
- **Skeleton Screens**: Show loading structure, not spinners
- **Empty States**: Helpful messages with clear CTAs
- **Error Handling**: Friendly error messages with recovery actions
- **Mobile Gestures**: Swipe for actions, pull to refresh

### Accessibility
- WCAG 2.1 AA compliance
- Keyboard navigation support
- Screen reader compatibility (ARIA labels)
- Color contrast ratio > 4.5:1
- Focus indicators on all interactive elements
- Skip navigation links

---

## Performance Targets

### Backend
- API response time: < 200ms (95th percentile)
- Database queries: < 50ms
- Background jobs: Process within 5 minutes
- Support 10,000 concurrent users

### Frontend
- First Contentful Paint: < 1.5s
- Time to Interactive: < 3.5s
- Lighthouse score: > 90