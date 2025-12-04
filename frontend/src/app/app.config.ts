// Application Configuration
// Configures application-wide providers and settings
// Dependencies: Router, HttpClient, Interceptors

import { ApplicationConfig, provideZoneChangeDetection } from '@angular/core';
import { provideRouter } from '@angular/router';
import { provideHttpClient, withInterceptors } from '@angular/common/http';

import { routes } from './app.routes';
import { authInterceptor } from './core/interceptors/auth.interceptor';

export const appConfig: ApplicationConfig = {
  providers: [
    // Zone change detection optimization
    provideZoneChangeDetection({ eventCoalescing: true }),

    // Router configuration
    provideRouter(routes),

    // HTTP Client with interceptors
    provideHttpClient(
      withInterceptors([authInterceptor])
    )
  ]
};
