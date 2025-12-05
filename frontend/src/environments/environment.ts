// Production environment configuration
// This file is used when running `ng build --configuration=production`
export const environment = {
  production: true,
  apiUrl: 'http://localhost:8096/api/v1',
  apiTimeout: 30000, // 30 seconds
  enableDebugMode: false,
};
