-- Insurance Broker Platform - Initial Database Schema
-- Version: 1.0.0
-- Description: Core tables for authentication, users, providers, and audit
-- Dependencies: PostgreSQL 16+, pgcrypto extension

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";      -- UUID generation
CREATE EXTENSION IF NOT EXISTS "pgcrypto";        -- Encryption functions
CREATE EXTENSION IF NOT EXISTS "pg_trgm";         -- Text search trigrams

-- =============================================================================
-- CORE AUTHENTICATION & USERS
-- =============================================================================

-- Users table: Core authentication and user management
-- Supports multiple authentication methods (password, OTP, OAuth)
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    phone VARCHAR(20) UNIQUE,
    password_hash VARCHAR(255),  -- bcrypt hash (nullable for OAuth users)
    role VARCHAR(50) NOT NULL DEFAULT 'user',  -- user, broker, provider_admin, system_admin
    email_verified BOOLEAN DEFAULT FALSE,
    phone_verified BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    last_login_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT valid_email CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$'),
    CONSTRAINT valid_role CHECK (role IN ('user', 'broker', 'provider_admin', 'system_admin'))
);

-- Indexes for performance
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_phone ON users(phone);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_created_at ON users(created_at DESC);

-- User profiles: Extended user information
CREATE TABLE user_profiles (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    full_name VARCHAR(255),
    date_of_birth DATE,
    gender VARCHAR(20),
    occupation VARCHAR(255),
    id_number VARCHAR(50),  -- National ID or passport
    profile_photo_url TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT valid_gender CHECK (gender IN ('male', 'female', 'other', 'prefer_not_to_say'))
);

-- User addresses: Support multiple addresses
CREATE TABLE user_addresses (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL,  -- home, work, billing, shipping
    street TEXT,
    city VARCHAR(100),
    state_province VARCHAR(100),
    postal_code VARCHAR(20),
    country VARCHAR(2) DEFAULT 'KE',  -- ISO 3166-1 alpha-2
    is_primary BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT valid_address_type CHECK (type IN ('home', 'work', 'billing', 'shipping'))
);

CREATE INDEX idx_user_addresses_user_id ON user_addresses(user_id);
CREATE INDEX idx_user_addresses_primary ON user_addresses(user_id, is_primary) WHERE is_primary = TRUE;

-- User documents: KYC and verification documents
CREATE TABLE user_documents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    document_type VARCHAR(50) NOT NULL,  -- national_id, passport, proof_of_address, etc.
    file_url TEXT NOT NULL,
    file_size INTEGER,  -- bytes
    mime_type VARCHAR(100),
    verification_status VARCHAR(50) DEFAULT 'pending',  -- pending, verified, rejected
    verified_at TIMESTAMP WITH TIME ZONE,
    verified_by UUID REFERENCES users(id),
    rejection_reason TEXT,
    uploaded_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT valid_verification_status CHECK (verification_status IN ('pending', 'verified', 'rejected'))
);

CREATE INDEX idx_user_documents_user_id ON user_documents(user_id);
CREATE INDEX idx_user_documents_status ON user_documents(verification_status);

-- User preferences: Language, currency, notifications
CREATE TABLE user_preferences (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    language VARCHAR(10) DEFAULT 'en',  -- ISO 639-1 language code
    currency VARCHAR(3) DEFAULT 'KES',   -- ISO 4217 currency code
    timezone VARCHAR(50) DEFAULT 'Africa/Nairobi',
    notification_email BOOLEAN DEFAULT TRUE,
    notification_sms BOOLEAN DEFAULT TRUE,
    notification_push BOOLEAN DEFAULT TRUE,
    data_saver_mode BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- =============================================================================
-- AUTHENTICATION & SECURITY
-- =============================================================================

-- Refresh tokens: JWT refresh token management
CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL UNIQUE,  -- SHA-256 hash of token
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    is_revoked BOOLEAN DEFAULT FALSE,
    revoked_at TIMESTAMP WITH TIME ZONE,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);
CREATE INDEX idx_refresh_tokens_token_hash ON refresh_tokens(token_hash);

-- OTP codes: One-time passwords for 2FA and verification
CREATE TABLE otp_codes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code VARCHAR(10) NOT NULL,  -- 6-digit numeric code
    purpose VARCHAR(50) NOT NULL,  -- email_verification, phone_verification, login, password_reset
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    is_used BOOLEAN DEFAULT FALSE,
    used_at TIMESTAMP WITH TIME ZONE,
    attempts INTEGER DEFAULT 0,  -- Track failed attempts
    max_attempts INTEGER DEFAULT 3,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT valid_otp_purpose CHECK (purpose IN ('email_verification', 'phone_verification', 'login', 'password_reset'))
);

CREATE INDEX idx_otp_codes_user_id ON otp_codes(user_id);
CREATE INDEX idx_otp_codes_expires_at ON otp_codes(expires_at);

-- Password reset tokens: Secure password reset flow
CREATE TABLE password_reset_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL UNIQUE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    is_used BOOLEAN DEFAULT FALSE,
    used_at TIMESTAMP WITH TIME ZONE,
    ip_address INET,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_password_reset_tokens_user_id ON password_reset_tokens(user_id);
