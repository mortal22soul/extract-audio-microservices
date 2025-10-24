import { Server, Socket } from 'socket.io';
import { AuthenticatedSocket } from '../types';
import { config } from '../config';

export interface HeartbeatConfig {
  interval: number;
  timeout: number;
  maxMissedBeats: number;
}

export interface ConnectionHealth {
  socketId: string;
  userId: string;
  lastHeartbeat: Date;
  missedBeats: number;
  isHealthy: boolean;
  latency: number;
}

export class HeartbeatManager {
  private heartbeatConfig: HeartbeatConfig;
  private connectionHealth: Map<string, ConnectionHealth> = new Map();
  private heartbeatInterval?: NodeJS.Timeout;
  private cleanupInterval?: NodeJS.Timeout;

  constructor(
    private io: Server,
    config?: Partial<HeartbeatConfig>
  ) {
    this.heartbeatConfig = {
      interval: config?.interval || 30000, // 30 seconds
      timeout: config?.timeout || 10000, // 10 seconds
      maxMissedBeats: config?.maxMissedBeats || 3,
    };
  }

  /**
   * Start heartbeat monitoring
   */
  public start(): void {
    console.log('Starting heartbeat manager...');

    // Start heartbeat interval
    this.heartbeatInterval = setInterval(() => {
      this.sendHeartbeats();
    }, this.heartbeatConfig.interval);

    // Start cleanup interval
    this.cleanupInterval = setInterval(() => {
      this.cleanupUnhealthyConnections();
    }, this.heartbeatConfig.interval / 2);
  }

  /**
   * Stop heartbeat monitoring
   */
  public stop(): void {
    if (this.heartbeatInterval) {
      clearInterval(this.heartbeatInterval);
    }
    if (this.cleanupInterval) {
      clearInterval(this.cleanupInterval);
    }
    this.connectionHealth.clear();
    console.log('Heartbeat manager stopped');
  }

  /**
   * Add connection to heartbeat monitoring
   */
  public addConnection(socket: AuthenticatedSocket): void {
    const health: ConnectionHealth = {
      socketId: socket.id,
      userId: socket.userId,
      lastHeartbeat: new Date(),
      missedBeats: 0,
      isHealthy: true,
      latency: 0,
    };

    this.connectionHealth.set(socket.id, health);
    this.setupSocketHeartbeat(socket);

    console.log(`Added connection ${socket.id} to heartbeat monitoring`);
  }

  /**
   * Remove connection from heartbeat monitoring
   */
  public removeConnection(socketId: string): void {
    this.connectionHealth.delete(socketId);
    console.log(`Removed connection ${socketId} from heartbeat monitoring`);
  }

  /**
   * Setup heartbeat handlers for a socket
   */
  private setupSocketHeartbeat(socket: AuthenticatedSocket): void {
    // Handle heartbeat response
    socket.on('heartbeat:pong', (data: { timestamp: number }) => {
      const health = this.connectionHealth.get(socket.id);
      if (health) {
        const now = Date.now();
        health.lastHeartbeat = new Date();
        health.missedBeats = 0;
        health.isHealthy = true;
        health.latency = now - data.timestamp;

        console.log(
          `Heartbeat received from ${socket.userId}, latency: ${health.latency}ms`
        );
      }
    });

    // Handle manual ping from client
    socket.on('ping', () => {
      const health = this.connectionHealth.get(socket.id);
      if (health) {
        health.lastHeartbeat = new Date();
        health.missedBeats = 0;
      }
      socket.emit('pong', { timestamp: Date.now() });
    });

    // Handle connection quality report
    socket.on(
      'connection:quality',
      (data: { quality: 'good' | 'poor' | 'unstable' }) => {
        console.log(
          `Connection quality report from ${socket.userId}: ${data.quality}`
        );
        // Could be used for adaptive heartbeat intervals
      }
    );
  }

