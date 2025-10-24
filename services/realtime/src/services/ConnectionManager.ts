import { Server, Socket } from 'socket.io';
import { AuthenticatedSocket, UserSession } from '../types';
import { RedisService } from './RedisService';
import { HeartbeatManager } from './HeartbeatManager';
import { RoomManager } from './RoomManager';

export class ConnectionManager {
  private userConnections: Map<string, Set<string>> = new Map(); // userId -> Set of socketIds
  private socketUsers: Map<string, string> = new Map(); // socketId -> userId
  private userSessions: Map<string, UserSession> = new Map();
  private connectionCount: number = 0;
  private heartbeatManager: HeartbeatManager;
  private roomManager: RoomManager;

  constructor(
    private io: Server,
    private redisService: RedisService
  ) {
    this.heartbeatManager = new HeartbeatManager(this.io);
    this.roomManager = new RoomManager(this.io);
    this.heartbeatManager.start();
  }

  /**
   * Add a new user connection
   */
  public addConnection(socket: AuthenticatedSocket): void {
    const { userId } = socket;
    const socketId = socket.id;

    // Track socket to user mapping
    this.socketUsers.set(socketId, userId);

    // Track user connections
    if (!this.userConnections.has(userId)) {
      this.userConnections.set(userId, new Set());
    }
    this.userConnections.get(userId)!.add(socketId);

    // Create user session
    const session: UserSession = {
      userId,
      socketId,
      connectedAt: new Date(),
      lastActivity: new Date(),
    };
    this.userSessions.set(socketId, session);

    // Add to heartbeat monitoring
    this.heartbeatManager.addConnection(socket);

    // Join user-specific room and default broadcast room
    this.roomManager.joinUserRoom(socket);
    this.roomManager.joinBroadcastRoom(socket, 'general');

    // Store session in Redis for persistence across instances
    this.redisService.storeUserSession(userId, {
      socketId,
      connectedAt: session.connectedAt.toISOString(),
      lastActivity: session.lastActivity.toISOString(),
    });

    this.connectionCount++;
    console.log(
      `User ${userId} connected (${socketId}). Total connections: ${this.connectionCount}`
    );
  }

  /**
   * Remove a user connection
   */
  public removeConnection(socket: Socket): void {
    const socketId = socket.id;
    const userId = this.socketUsers.get(socketId);

    if (!userId) {
      return;
    }

    // Remove from tracking maps
    this.socketUsers.delete(socketId);
    this.userSessions.delete(socketId);

    // Remove from user connections
    const userSockets = this.userConnections.get(userId);
    if (userSockets) {
      userSockets.delete(socketId);
      if (userSockets.size === 0) {
        this.userConnections.delete(userId);
        // Remove session from Redis when user has no more connections
        this.redisService.removeUserSession(userId);
      }
    }

    // Remove from heartbeat monitoring
    this.heartbeatManager.removeConnection(socketId);

    // Leave all rooms
    this.roomManager.leaveAllRooms(socket);

    this.connectionCount--;
    console.log(
      `User ${userId} disconnected (${socketId}). Total connections: ${this.connectionCount}`
    );
  }

  /**
   * Update user activity timestamp
   */
  public updateUserActivity(socketId: string): void {
    const session = this.userSessions.get(socketId);
    if (session) {
      session.lastActivity = new Date();
    }
  }

  /**
   * Get all socket IDs for a user
   */
  public getUserSockets(userId: string): string[] {
    const sockets = this.userConnections.get(userId);
    return sockets ? Array.from(sockets) : [];
  }

  /**
   * Get user ID from socket ID
   */
  public getUserFromSocket(socketId: string): string | undefined {
    return this.socketUsers.get(socketId);
  }

  /**
   * Check if user is connected
   */
  public isUserConnected(userId: string): boolean {
    return (
      this.userConnections.has(userId) &&
      this.userConnections.get(userId)!.size > 0
    );
  }

  /**
   * Get total connection count
   */
  public getConnectionCount(): number {
    return this.connectionCount;
  }

  /**
   * Get connected users count
   */
  public getConnectedUsersCount(): number {
    return this.userConnections.size;
  }

  /**
   * Send message to specific user
   */
  public sendToUser(userId: string, event: string, data: any): void {
    this.roomManager.sendToUser(userId, event, data);
  }

  /**
   * Send message to all connected users
   */
  public broadcast(event: string, data: any): void {
    this.roomManager.broadcast('general', event, data);
  }

  /**
   * Send message to a group
   */
  public sendToGroup(groupId: string, event: string, data: any): void {
    this.roomManager.sendToGroup(groupId, event, data);
  }

  /**
   * Join user to a group room
   */
  public joinGroup(userId: string, groupId: string): void {
    const sockets = this.getUserSockets(userId);
    sockets.forEach(socketId => {
      const socket = this.io.sockets.sockets.get(
        socketId
      ) as AuthenticatedSocket;
      if (socket) {
        this.roomManager.joinGroupRoom(socket, groupId);
      }
    });
  }

  /**
   * Remove user from a group room
   */
  public leaveGroup(userId: string, groupId: string): void {
    const sockets = this.getUserSockets(userId);
    sockets.forEach(socketId => {
      const socket = this.io.sockets.sockets.get(socketId);
      if (socket) {
        this.roomManager.leaveRoom(socket, `group:${groupId}`);
      }
    });
  }

  /**
   * Get user session information
   */
  public getUserSession(socketId: string): UserSession | undefined {
    return this.userSessions.get(socketId);
  }

  /**
   * Get all active sessions
   */
  public getAllSessions(): UserSession[] {
    return Array.from(this.userSessions.values());
  }

  /**
   * Clean up inactive connections
   */
  public cleanupInactiveConnections(maxInactiveTime: number = 300000): void {
    // 5 minutes
    const now = new Date();
    const inactiveSessions: string[] = [];

    this.userSessions.forEach((session, socketId) => {
      const inactiveTime = now.getTime() - session.lastActivity.getTime();
      if (inactiveTime > maxInactiveTime) {
        inactiveSessions.push(socketId);
      }
    });

    inactiveSessions.forEach(socketId => {
      const socket = this.io.sockets.sockets.get(socketId);
      if (socket) {
        console.log(`Disconnecting inactive socket: ${socketId}`);
        socket.disconnect(true);
      }
    });
  }

  /**
   * Get connection statistics
   */
  public getStats(): {
    totalConnections: number;
    connectedUsers: number;
    averageConnectionsPerUser: number;
    heartbeatStats?: any;
    roomStats?: any;
  } {
    const totalConnections = this.connectionCount;
    const connectedUsers = this.userConnections.size;
    const averageConnectionsPerUser =
      connectedUsers > 0 ? totalConnections / connectedUsers : 0;

    return {
      totalConnections,
      connectedUsers,
      averageConnectionsPerUser:
        Math.round(averageConnectionsPerUser * 100) / 100,
      heartbeatStats: this.heartbeatManager.getHealthStats(),
      roomStats: this.roomManager.getRoomStats(),
    };
  }

  /**
   * Get detailed connection health information
   */
  public getHealthInfo(): {
    connections: any[];
    rooms: any[];
    heartbeat: any;
  } {
    return {
      connections: this.getAllSessions(),
      rooms: this.roomManager.getDetailedRoomInfo(),
      heartbeat: this.heartbeatManager.getAllConnectionHealth(),
    };
  }

  /**
   * Cleanup all managers
   */
  public cleanup(): void {
    this.heartbeatManager.stop();
    this.roomManager.stop();
    this.userConnections.clear();
    this.socketUsers.clear();
    this.userSessions.clear();
    this.connectionCount = 0;
  }
}