CREATE INDEX idx_password_reset_tokens_expires_at ON password_reset_tokens(expires_at);

-- =============================================================================
-- INSURANCE PROVIDERS
-- =============================================================================

-- Insurance providers: Companies offering insurance products
CREATE TABLE insurance_providers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL UNIQUE,
    license_number VARCHAR(100) UNIQUE NOT NULL,
    country_code VARCHAR(2) DEFAULT 'KE',  -- ISO 3166-1 alpha-2
    status VARCHAR(50) DEFAULT 'pending',  -- pending, active, suspended, inactive
    logo_url TEXT,
    description TEXT,
    website VARCHAR(255),
    support_phone VARCHAR(20),
    support_email VARCHAR(255),
    api_enabled BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT valid_provider_status CHECK (status IN ('pending', 'active', 'suspended', 'inactive'))
);

CREATE INDEX idx_insurance_providers_status ON insurance_providers(status);
CREATE INDEX idx_insurance_providers_country ON insurance_providers(country_code);

-- Provider API configurations: API integration settings
CREATE TABLE provider_api_configs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    provider_id UUID NOT NULL REFERENCES insurance_providers(id) ON DELETE CASCADE,
    api_type VARCHAR(50) NOT NULL,  -- rest, soap, graphql
    endpoint_url TEXT NOT NULL,
    auth_method VARCHAR(50) NOT NULL,  -- api_key, oauth2, basic_auth
    credentials_encrypted TEXT,  -- Encrypted credentials (pgcrypto)
    is_active BOOLEAN DEFAULT TRUE,
    last_successful_call TIMESTAMP WITH TIME ZONE,
    last_error TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT valid_api_type CHECK (api_type IN ('rest', 'soap', 'graphql')),
    CONSTRAINT valid_auth_method CHECK (auth_method IN ('api_key', 'oauth2', 'basic_auth', 'jwt'))
);

CREATE INDEX idx_provider_api_configs_provider_id ON provider_api_configs(provider_id);

-- =============================================================================
-- CONSENT & DATA SHARING
-- =============================================================================

-- User provider consents: GDPR/KDPA compliant consent management
CREATE TABLE user_provider_consents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider_id UUID NOT NULL REFERENCES insurance_providers(id) ON DELETE CASCADE,
    consent_type VARCHAR(100) NOT NULL,  -- data_sharing, marketing, communications
    granted_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE,
    revoked_at TIMESTAMP WITH TIME ZONE,
    consent_proof_signature TEXT,  -- Digital signature or consent form reference
    ip_address INET,
    user_agent TEXT,
    UNIQUE(user_id, provider_id, consent_type)
);

CREATE INDEX idx_user_consents_user_id ON user_provider_consents(user_id);
CREATE INDEX idx_user_consents_provider_id ON user_provider_consents(provider_id);
CREATE INDEX idx_user_consents_expires_at ON user_provider_consents(expires_at);

-- =============================================================================
-- AUDIT & COMPLIANCE
-- =============================================================================

-- Audit logs: Comprehensive audit trail for compliance
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    action VARCHAR(100) NOT NULL,  -- login, logout, create, update, delete, view
    entity_type VARCHAR(100),  -- users, policies, claims, etc.
    entity_id UUID,
    ip_address INET,
    user_agent TEXT,
    request_path TEXT,
    request_method VARCHAR(10),
    changes JSONB,  -- Before/after values for updates
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_entity ON audit_logs(entity_type, entity_id);
CREATE INDEX idx_audit_logs_timestamp ON audit_logs(timestamp DESC);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);

-- =============================================================================
-- FUNCTIONS & TRIGGERS
-- =============================================================================

-- Function to update updated_at timestamp automatically
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply updated_at trigger to relevant tables
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_user_profiles_updated_at BEFORE UPDATE ON user_profiles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_user_addresses_updated_at BEFORE UPDATE ON user_addresses
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_user_preferences_updated_at BEFORE UPDATE ON user_preferences
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_insurance_providers_updated_at BEFORE UPDATE ON insurance_providers
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_provider_api_configs_updated_at BEFORE UPDATE ON provider_api_configs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- =============================================================================
-- SEED DATA
-- =============================================================================

-- Create default system admin user (password: admin123 - CHANGE IN PRODUCTION!)
-- Password hash generated with bcrypt (cost 10)
INSERT INTO users (email, password_hash, role, email_verified, is_active)
VALUES (
    'admin@insurance.local',
    '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi',  -- admin123
    'system_admin',
    TRUE,
    TRUE
);

-- Insert default system admin profile
INSERT INTO user_profiles (user_id, full_name)
VALUES (
    (SELECT id FROM users WHERE email = 'admin@insurance.local'),
    'System Administrator'
);

-- Insert default user preferences for admin
INSERT INTO user_preferences (user_id)
VALUES (
    (SELECT id FROM users WHERE email = 'admin@insurance.local')
);

-- Migration complete
-- Schema version: 001
