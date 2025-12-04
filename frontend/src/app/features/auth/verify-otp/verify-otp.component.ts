import { CommonModule } from '@angular/common';
import { Component, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import { AuthService } from '../../../core/services/auth.service';
import { VerifyOTPRequest } from '../../../shared/models/auth.model';

@Component({
  selector: 'app-verify-otp',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule, RouterLink],
  templateUrl: './verify-otp.component.html',
  styleUrls: ['./verify-otp.component.css']
})
export class VerifyOtpComponent {
  private readonly fb = inject(FormBuilder);
  private readonly authService = inject(AuthService);
  private readonly route = inject(ActivatedRoute);
  private readonly router = inject(Router);

  private pendingUserId = this.route.snapshot.queryParamMap.get('userId') ?? '';
  private purpose = (this.route.snapshot.queryParamMap.get('purpose') ?? 'phone_verification') as VerifyOTPRequest['purpose'];

  otpForm = this.fb.nonNullable.group({
    code: ['', [Validators.required, Validators.pattern(/^[0-9]{6}$/)]]
  });

  loading = signal(false);
  message = signal('');
  errorMessage = signal('');

  submit(): void {
    if (this.otpForm.invalid || this.loading()) {
      this.otpForm.markAllAsTouched();
      return;
    }

    if (!this.pendingUserId) {
      this.errorMessage.set('Missing user reference. Please register again.');
      return;
    }

    const payload: VerifyOTPRequest = {
      userId: this.pendingUserId,
      purpose: this.purpose,
      code: this.otpForm.value.code ?? ''
    };

    this.loading.set(true);
    this.errorMessage.set('');
    this.message.set('');

    this.authService.verifyOTP(payload).subscribe({
      next: () => {
        this.loading.set(false);
        this.message.set('Verified! You can now sign in.');
        setTimeout(() => this.router.navigate(['/auth/login']), 1200);
      },
      error: (err: Error) => {
        this.loading.set(false);
        this.errorMessage.set(err.message ?? 'Invalid code.');
      }
    });
  }

  resend(): void {
    // Future enhancement: call resend endpoint
    this.errorMessage.set('Resend not yet available. Please check your inbox.');
  }
}
