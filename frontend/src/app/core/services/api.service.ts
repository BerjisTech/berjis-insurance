// API Service - HTTP Client Wrapper
// Handles all HTTP requests to the backend API
// Dependencies: HttpClient, environment
// Usage: inject(ApiService).get<T>('/endpoint')

import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpHeaders, HttpParams, HttpErrorResponse } from '@angular/common/http';
import { Observable, throwError } from 'rxjs';
import { catchError, timeout } from 'rxjs/operators';
import { environment } from '../../../environments/environment';

@Injectable({
  providedIn: 'root'
})
export class ApiService {
  private readonly http = inject(HttpClient);
  private readonly baseUrl = environment.apiUrl;
  private readonly timeout = environment.apiTimeout;

  /**
   * GET request
   * @param endpoint - API endpoint (e.g., '/users/me')
   * @param params - Query parameters
   * @returns Observable of response
   */
  get<T>(endpoint: string, params?: HttpParams): Observable<T> {
    const url = this.buildUrl(endpoint);
    return this.http.get<T>(url, { params })
      .pipe(
        timeout(this.timeout),
        catchError(this.handleError)
      );
  }

  /**
   * POST request
   * @param endpoint - API endpoint
   * @param body - Request body
   * @returns Observable of response
   */
  post<T>(endpoint: string, body: any): Observable<T> {
    const url = this.buildUrl(endpoint);
    return this.http.post<T>(url, body)
      .pipe(
        timeout(this.timeout),
        catchError(this.handleError)
      );
  }

  /**
   * PUT request
   * @param endpoint - API endpoint
   * @param body - Request body
   * @returns Observable of response
   */
  put<T>(endpoint: string, body: any): Observable<T> {
    const url = this.buildUrl(endpoint);
    return this.http.put<T>(url, body)
      .pipe(
        timeout(this.timeout),
        catchError(this.handleError)
      );
  }

  /**
   * PATCH request
   * @param endpoint - API endpoint
   * @param body - Request body
   * @returns Observable of response
   */
  patch<T>(endpoint: string, body: any): Observable<T> {
    const url = this.buildUrl(endpoint);
    return this.http.patch<T>(url, body)
      .pipe(
        timeout(this.timeout),
        catchError(this.handleError)
      );
  }

  /**
   * DELETE request
   * @param endpoint - API endpoint
   * @returns Observable of response
   */
  delete<T>(endpoint: string): Observable<T> {
    const url = this.buildUrl(endpoint);
    return this.http.delete<T>(url)
      .pipe(
        timeout(this.timeout),
        catchError(this.handleError)
      );
  }

  /**
   * Build full URL from endpoint
   * @param endpoint - API endpoint
   * @returns Full URL
   */
  private buildUrl(endpoint: string): string {
    // Remove leading slash if present
    const cleanEndpoint = endpoint.startsWith('/') ? endpoint.slice(1) : endpoint;
    return `${this.baseUrl}/${cleanEndpoint}`;
  }

  /**
   * Handle HTTP errors
   * @param error - HTTP error response
   * @returns Observable error
   */
  private handleError(error: HttpErrorResponse): Observable<never> {
    let errorMessage = 'An unknown error occurred';

    if (error.error instanceof ErrorEvent) {
      // Client-side or network error
      errorMessage = `Network error: ${error.error.message}`;
    } else {
      // Backend returned an unsuccessful response code
      errorMessage = error.error?.message || error.message || `Server error: ${error.status}`;
    }

    if (environment.enableDebugMode) {
      console.error('API Error:', error);
    }

    return throwError(() => new Error(errorMessage));
  }
}
