import { Server, Socket } from 'socket.io';
import { AuthenticatedSocket } from '../types';

export interface RoomInfo {
  name: string;
  type: 'user' | 'group' | 'broadcast' | 'system';
  members: Set<string>; // socket IDs
  metadata: Record<string, any>;
  createdAt: Date;
  lastActivity: Date;
}

export interface RoomStats {
  totalRooms: number;
  userRooms: number;
  groupRooms: number;
  broadcastRooms: number;
  systemRooms: number;
  totalMembers: number;
  averageMembersPerRoom: number;
}

export class RoomManager {
  private rooms: Map<string, RoomInfo> = new Map();
  private socketRooms: Map<string, Set<string>> = new Map(); // socketId -> room names
  private cleanupInterval?: NodeJS.Timeout;

  constructor(private io: Server) {
    this.startCleanupInterval();
  }

  /**
   * Create or join a user-specific room
   */
  public joinUserRoom(socket: AuthenticatedSocket): void {
    const roomName = `user:${socket.userId}`;
    this.joinRoom(socket, roomName, 'user', { userId: socket.userId });
  }

  /**
   * Create or join a group room
   */
  public joinGroupRoom(
    socket: AuthenticatedSocket,
    groupId: string,
    metadata?: Record<string, any>
  ): void {
    const roomName = `group:${groupId}`;
    this.joinRoom(socket, roomName, 'group', { groupId, ...metadata });
  }

  /**
   * Join a broadcast room (for system-wide announcements)
   */
  public joinBroadcastRoom(
    socket: AuthenticatedSocket,
    broadcastType: string
  ): void {
    const roomName = `broadcast:${broadcastType}`;
    this.joinRoom(socket, roomName, 'broadcast', { broadcastType });
  }

  /**
   * Join a system room (for admin/monitoring purposes)
   */
  public joinSystemRoom(socket: AuthenticatedSocket, systemType: string): void {
    const roomName = `system:${systemType}`;
    this.joinRoom(socket, roomName, 'system', { systemType });
  }

  /**
   * Generic room join method
   */
  private joinRoom(
    socket: AuthenticatedSocket,
    roomName: string,
    type: RoomInfo['type'],
    metadata: Record<string, any> = {}
  ): void {
    // Join the Socket.IO room
    socket.join(roomName);

    // Track room membership
    if (!this.rooms.has(roomName)) {
      this.rooms.set(roomName, {
        name: roomName,
        type,
        members: new Set(),
        metadata,
        createdAt: new Date(),
        lastActivity: new Date(),
      });
    }

    const room = this.rooms.get(roomName)!;
    room.members.add(socket.id);
    room.lastActivity = new Date();

    // Track socket's room membership
    if (!this.socketRooms.has(socket.id)) {
      this.socketRooms.set(socket.id, new Set());
    }
    this.socketRooms.get(socket.id)!.add(roomName);

    console.log(
      `Socket ${socket.id} joined room ${roomName} (${room.members.size} members)`
    );
  }

  /**
   * Leave a specific room
   */
  public leaveRoom(socket: Socket, roomName: string): void {
    // Leave the Socket.IO room
    socket.leave(roomName);

    // Update room membership
    const room = this.rooms.get(roomName);
    if (room) {
      room.members.delete(socket.id);
      room.lastActivity = new Date();

      // Remove empty rooms (except persistent system rooms)
      if (room.members.size === 0 && room.type !== 'system') {
        this.rooms.delete(roomName);
        console.log(`Removed empty room: ${roomName}`);
      }
    }

    // Update socket's room membership
    const socketRooms = this.socketRooms.get(socket.id);
    if (socketRooms) {
      socketRooms.delete(roomName);
    }

    console.log(`Socket ${socket.id} left room ${roomName}`);
  }

  /**
   * Leave all rooms for a socket
   */
  public leaveAllRooms(socket: Socket): void {
    const socketRooms = this.socketRooms.get(socket.id);
    if (socketRooms) {
      const roomNames = Array.from(socketRooms);
      roomNames.forEach(roomName => {
        this.leaveRoom(socket, roomName);
      });
      this.socketRooms.delete(socket.id);
    }
  }

  /**
   * Send message to a specific room
   */
  public sendToRoom(roomName: string, event: string, data: any): void {
    const room = this.rooms.get(roomName);
    if (room) {
      room.lastActivity = new Date();
      this.io.to(roomName).emit(event, data);
      console.log(
        `Sent ${event} to room ${roomName} (${room.members.size} members)`
      );
    } else {
      console.warn(`Attempted to send to non-existent room: ${roomName}`);
    }
  }

