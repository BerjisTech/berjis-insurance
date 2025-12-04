# Insurance Broker Platform - Frontend Dockerfile
# Multi-stage build for optimized Angular 18 production
# Node 20 LTS with standalone components

# =============================================================================
# Stage 1: Builder
# =============================================================================
FROM node:20-alpine AS builder

# Set working directory
WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies
RUN npm ci --legacy-peer-deps

# Copy source code
COPY . .

# Build Angular app for production
# --configuration=production: optimized build
# --output-path: custom output directory
RUN npm run build -- --configuration=production

# =============================================================================
# Stage 2: Runtime with Nginx
# =============================================================================
FROM nginx:1.27-alpine

# Copy custom nginx configuration
COPY nginx.conf /etc/nginx/nginx.conf

# Copy built Angular app from builder
COPY --from=builder /app/dist/insurance-broker/browser /usr/share/nginx/html

# Create non-root user for security
RUN addgroup -g 1001 -S nginx_custom && \
    adduser -u 1001 -S nginx_custom -G nginx_custom && \
    chown -R nginx_custom:nginx_custom /usr/share/nginx/html && \
    chown -R nginx_custom:nginx_custom /var/cache/nginx && \
    chown -R nginx_custom:nginx_custom /var/log/nginx && \
    chmod -R 755 /usr/share/nginx/html && \
    touch /var/run/nginx.pid && \
    chown nginx_custom:nginx_custom /var/run/nginx.pid

# Switch to non-root user
USER nginx_custom

# Expose port 80
EXPOSE 80

# Health check
HEALTHCHECK --interval=30s --timeout=10s --retries=3 \
    CMD wget --quiet --tries=1 --spider http://localhost:80 || exit 1

# Start nginx
CMD ["nginx", "-g", "daemon off;"]
