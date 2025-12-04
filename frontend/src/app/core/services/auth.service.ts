// Authentication Service - Handles user authentication and authorization
// Manages JWT tokens, login/logout, and user session
// Dependencies: ApiService, StorageService
// Usage: inject(AuthService).login(credentials)

import { Injectable, inject, signal, computed } from '@angular/core';
import { Router } from '@angular/router';
import { Observable, tap, catchError, of } from 'rxjs';
import { ApiService } from './api.service';
import { StorageService } from './storage.service';
import {
  LoginRequest,
  LoginResponse,
  RegisterRequest,
  RegisterResponse,
  RefreshTokenRequest,
  RefreshTokenResponse,
  VerifyOTPRequest,
  VerifyOTPResponse,
  PasswordResetRequest,
  PasswordResetConfirm,
  ResendOTPRequest
} from '../../shared/models/auth.model';
import { User } from '../../shared/models/user.model';

const TOKEN_KEY = 'access_token';
const REFRESH_TOKEN_KEY = 'refresh_token';
const USER_KEY = 'current_user';

@Injectable({
  providedIn: 'root'
})
export class AuthService {
  private readonly api = inject(ApiService);
  private readonly storage = inject(StorageService);
  private readonly router = inject(Router);

  // Signals for reactive state management
  private currentUserSignal = signal<User | null>(null);
  private isAuthenticatedSignal = signal<boolean>(false);

  // Public computed signals
  currentUser = this.currentUserSignal.asReadonly();
  isAuthenticated = this.isAuthenticatedSignal.asReadonly();
  userRole = computed(() => this.currentUserSignal()?.role || null);

  constructor() {
    // Initialize from storage on app start
    this.initializeFromStorage();
  }

  /**
   * Initialize authentication state from stored tokens
   * Called on app startup
   */
  private initializeFromStorage(): void {
    const token = this.storage.getItem<string>(TOKEN_KEY);
    const user = this.storage.getItem<User>(USER_KEY);

    if (token && user) {
      this.currentUserSignal.set(user);
      this.isAuthenticatedSignal.set(true);
    }
  }

  /**
   * Login with email and password
   * @param credentials - Login credentials
   * @returns Observable of login response
   */
  login(credentials: LoginRequest): Observable<LoginResponse> {
    return this.api.post<LoginResponse>('/auth/login', credentials)
      .pipe(
        tap(response => {
          this.persistSession(response);
        })
      );
  }

  /**
   * Register new user account
   * @param data - Registration data
   * @returns Observable of registration response
   */
  register(data: RegisterRequest): Observable<RegisterResponse> {
    return this.api.post<RegisterResponse>('/auth/register', data);
  }

  /**
   * Logout current user
   * Clears tokens and redirects to login
   */
  logout(): void {
    const refreshToken = this.getRefreshToken();
    if (refreshToken) {
      this.api.post('/auth/logout', { refreshToken }).subscribe({
        next: () => undefined,
        error: () => undefined
      });
    }

    // Clear storage
    this.storage.removeItem(TOKEN_KEY);
    this.storage.removeItem(REFRESH_TOKEN_KEY);
    this.storage.removeItem(USER_KEY);

    // Update signals
    this.currentUserSignal.set(null);
    this.isAuthenticatedSignal.set(false);

    // Redirect to login
    this.router.navigate(['/auth/login']);
  }

  /**
   * Get access token
   * @returns Access token or null
   */
  getAccessToken(): string | null {
    return this.storage.getItem<string>(TOKEN_KEY);
  }

  /**
   * Get refresh token
   * @returns Refresh token or null
   */
  getRefreshToken(): string | null {
    return this.storage.getItem<string>(REFRESH_TOKEN_KEY);
  }

  /**
   * Refresh access token using refresh token
   * @returns Observable of new tokens
   */
  refreshAccessToken(): Observable<RefreshTokenResponse> {
    const refreshToken = this.getRefreshToken();
    if (!refreshToken) {
      return of({} as RefreshTokenResponse);
    }

    return this.api.post<RefreshTokenResponse>('/auth/refresh', { refreshToken } satisfies RefreshTokenRequest)
      .pipe(
        tap(response => {
          this.persistSession(response);
        }),
        catchError(error => {
          // If refresh fails, logout user
          this.logout();
          return of({} as RefreshTokenResponse);
        })
      );
  }

  /**
   * Verify OTP code
   * @param data - OTP verification data
   * @returns Observable of verification response
   */
  verifyOTP(data: VerifyOTPRequest): Observable<VerifyOTPResponse> {
    return this.api.post<VerifyOTPResponse>('/auth/verify-otp', data);
  }

  /**
   * Resend OTP code
   */
  resendOTP(data: ResendOTPRequest): Observable<any> {
    return this.api.post('/auth/resend-otp', data);
  }

  /**
   * Request password reset
   * @param data - Password reset request data
   * @returns Observable of response
   */
  requestPasswordReset(data: PasswordResetRequest): Observable<any> {
    return this.api.post('/auth/password-reset/request', data);
  }

  /**
   * Confirm password reset with token
   */
  confirmPasswordReset(data: PasswordResetConfirm): Observable<any> {
    return this.api.post('/auth/password-reset/confirm', data);
  }

  /**
   * Check if user has specific role
   * @param role - Role to check
   * @returns True if user has role
   */
  hasRole(role: string): boolean {
    return this.currentUserSignal()?.role === role;
  }

  /**
   * Check if user has any of the specified roles
   * @param roles - Array of roles to check
   * @returns True if user has any of the roles
   */
  hasAnyRole(roles: string[]): boolean {
    const userRole = this.currentUserSignal()?.role;
    return userRole ? roles.includes(userRole) : false;
  }

  /**
   * Check if user is authenticated
   */
  isLoggedIn(): boolean {
    return this.isAuthenticatedSignal();
  }

  /**
   * Persist tokens and user info locally
   */
  private persistSession(payload: LoginResponse | RefreshTokenResponse): void {
    if (!payload.accessToken || !payload.refreshToken) {
      return;
    }

    this.storage.setItem(TOKEN_KEY, payload.accessToken);
    this.storage.setItem(REFRESH_TOKEN_KEY, payload.refreshToken);

    if (payload.user) {
      this.storage.setItem(USER_KEY, payload.user);
      this.currentUserSignal.set(payload.user);
    }

    this.isAuthenticatedSignal.set(true);
  }
}