  /**
   * Send message to user's room
   */
  public sendToUser(userId: string, event: string, data: any): void {
    const roomName = `user:${userId}`;
    this.sendToRoom(roomName, event, data);
  }

  /**
   * Send message to group room
   */
  public sendToGroup(groupId: string, event: string, data: any): void {
    const roomName = `group:${groupId}`;
    this.sendToRoom(roomName, event, data);
  }

  /**
   * Broadcast to all users in broadcast rooms
   */
  public broadcast(broadcastType: string, event: string, data: any): void {
    const roomName = `broadcast:${broadcastType}`;
    this.sendToRoom(roomName, event, data);
  }

  /**
   * Get room information
   */
  public getRoomInfo(roomName: string): RoomInfo | undefined {
    return this.rooms.get(roomName);
  }

  /**
   * Get all rooms for a socket
   */
  public getSocketRooms(socketId: string): string[] {
    const rooms = this.socketRooms.get(socketId);
    return rooms ? Array.from(rooms) : [];
  }

  /**
   * Get all members of a room
   */
  public getRoomMembers(roomName: string): string[] {
    const room = this.rooms.get(roomName);
    return room ? Array.from(room.members) : [];
  }

  /**
   * Check if socket is in room
   */
  public isSocketInRoom(socketId: string, roomName: string): boolean {
    const room = this.rooms.get(roomName);
    return room ? room.members.has(socketId) : false;
  }

  /**
   * Get room statistics
   */
  public getRoomStats(): RoomStats {
    const rooms = Array.from(this.rooms.values());
    const totalMembers = rooms.reduce(
      (sum, room) => sum + room.members.size,
      0
    );

    return {
      totalRooms: rooms.length,
      userRooms: rooms.filter(r => r.type === 'user').length,
      groupRooms: rooms.filter(r => r.type === 'group').length,
      broadcastRooms: rooms.filter(r => r.type === 'broadcast').length,
      systemRooms: rooms.filter(r => r.type === 'system').length,
      totalMembers,
      averageMembersPerRoom:
        rooms.length > 0
          ? Math.round((totalMembers / rooms.length) * 100) / 100
          : 0,
    };
  }

  /**
   * Get rooms by type
   */
  public getRoomsByType(type: RoomInfo['type']): RoomInfo[] {
    return Array.from(this.rooms.values()).filter(room => room.type === type);
  }

  /**
   * Get active rooms (with recent activity)
   */
  public getActiveRooms(maxInactiveTime: number = 300000): RoomInfo[] {
    // 5 minutes
    const now = new Date();
    return Array.from(this.rooms.values()).filter(room => {
      const inactiveTime = now.getTime() - room.lastActivity.getTime();
      return inactiveTime <= maxInactiveTime;
    });
  }

  /**
   * Cleanup inactive rooms
   */
  private cleanupInactiveRooms(): void {
    const now = new Date();
    const maxInactiveTime = 3600000; // 1 hour
    const roomsToDelete: string[] = [];

    this.rooms.forEach((room, roomName) => {
      const inactiveTime = now.getTime() - room.lastActivity.getTime();

      // Don't cleanup system rooms or rooms with members
      if (
        room.type !== 'system' &&
        room.members.size === 0 &&
        inactiveTime > maxInactiveTime
      ) {
        roomsToDelete.push(roomName);
      }
    });

    roomsToDelete.forEach(roomName => {
      this.rooms.delete(roomName);
      console.log(`Cleaned up inactive room: ${roomName}`);
    });

    if (roomsToDelete.length > 0) {
      console.log(`Cleaned up ${roomsToDelete.length} inactive rooms`);
    }
  }

  /**
   * Start cleanup interval
   */
  private startCleanupInterval(): void {
    this.cleanupInterval = setInterval(() => {
      this.cleanupInactiveRooms();
    }, 300000); // Check every 5 minutes
  }

  /**
   * Stop room manager
   */
  public stop(): void {
    if (this.cleanupInterval) {
      clearInterval(this.cleanupInterval);
    }
    this.rooms.clear();
    this.socketRooms.clear();
    console.log('Room manager stopped');
  }

  /**
   * Get detailed room information for monitoring
   */
  public getDetailedRoomInfo(): Array<{
    name: string;
    type: string;
    memberCount: number;
    metadata: Record<string, any>;
    createdAt: string;
    lastActivity: string;
    inactiveTime: number;
  }> {
    const now = new Date();

    return Array.from(this.rooms.values()).map(room => ({
      name: room.name,
      type: room.type,
      memberCount: room.members.size,
      metadata: room.metadata,
      createdAt: room.createdAt.toISOString(),
      lastActivity: room.lastActivity.toISOString(),
      inactiveTime: now.getTime() - room.lastActivity.getTime(),
    }));
  }
}
