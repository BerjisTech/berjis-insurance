// Production environment configuration
// This file is used when running `ng build --configuration=production`
export const environment = {
  production: true,
  apiUrl: 'https://api.insurance.local/api/v1',
  apiTimeout: 30000, // 30 seconds
  enableDebugMode: false,
};
