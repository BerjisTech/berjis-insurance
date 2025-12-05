// Shared form validator patterns and helper messages
// Keeps Angular forms aligned with backend validation rules

export const PASSWORD_COMPLEXITY_PATTERN = /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d).{8,}$/;
export const OTP_CODE_PATTERN = /^[0-9]{6}$/;
export const PHONE_E164_PATTERN = /^\+?[1-9]\d{9,14}$/;

export const PASSWORD_REQUIREMENTS = 'Use at least 8 characters with uppercase, lowercase and a number.';
