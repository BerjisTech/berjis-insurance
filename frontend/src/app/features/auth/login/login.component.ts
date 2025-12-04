import { CommonModule } from '@angular/common';
import { Component, signal, inject } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';
import { AuthService } from '../../../core/services/auth.service';
import { LoginRequest } from '../../../shared/models/auth.model';

@Component({
  selector: 'app-login',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule, RouterLink],
  templateUrl: './login.component.html',
  styleUrls: ['./login.component.css']
})
export class LoginComponent {
  private readonly fb = inject(FormBuilder);
  private readonly authService = inject(AuthService);
  private readonly router = inject(Router);

  loginForm = this.fb.nonNullable.group({
    email: ['', [Validators.required, Validators.email]],
    password: ['', [Validators.required, Validators.minLength(8)]]
  });

  loading = signal(false);
  errorMessage = signal('');
  successMessage = signal('');

  submit(): void {
    if (this.loginForm.invalid || this.loading()) {
      this.loginForm.markAllAsTouched();
      return;
    }

    const payload: LoginRequest = {
      ...this.loginForm.getRawValue(),
      deviceName: 'Web Browser'
    };

    this.loading.set(true);
    this.errorMessage.set('');
    this.successMessage.set('');

    this.authService.login(payload).subscribe({
      next: () => {
        this.successMessage.set('Welcome back! You are now signed in.');
        this.loading.set(false);
      },
      error: (err: Error) => {
        this.errorMessage.set(err.message ?? 'Unable to login. Please try again.');
        this.loading.set(false);
      }
    });
  }

  navigateToRegister(): void {
    this.router.navigate(['/auth/register']);
  }

  navigateToForgotPassword(): void {
    this.router.navigate(['/auth/password-reset']);
  }
}
