import jwt from 'jsonwebtoken';
import { config } from '../config';
import { JWTPayload } from '../types';

export class AuthService {
  /**
   * Validates JWT token and returns decoded payload
   */
  public validateToken(token: string): JWTPayload {
    try {
      const decoded = jwt.verify(token, config.jwtSecret) as JWTPayload;

      // Check if token is expired
      const now = Math.floor(Date.now() / 1000);
      if (decoded.exp && decoded.exp < now) {
        throw new Error('Token expired');
      }

      return decoded;
    } catch (error) {
      if (error instanceof jwt.JsonWebTokenError) {
        throw new Error('Invalid token');
      }
      if (error instanceof jwt.TokenExpiredError) {
        throw new Error('Token expired');
      }
      throw error;
    }
  }

  /**
   * Extracts token from socket handshake
   */
  public extractTokenFromSocket(socket: any): string {
    // Try auth object first
    if (socket.handshake.auth?.token) {
      return socket.handshake.auth.token;
    }

    // Try query parameters
    if (socket.handshake.query?.token) {
      return socket.handshake.query.token as string;
    }

    // Try headers
    const authHeader = socket.handshake.headers.authorization;
    if (authHeader && authHeader.startsWith('Bearer ')) {
      return authHeader.substring(7);
    }

    throw new Error('No token provided');
  }

  /**
   * Validates user permissions for specific actions
   */
  public hasPermission(
    userId: string,
    action: string,
    resourceId?: string
  ): boolean {
    // Basic permission check - can be extended with role-based access control
    switch (action) {
      case 'view_progress':
      case 'receive_notifications':
        return true; // All authenticated users can view their own progress
      case 'admin_broadcast':
        // Only admin users can broadcast to all users
        return false; // Implement admin check here
      default:
        return false;
    }
  }
}
