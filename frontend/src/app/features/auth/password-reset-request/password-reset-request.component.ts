import { CommonModule } from '@angular/common';
import { Component, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { RouterLink } from '@angular/router';
import { AuthService } from '../../../core/services/auth.service';

@Component({
  selector: 'app-password-reset-request',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule, RouterLink],
  templateUrl: './password-reset-request.component.html',
  styleUrls: ['./password-reset-request.component.css']
})
export class PasswordResetRequestComponent {
  private readonly fb = inject(FormBuilder);
  private readonly authService = inject(AuthService);

  requestForm = this.fb.nonNullable.group({
    email: ['', [Validators.required, Validators.email]]
  });

  loading = signal(false);
  message = signal('');
  errorMessage = signal('');

  submit(): void {
    if (this.requestForm.invalid || this.loading()) {
      this.requestForm.markAllAsTouched();
      return;
    }

    this.loading.set(true);
    this.message.set('');
    this.errorMessage.set('');

    this.authService.requestPasswordReset(this.requestForm.getRawValue()).subscribe({
      next: response => {
        this.loading.set(false);
        const baseMessage = 'Check your email for password reset instructions.';
        this.message.set(
          response?.debugToken ? `${baseMessage} (Dev token: ${response.debugToken})` : baseMessage
        );
      },
      error: (err: Error) => {
        this.loading.set(false);
        this.errorMessage.set(err.message ?? 'Unable to process request.');
      }
    });
  }
}
