export interface AuthenticatedSocket extends Socket {
  userId: string;
  userEmail?: string;
}

export interface JWTPayload {
  userId: string;
  email: string;
  iat: number;
  exp: number;
}

export interface ConversionProgressData {
  userId: string;
  videoId: string;
  jobId: string;
  progress: number;
  status: 'processing' | 'completed' | 'failed';
  estimatedTime?: number;
  errorMessage?: string;
}

export interface UserSession {
  userId: string;
  socketId: string;
  connectedAt: Date;
  lastActivity: Date;
}

export interface RedisMessage {
  userId: string;
  data: any;
  timestamp: number;
}

import { Socket } from 'socket.io';
