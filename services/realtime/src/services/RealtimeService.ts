import { Server, Socket } from 'socket.io';
import { createServer } from 'http';
import { config } from '../config';
import { AuthService } from './AuthService';
import { RedisService } from './RedisService';
import { ConnectionManager } from './ConnectionManager';
import { EventHandler } from './EventHandler';
import { AuthenticatedSocket, ConversionProgressData } from '../types';

export class RealtimeService {
  private httpServer;
  private io: Server;
  private authService: AuthService;
  private redisService: RedisService;
  private connectionManager: ConnectionManager;
  private eventHandler: EventHandler;
  private heartbeatInterval?: NodeJS.Timeout;

  constructor() {
    this.httpServer = createServer();
    this.io = new Server(this.httpServer, {
      cors: {
        origin: config.corsOrigin,
        methods: ['GET', 'POST'],
        credentials: true,
      },
      pingTimeout: config.connectionTimeout,
      pingInterval: config.heartbeatInterval,
      maxHttpBufferSize: 1e6, // 1MB
      transports: ['websocket', 'polling'],
    });

    this.authService = new AuthService();
    this.redisService = new RedisService();
    this.connectionManager = new ConnectionManager(this.io, this.redisService);
    this.eventHandler = new EventHandler(
      this.connectionManager,
      this.redisService
    );

    this.setupMiddleware();
    this.setupEventHandlers();
    this.setupRedisSubscriptions();
  }

  /**
   * Setup Socket.IO middleware
   */
  private setupMiddleware(): void {
    // Authentication middleware
    this.io.use(async (socket: Socket, next) => {
      try {
        const token = this.authService.extractTokenFromSocket(socket);
        const decoded = this.authService.validateToken(token);

        // Attach user info to socket
        (socket as AuthenticatedSocket).userId = decoded.userId;
        (socket as AuthenticatedSocket).userEmail = decoded.email;

        console.log(`Authentication successful for user: ${decoded.userId}`);
        next();
      } catch (error) {
        console.error('Authentication failed:', error);
        next(new Error('Authentication failed'));
      }
    });

    // Rate limiting middleware (basic implementation)
    this.io.use((socket: Socket, next) => {
      // Implement rate limiting logic here if needed
      next();
    });
  }

  /**
   * Setup Socket.IO event handlers
   */
  private setupEventHandlers(): void {
    this.io.on('connection', (socket: Socket) => {
      const authenticatedSocket = socket as AuthenticatedSocket;

      // Add connection to manager
      this.connectionManager.addConnection(authenticatedSocket);

      // Setup socket event handlers
      this.setupSocketHandlers(authenticatedSocket);

      // Send connection confirmation
      socket.emit('connected', {
        userId: authenticatedSocket.userId,
        timestamp: new Date().toISOString(),
        message: 'Successfully connected to realtime service',
      });
    });

    // Handle server-level events
    this.io.on('disconnect', (socket: Socket) => {
      this.connectionManager.removeConnection(socket);
    });
  }

  /**
   * Setup individual socket event handlers
   */
  private setupSocketHandlers(socket: AuthenticatedSocket): void {
    // Handle disconnection
    socket.on('disconnect', reason => {
      console.log(`User ${socket.userId} disconnected: ${reason}`);
      this.connectionManager.removeConnection(socket);
    });

    // Handle ping/pong for connection health
    socket.on('ping', () => {
      this.connectionManager.updateUserActivity(socket.id);
      socket.emit('pong', { timestamp: new Date().toISOString() });
    });

    // Handle user activity updates
    socket.on('activity', () => {
      this.connectionManager.updateUserActivity(socket.id);
    });

    // Handle subscription to specific events
    socket.on('subscribe', (data: { events: string[] }) => {
      if (data.events && Array.isArray(data.events)) {
        data.events.forEach(event => {
          if (this.isValidSubscriptionEvent(event)) {
            socket.join(event);
            console.log(`User ${socket.userId} subscribed to ${event}`);
          }
        });
      }
    });

    // Handle unsubscription from events
    socket.on('unsubscribe', (data: { events: string[] }) => {
      if (data.events && Array.isArray(data.events)) {
        data.events.forEach(event => {
          socket.leave(event);
          console.log(`User ${socket.userId} unsubscribed from ${event}`);
        });
      }
    });

    // Handle client errors
    socket.on('error', error => {
      console.error(`Socket error for user ${socket.userId}:`, error);
    });

    // Handle custom events
    socket.on('request_status', async () => {
      // Send current status to user
      const stats = this.connectionManager.getStats();
      socket.emit('status', {
        connected: true,
        userId: socket.userId,
        connectionStats: stats,
        timestamp: new Date().toISOString(),
      });
    });
  }

