// Authentication domain models and interfaces
// Defines TypeScript interfaces for authentication-related data structures

export interface LoginRequest {
  email: string;
  password: string;
}

export interface LoginResponse {
  accessToken: string;
  refreshToken: string;
  user: {
    id: string;
    email: string;
    role: string;
  };
}

export interface RegisterRequest {
  email: string;
  password: string;
  phone?: string;
  fullName?: string;
}

export interface RegisterResponse {
  message: string;
  userId: string;
  requiresVerification: boolean;
}

export interface RefreshTokenRequest {
  refreshToken: string;
}

export interface RefreshTokenResponse {
  accessToken: string;
  refreshToken: string;
}

export interface VerifyOTPRequest {
  userId: string;
  code: string;
  purpose: OTPPurpose;
}

export interface VerifyOTPResponse {
  success: boolean;
  message: string;
}

export interface PasswordResetRequest {
  email: string;
}

export interface PasswordResetConfirm {
  token: string;
  newPassword: string;
}

export type OTPPurpose =
  | 'email_verification'
  | 'phone_verification'
  | 'login'
  | 'password_reset';
