// User types
export interface User {
  id: number;
  email: string;
  firstName?: string;
  lastName?: string;
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
}

// Authentication types
export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  email: string;
  password: string;
  firstName?: string;
  lastName?: string;
}

export interface AuthResponse {
  token: string;
  user: User;
}

// Video types
export interface Video {
  id: string;
  userId: number;
  originalFilename: string;
  mimeType: string;
  size: number;
  uploadedAt: string;
  status: 'uploaded' | 'processing' | 'completed' | 'failed';
  conversionJobId?: string;
  mp3FileId?: string;
  metadata?: VideoMetadata;
  analytics?: VideoAnalytics;
}

export interface VideoMetadata {
  duration: number;
  resolution: string;
  codec: string;
  bitrate: number;
}

export interface VideoAnalytics {
  thumbnails: string[];
  qualityScore: number;
  safetyScore: number;
  tags: string[];
}

// Conversion types
export interface ConversionJob {
  id: string;
  videoId: string;
  userId: number;
  status: 'pending' | 'processing' | 'completed' | 'failed';
  progress: number;
  startedAt?: string;
  completedAt?: string;
  errorMessage?: string;
  processingNode?: string;
}

// API Response types
export interface ApiResponse<T = any> {
  data?: T;
  error?: string;
  message?: string;
}

// WebSocket event types
export interface ProgressEvent {
  jobId: string;
  videoId: string;
  userId: number;
  progress: number;
  status: string;
  estimatedTime?: number;
}

export interface CompletionEvent {
  jobId: string;
  videoId: string;
  userId: number;
  mp3FileId: string;
  downloadUrl: string;
}

export interface ErrorEvent {
  jobId: string;
  videoId: string;
  userId: number;
  error: string;
  details?: string;
}