  /**
   * Setup Redis subscriptions for cross-service communication
   */
  private setupRedisSubscriptions(): void {
    const channels = Object.values(config.redisChannels);

    this.redisService.onMessage((channel: string, message: string) => {
      try {
        const data = JSON.parse(message);
        this.handleRedisMessage(channel, data);
      } catch (error) {
        console.error('Failed to parse Redis message:', error);
      }
    });

    // Subscribe to channels after Redis connection is established
    this.redisService
      .connect()
      .then(() => this.redisService.subscribeToChannels(channels))
      .catch(error => {
        console.error('Failed to setup Redis subscriptions:', error);
      });
  }

  /**
   * Handle incoming Redis messages
   */
  private handleRedisMessage(channel: string, data: any): void {
    switch (channel) {
      case config.redisChannels.conversionProgress:
        this.handleConversionProgress(data);
        break;

      case config.redisChannels.conversionComplete:
        this.handleConversionComplete(data);
        break;

      case config.redisChannels.conversionError:
        this.handleConversionError(data);
        break;

      case config.redisChannels.notification:
        this.handleNotification(data);
        break;

      case config.redisChannels.analytics:
        this.handleAnalyticsUpdate(data);
        break;

      default:
        console.warn(`Unhandled Redis channel: ${channel}`);
    }
  }

  /**
   * Handle conversion progress updates
   */
  private async handleConversionProgress(
    data: ConversionProgressData
  ): Promise<void> {
    await this.eventHandler.handleConversionProgress(data);
  }

  /**
   * Handle conversion completion
   */
  private async handleConversionComplete(
    data: ConversionProgressData
  ): Promise<void> {
    await this.eventHandler.handleConversionComplete(data);
  }

  /**
   * Handle conversion errors
   */
  private async handleConversionError(
    data: ConversionProgressData
  ): Promise<void> {
    await this.eventHandler.handleConversionError(data);
  }

  /**
   * Handle general notifications
   */
  private handleNotification(data: any): void {
    if (data.userId) {
      // Send to specific user
      this.connectionManager.sendToUser(data.userId, 'notification', data);
    } else if (data.broadcast) {
      // Broadcast to all users
      this.connectionManager.broadcast('notification', data);
    }
  }

  /**
   * Handle analytics updates
   */
  private handleAnalyticsUpdate(data: any): void {
    if (data.userId) {
      this.connectionManager.sendToUser(data.userId, 'analytics:update', data);
    }
  }

  /**
   * Validate subscription event names
   */
  private isValidSubscriptionEvent(event: string): boolean {
    const validEvents = [
      'conversion:progress',
      'conversion:complete',
      'conversion:error',
      'notification',
      'analytics:update',
    ];
    return validEvents.includes(event);
  }

  /**
   * Start the realtime service
   */
  public async start(): Promise<void> {
    try {
      // Start cleanup interval for inactive connections
      this.heartbeatInterval = setInterval(() => {
        this.connectionManager.cleanupInactiveConnections();
      }, 60000); // Check every minute

      // Start HTTP server
      this.httpServer.listen(config.port, () => {
        console.log(`Realtime service listening on port ${config.port}`);
        console.log(`Environment: ${config.nodeEnv}`);
        console.log(`CORS origin: ${config.corsOrigin}`);
      });
    } catch (error) {
      console.error('Failed to start realtime service:', error);
      throw error;
    }
  }

  /**
   * Stop the realtime service
   */
  public async stop(): Promise<void> {
    try {
      // Clear intervals
      if (this.heartbeatInterval) {
        clearInterval(this.heartbeatInterval);
      }

      // Cleanup event handler
      this.eventHandler.cleanup();

      // Cleanup connection manager
      this.connectionManager.cleanup();

      // Close Socket.IO server
      this.io.close();

      // Close HTTP server
      this.httpServer.close();

      // Disconnect from Redis
      await this.redisService.disconnect();

      console.log('Realtime service stopped');
    } catch (error) {
      console.error('Error stopping realtime service:', error);
      throw error;
    }
  }

  /**
   * Get service health status
   */
  public getHealthStatus(): {
    status: string;
    connections: number;
    users: number;
    redis: boolean;
    uptime: number;
  } {
    const stats = this.connectionManager.getStats();

    return {
      status: 'healthy',
      connections: stats.totalConnections,
      users: stats.connectedUsers,
      redis: this.redisService.isRedisConnected(),
      uptime: process.uptime(),
    };
  }
}
