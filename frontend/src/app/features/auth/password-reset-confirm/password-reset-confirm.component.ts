import { CommonModule } from '@angular/common';
import { Component, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { ActivatedRoute, RouterLink } from '@angular/router';
import { AuthService } from '../../../core/services/auth.service';

@Component({
  selector: 'app-password-reset-confirm',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule, RouterLink],
  templateUrl: './password-reset-confirm.component.html',
  styleUrls: ['./password-reset-confirm.component.css']
})
export class PasswordResetConfirmComponent {
  private readonly fb = inject(FormBuilder);
  private readonly authService = inject(AuthService);
  private readonly route = inject(ActivatedRoute);

  confirmForm = this.fb.nonNullable.group({
    token: ['', Validators.required],
    newPassword: ['', [Validators.required, Validators.minLength(8)]],
    confirmPassword: ['', Validators.required]
  });

  loading = signal(false);
  message = signal('');
  errorMessage = signal('');

  constructor() {
    const token = this.route.snapshot.queryParamMap.get('token');
    if (token) {
      this.confirmForm.patchValue({ token });
    }
  }

  submit(): void {
    if (this.confirmForm.invalid || this.loading()) {
      this.confirmForm.markAllAsTouched();
      return;
    }

    if (this.confirmForm.value.newPassword !== this.confirmForm.value.confirmPassword) {
      this.errorMessage.set('Passwords do not match.');
      return;
    }

    this.loading.set(true);
    this.message.set('');
    this.errorMessage.set('');

    this.authService.confirmPasswordReset({
      token: this.confirmForm.value.token!,
      newPassword: this.confirmForm.value.newPassword!
    }).subscribe({
      next: () => {
        this.loading.set(false);
        this.message.set('Password updated! You can now sign in.');
      },
      error: (err: Error) => {
        this.loading.set(false);
        this.errorMessage.set(err.message ?? 'Unable to reset password.');
      }
    });
  }
}
