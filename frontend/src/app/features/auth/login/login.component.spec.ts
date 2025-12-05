import { TestBed } from '@angular/core/testing';
import { ActivatedRoute, Router, convertToParamMap } from '@angular/router';
import { of } from 'rxjs';
import { LoginComponent } from './login.component';
import { AuthService } from '../../../core/services/auth.service';

class AuthServiceMock {
  login = jasmine.createSpy('login').and.returnValue(of({}));
}

class RouterMock {
  navigate = jasmine.createSpy('navigate');
  navigateByUrl = jasmine.createSpy('navigateByUrl');
}

class ActivatedRouteMock {
  snapshot = {
    queryParamMap: convertToParamMap({})
  };
}

describe('LoginComponent', () => {
  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [LoginComponent],
      providers: [
        { provide: AuthService, useClass: AuthServiceMock },
        { provide: Router, useClass: RouterMock },
        { provide: ActivatedRoute, useClass: ActivatedRouteMock }
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
    expect(router.navigateByUrl).toHaveBeenCalledWith('/dashboard');
  });

  it('redirects to returnUrl query param when provided', () => {
    const fixture = TestBed.createComponent(LoginComponent);
    const component = fixture.componentInstance;
    const router = TestBed.inject(Router) as unknown as RouterMock;
    const route = TestBed.inject(ActivatedRoute) as unknown as ActivatedRouteMock;
    route.snapshot.queryParamMap = convertToParamMap({ returnUrl: '/quotes' });

    component.loginForm.setValue({ email: 'user@example.com', password: 'Password1' });
    component.submit();

    expect(router.navigateByUrl).toHaveBeenCalledWith('/quotes');
  });
});
