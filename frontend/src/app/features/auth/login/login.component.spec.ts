import { TestBed } from '@angular/core/testing';
import { Router } from '@angular/router';
import { of } from 'rxjs';
import { LoginComponent } from './login.component';
import { AuthService } from '../../../core/services/auth.service';

class AuthServiceMock {
  login = jasmine.createSpy('login').and.returnValue(of({}));
}

class RouterMock {
  navigate = jasmine.createSpy('navigate');
}

describe('LoginComponent', () => {
  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [LoginComponent],
      providers: [
        { provide: AuthService, useClass: AuthServiceMock },
        { provide: Router, useClass: RouterMock }
      ]
    }).compileComponents();
  });

  it('marks form invalid by default', () => {
    const fixture = TestBed.createComponent(LoginComponent);
    const component = fixture.componentInstance;
    expect(component.loginForm.valid).toBeFalse();
  });

  it('submits credentials when valid', () => {
    const fixture = TestBed.createComponent(LoginComponent);
    const component = fixture.componentInstance;
    const authService = TestBed.inject(AuthService) as unknown as AuthServiceMock;
    const router = TestBed.inject(Router) as unknown as RouterMock;

    component.loginForm.setValue({ email: 'user@example.com', password: 'Password1' });
    component.submit();

    expect(authService.login).toHaveBeenCalled();
    expect(router.navigate).toHaveBeenCalledWith(['/dashboard']);
  });
});
