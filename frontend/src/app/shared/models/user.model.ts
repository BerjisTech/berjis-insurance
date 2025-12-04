// User domain models and interfaces
// Defines TypeScript interfaces for user-related data structures

export interface User {
  id: string;
  email: string;
  phone?: string;
  role: UserRole;
  status: UserStatus;
  emailVerified: boolean;
  phoneVerified: boolean;
  isActive?: boolean;
  lastLoginAt?: Date;
  createdAt: Date;
  updatedAt: Date;
  profile?: UserProfile;
}

export interface UserProfile {
  userId: string;
  fullName?: string;
  dateOfBirth?: Date;
  gender?: Gender;
  occupation?: string;
  idNumber?: string;
  profilePhotoUrl?: string;
}

export interface UserAddress {
  id: string;
  userId: string;
  type: AddressType;
  street?: string;
  city?: string;
  stateProvince?: string;
  postalCode?: string;
  country: string;
  isPrimary: boolean;
}

export interface UserDocument {
  id: string;
  userId: string;
  documentType: DocumentType;
  fileUrl: string;
  fileSize?: number;
  mimeType?: string;
  verificationStatus: VerificationStatus;
  verifiedAt?: Date;
  verifiedBy?: string;
  rejectionReason?: string;
  uploadedAt: Date;
}

export interface UserPreferences {
  userId: string;
  language: string;
  currency: string;
  timezone: string;
  notificationEmail: boolean;
  notificationSms: boolean;
  notificationPush: boolean;
  dataSaverMode: boolean;
}

// Enums
export type UserRole = 'user' | 'broker' | 'provider_admin' | 'system_admin';

export type Gender = 'male' | 'female' | 'other' | 'prefer_not_to_say';

export type AddressType = 'home' | 'work' | 'billing' | 'shipping';

export type DocumentType =
  | 'national_id'
  | 'passport'
  | 'proof_of_address'
  | 'drivers_license'
  | 'birth_certificate';

export type VerificationStatus = 'pending' | 'verified' | 'rejected';

export type UserStatus = 'active' | 'inactive' | 'suspended' | 'pending_verification';
