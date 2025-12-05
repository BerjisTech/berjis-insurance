import { TestBed, fakeAsync, tick, discardPeriodicTasks } from '@angular/core/testing';
import { ActivatedRoute, convertToParamMap } from '@angular/router';
import { RouterTestingModule } from '@angular/router/testing';
import { of } from 'rxjs';
import { VerifyOtpComponent } from './verify-otp.component';
import { AuthService } from '../../../core/services/auth.service';

class AuthServiceMock {
  verifyOTP = jasmine.createSpy('verifyOTP').and.returnValue(of({}));
  resendOTP = jasmine.createSpy('resendOTP').and.returnValue(of({ message: 'sent' }));
}

class ActivatedRouteMock {
  snapshot = {
    queryParamMap: convertToParamMap({
      userId: 'user-123',
      purpose: 'email_verification'
    })
  };
}

describe('VerifyOtpComponent', () => {
  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [VerifyOtpComponent, RouterTestingModule],
      providers: [
        { provide: AuthService, useClass: AuthServiceMock },
        { provide: ActivatedRoute, useClass: ActivatedRouteMock }
      ]
    }).compileComponents();
  });

  it('throttles resend requests with cooldown timer', fakeAsync(() => {
    const fixture = TestBed.createComponent(VerifyOtpComponent);
    const component = fixture.componentInstance;
    const authService = TestBed.inject(AuthService) as unknown as AuthServiceMock;

    component.resend();
    expect(authService.resendOTP).toHaveBeenCalledTimes(1);
    expect(authService.resendOTP).toHaveBeenCalledWith({
      userId: 'user-123',
      purpose: 'email_verification'
    });
    expect(component.cooldown()).toBe(30);

    component.resend();
    expect(authService.resendOTP).toHaveBeenCalledTimes(1);

    tick(5000);
    expect(component.cooldown()).toBe(25);

    tick(25000);
    expect(component.cooldown()).toBe(0);

    component.resend();
    expect(authService.resendOTP).toHaveBeenCalledTimes(2);

    tick(30000);
    discardPeriodicTasks();
  }));
});
