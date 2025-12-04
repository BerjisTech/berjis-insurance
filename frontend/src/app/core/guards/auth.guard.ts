// Authentication Guard - Route Protection
// Prevents unauthorized access to protected routes
// Dependencies: AuthService, Router
// Usage: Add to route canActivate: [authGuard]

import { inject } from '@angular/core';
import { Router, CanActivateFn } from '@angular/router';
import { AuthService } from '../services/auth.service';

/**
 * Auth guard function (functional guard for Angular 18+)
 * Checks if user is authenticated before allowing route access
 * Redirects to login if not authenticated
 */
export const authGuard: CanActivateFn = (route, state) => {
  const authService = inject(AuthService);
  const router = inject(Router);

  // Check if user is authenticated
  if (authService.isLoggedIn()) {
    return true;
  }

  // Store intended URL for redirect after login
  const returnUrl = state.url;

  // Redirect to login with return URL
  router.navigate(['/auth/login'], {
    queryParams: { returnUrl }
  });

  return false;
};

/**
 * Role-based guard factory
 * Creates a guard that checks for specific roles
 * @param allowedRoles - Array of allowed roles
 * @returns CanActivateFn
 */
export const roleGuard = (allowedRoles: string[]): CanActivateFn => {
  return (route, state) => {
    const authService = inject(AuthService);
    const router = inject(Router);

    // Check if user is authenticated
    if (!authService.isLoggedIn()) {
      router.navigate(['/auth/login'], {
        queryParams: { returnUrl: state.url }
      });
      return false;
    }

    // Check if user has required role
    if (authService.hasAnyRole(allowedRoles)) {
      return true;
    }

    // User doesn't have required role - redirect to dashboard
    router.navigate(['/dashboard']);
    return false;
  };
};
