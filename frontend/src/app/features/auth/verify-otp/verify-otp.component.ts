import { CommonModule } from '@angular/common';
import { Component, inject, signal, OnDestroy } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import { AuthService } from '../../../core/services/auth.service';
import { VerifyOTPRequest } from '../../../shared/models/auth.model';
import { OTP_CODE_PATTERN } from '../../../shared/utils/form-validators';

@Component({
  selector: 'app-verify-otp',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule, RouterLink],
  templateUrl: './verify-otp.component.html',
  styleUrls: ['./verify-otp.component.css']
})
export class VerifyOtpComponent implements OnDestroy {
  private readonly fb = inject(FormBuilder);
  private readonly authService = inject(AuthService);
  private readonly route = inject(ActivatedRoute);
  private readonly router = inject(Router);

  private readonly defaultPurpose: VerifyOTPRequest['purpose'] = 'email_verification';
  private pendingUserId = this.route.snapshot.queryParamMap.get('userId') ?? '';
  private purpose = (this.route.snapshot.queryParamMap.get('purpose') ?? this.defaultPurpose) as VerifyOTPRequest['purpose'];

  otpForm = this.fb.nonNullable.group({
    code: ['', [Validators.required, Validators.pattern(OTP_CODE_PATTERN)]]
  });

  loading = signal(false);
  message = signal('');
  errorMessage = signal('');
  cooldown = signal(0);
  private cooldownTimer: ReturnType<typeof setInterval> | null = null;

  ngOnDestroy(): void {
    this.clearCooldown();
  }

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
    if (!this.pendingUserId) {
      this.errorMessage.set('Missing user reference.');
      return;
    }

    if (this.cooldown() > 0 || this.loading()) {
      return;
    }

    this.loading.set(true);
    this.errorMessage.set('');
    this.message.set('');

    this.authService.resendOTP({
      userId: this.pendingUserId,
      purpose: this.purpose
    }).subscribe({
      next: (response) => {
        this.loading.set(false);
        this.message.set(response?.message ?? 'A new code has been sent.');
        this.startCooldown(30);
      },
      error: (err: Error) => {
        this.loading.set(false);
        this.errorMessage.set(err.message ?? 'Unable to resend code.');
      }
    });
  }

  private startCooldown(seconds: number): void {
    this.clearCooldown();
    this.cooldown.set(seconds);

    this.cooldownTimer = setInterval(() => {
      this.cooldown.update(value => {
        if (value <= 1) {
          this.clearCooldown();
          return 0;
        }
        return value - 1;
      });
    }, 1000);
  }

  private clearCooldown(): void {
    if (this.cooldownTimer) {
      clearInterval(this.cooldownTimer);
      this.cooldownTimer = null;
    }
  }
}
