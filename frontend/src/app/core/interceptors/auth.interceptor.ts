// Authentication Interceptor - JWT Token Injection
// Automatically adds JWT token to all HTTP requests
// Dependencies: HttpInterceptorFn, AuthService
// Usage: Configured in app.config.ts providers

import { HttpInterceptorFn, HttpErrorResponse } from '@angular/common/http';
import { inject } from '@angular/core';
import { catchError, throwError, switchMap } from 'rxjs';
import { AuthService } from '../services/auth.service';

/**
 * Auth interceptor function (functional interceptor for Angular 18+)
 * Adds Authorization header with JWT token to requests
 * Handles 401 errors by attempting token refresh
 */
export const authInterceptor: HttpInterceptorFn = (req, next) => {
  const authService = inject(AuthService);

  // Skip token injection for auth endpoints
  if (req.url.includes('/auth/login') || req.url.includes('/auth/register')) {
    return next(req);
  }

  // Get access token
  const token = authService.getAccessToken();

  // Clone request and add Authorization header if token exists
  if (token) {
    req = req.clone({
      setHeaders: {
        Authorization: `Bearer ${token}`
      }
    });
  }

  // Handle the request and catch 401 errors
  return next(req).pipe(
    catchError((error: HttpErrorResponse) => {
      // If 401 Unauthorized, attempt to refresh token
      if (error.status === 401 && !req.url.includes('/auth/refresh')) {
        return authService.refreshAccessToken().pipe(
          switchMap(response => {
            // Retry original request with new token
            const newReq = req.clone({
              setHeaders: {
                Authorization: `Bearer ${response.accessToken}`
              }
            });
            return next(newReq);
          }),
          catchError(refreshError => {
            // Refresh failed, logout user
            authService.logout();
            return throwError(() => refreshError);
          })
        );
      }

      return throwError(() => error);
    })
  );
};