  /**
   * Send heartbeat to all connected clients
   */
  private sendHeartbeats(): void {
    const timestamp = Date.now();

    this.connectionHealth.forEach((health, socketId) => {
      const socket = this.io.sockets.sockets.get(socketId);

      if (socket) {
        // Send heartbeat ping
        socket.emit('heartbeat:ping', { timestamp });

        // Update missed beats counter
        const timeSinceLastBeat = Date.now() - health.lastHeartbeat.getTime();
        if (timeSinceLastBeat > this.heartbeatConfig.timeout) {
          health.missedBeats++;
          health.isHealthy =
            health.missedBeats <= this.heartbeatConfig.maxMissedBeats;

          console.log(
            `Missed heartbeat from ${health.userId} (${health.missedBeats}/${this.heartbeatConfig.maxMissedBeats})`
          );
        }
      } else {
        // Socket no longer exists, mark for cleanup
        health.isHealthy = false;
        health.missedBeats = this.heartbeatConfig.maxMissedBeats + 1;
      }
    });
  }

  /**
   * Cleanup unhealthy connections
   */
  private cleanupUnhealthyConnections(): void {
    const unhealthyConnections: string[] = [];

    this.connectionHealth.forEach((health, socketId) => {
      if (
        !health.isHealthy &&
        health.missedBeats > this.heartbeatConfig.maxMissedBeats
      ) {
        unhealthyConnections.push(socketId);
      }
    });

    unhealthyConnections.forEach(socketId => {
      const socket = this.io.sockets.sockets.get(socketId);
      const health = this.connectionHealth.get(socketId);

      if (health) {
        console.log(
          `Disconnecting unhealthy connection: ${health.userId} (${socketId})`
        );

        if (socket) {
          socket.emit('connection:unhealthy', {
            reason: 'missed_heartbeats',
            missedBeats: health.missedBeats,
            maxAllowed: this.heartbeatConfig.maxMissedBeats,
          });
          socket.disconnect(true);
        }

        this.connectionHealth.delete(socketId);
      }
    });
  }

  /**
   * Get connection health statistics
   */
  public getHealthStats(): {
    totalConnections: number;
    healthyConnections: number;
    unhealthyConnections: number;
    averageLatency: number;
    connectionsByLatency: {
      excellent: number; // < 50ms
      good: number; // 50-150ms
      fair: number; // 150-300ms
      poor: number; // > 300ms
    };
  } {
    const connections = Array.from(this.connectionHealth.values());
    const healthyConnections = connections.filter(c => c.isHealthy);
    const unhealthyConnections = connections.filter(c => !c.isHealthy);

    const totalLatency = healthyConnections.reduce(
      (sum, c) => sum + c.latency,
      0
    );
    const averageLatency =
      healthyConnections.length > 0
        ? totalLatency / healthyConnections.length
        : 0;

    const connectionsByLatency = {
      excellent: healthyConnections.filter(c => c.latency < 50).length,
      good: healthyConnections.filter(c => c.latency >= 50 && c.latency < 150)
        .length,
      fair: healthyConnections.filter(c => c.latency >= 150 && c.latency < 300)
        .length,
      poor: healthyConnections.filter(c => c.latency >= 300).length,
    };

    return {
      totalConnections: connections.length,
      healthyConnections: healthyConnections.length,
      unhealthyConnections: unhealthyConnections.length,
      averageLatency: Math.round(averageLatency),
      connectionsByLatency,
    };
  }

  /**
   * Get health status for specific connection
   */
  public getConnectionHealth(socketId: string): ConnectionHealth | undefined {
    return this.connectionHealth.get(socketId);
  }

  /**
   * Get all connection health data
   */
  public getAllConnectionHealth(): ConnectionHealth[] {
    return Array.from(this.connectionHealth.values());
  }

  /**
   * Force health check for all connections
   */
  public forceHealthCheck(): void {
    console.log('Forcing health check for all connections...');
    this.sendHeartbeats();
  }

  /**
   * Update heartbeat configuration
   */
  public updateConfig(newConfig: Partial<HeartbeatConfig>): void {
    this.heartbeatConfig = { ...this.heartbeatConfig, ...newConfig };
    console.log('Heartbeat configuration updated:', this.heartbeatConfig);

    // Restart with new configuration
    this.stop();
    this.start();
  }
}
