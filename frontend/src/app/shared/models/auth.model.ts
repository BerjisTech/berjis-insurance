// Authentication domain models and interfaces
// Defines TypeScript interfaces for authentication-related data structures

import { User } from './user.model';

export interface LoginRequest {
  email: string;
  password: string;
  deviceId?: string;
  deviceName?: string;
}

export interface LoginResponse {
  accessToken: string;
  refreshToken: string;
  user: User;
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
  debugOtp?: string;
}

export interface RefreshTokenRequest {
  refreshToken: string;
}

export interface RefreshTokenResponse {
  accessToken: string;
  refreshToken: string;
  user: User;
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

export interface ResendOTPRequest {
  userId: string;
  purpose?: OTPPurpose;
}

export interface ResendOTPResponse {
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
