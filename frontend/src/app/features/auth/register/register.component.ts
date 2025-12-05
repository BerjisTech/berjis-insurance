import { CommonModule } from '@angular/common';
import { Component, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';
import { AuthService } from '../../../core/services/auth.service';
import { RegisterRequest } from '../../../shared/models/auth.model';
import {
  PASSWORD_COMPLEXITY_PATTERN,
  PASSWORD_REQUIREMENTS,
  PHONE_E164_PATTERN
} from '../../../shared/utils/form-validators';

@Component({
  selector: 'app-register',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule, RouterLink],
  templateUrl: './register.component.html',
  styleUrls: ['./register.component.css']
})
export class RegisterComponent {
  private readonly fb = inject(FormBuilder);
  private readonly authService = inject(AuthService);
  private readonly router = inject(Router);

  readonly passwordRequirements = PASSWORD_REQUIREMENTS;

  registerForm = this.fb.nonNullable.group({
    fullName: ['', [Validators.required, Validators.minLength(3)]],
    email: ['', [Validators.required, Validators.email]],
    phone: ['', [Validators.pattern(PHONE_E164_PATTERN)]],
    password: ['', [Validators.required, Validators.pattern(PASSWORD_COMPLEXITY_PATTERN)]],
    confirmPassword: ['', [Validators.required]]
  });

  loading = signal(false);
  errorMessage = signal('');
  infoMessage = signal('');
  debugOtp = signal('');

  submit(): void {
    if (this.registerForm.invalid || this.loading()) {
      this.registerForm.markAllAsTouched();
      return;
    }

    if (this.registerForm.value.password !== this.registerForm.value.confirmPassword) {
      this.errorMessage.set('Passwords do not match.');
      return;
    }

    const phone = this.registerForm.value.phone?.trim();
    const fullName = this.registerForm.value.fullName?.trim();

    const payload: RegisterRequest = {
      email: this.registerForm.value.email!,
      password: this.registerForm.value.password!,
      phone: phone || undefined,
      fullName: fullName || undefined
    };

    this.loading.set(true);
    this.errorMessage.set('');
    this.infoMessage.set('');
    this.debugOtp.set('');

    this.authService.register(payload).subscribe({
      next: response => {
        this.loading.set(false);
        this.infoMessage.set('Account created! Please verify using the OTP we sent.');
        if (response.debugOtp) {
          this.debugOtp.set(`Development OTP: ${response.debugOtp}`);
        }

        const purpose = payload.phone ? 'phone_verification' : 'email_verification';
        this.router.navigate(['/auth/verify-otp'], {
          queryParams: {
            userId: response.userId,
            purpose
          }
        });
      },
      error: (err: Error) => {
        this.errorMessage.set(err.message ?? 'Unable to create account.');
        this.loading.set(false);
      }
    });
  }

  navigateToLogin(): void {
    this.router.navigate(['/auth/login']);
  }
}
