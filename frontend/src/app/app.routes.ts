import { Routes } from '@angular/router';

export const routes: Routes = [
  { path: '', redirectTo: 'auth/login', pathMatch: 'full' },
  {
    path: 'auth',
    children: [
      {
        path: 'login',
        loadComponent: () => import('./features/auth/login/login.component').then(m => m.LoginComponent)
      },
      {
        path: 'register',
        loadComponent: () => import('./features/auth/register/register.component').then(m => m.RegisterComponent)
      },
      {
        path: 'verify-otp',
        loadComponent: () => import('./features/auth/verify-otp/verify-otp.component').then(m => m.VerifyOtpComponent)
      },
      {
        path: 'password-reset',
        loadComponent: () => import('./features/auth/password-reset-request/password-reset-request.component').then(m => m.PasswordResetRequestComponent)
      },
      {
        path: 'password-reset/confirm',
        loadComponent: () => import('./features/auth/password-reset-confirm/password-reset-confirm.component').then(m => m.PasswordResetConfirmComponent)
      }
    ]
  },
  { path: '**', redirectTo: 'auth/login' }
];
