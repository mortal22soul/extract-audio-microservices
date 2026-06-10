import Redis from 'ioredis';
import { config } from '../config';
import { RedisMessage } from '../types';

export class RedisService {
  private subscriber: Redis;
  private publisher: Redis;
  private isConnected: boolean = false;

  constructor() {
    this.subscriber = new Redis(config.redisUrl, {
      maxRetriesPerRequest: 3,
      lazyConnect: true,
    });

    this.publisher = new Redis(config.redisUrl, {
      maxRetriesPerRequest: 3,
      lazyConnect: true,
    });

    this.setupEventHandlers();
  }

  private setupEventHandlers(): void {
    this.subscriber.on('connect', () => {
      console.log('Redis subscriber connected');
      this.isConnected = true;
    });

    this.subscriber.on('error', error => {
      console.error('Redis subscriber error:', error);
      this.isConnected = false;
    });

    this.publisher.on('connect', () => {
      console.log('Redis publisher connected');
    });

    this.publisher.on('error', error => {
      console.error('Redis publisher error:', error);
    });

    this.subscriber.on('reconnecting', () => {
      console.log('Redis subscriber reconnecting...');
    });
  }

  /**
   * Connect to Redis
   */
  public async connect(): Promise<void> {
    try {
      await Promise.all([this.subscriber.connect(), this.publisher.connect()]);
      console.log('Redis connections established');
    } catch (error) {
      console.error('Failed to connect to Redis:', error);
      throw error;
    }
  }

  /**
   * Subscribe to Redis channels
   */
  public async subscribeToChannels(channels: string[]): Promise<void> {
    try {
      await this.subscriber.subscribe(...channels);
      console.log('Subscribed to channels:', channels);
    } catch (error) {
      console.error('Failed to subscribe to channels:', error);
      throw error;
    }
  }

  /**
   * Set message handler for Redis messages
   */
  public onMessage(handler: (channel: string, message: string) => void): void {
    this.subscriber.on('message', handler);
  }

  /**
   * Publish message to Redis channel
   */
  public async publishMessage(
    channel: string,
    message: RedisMessage
  ): Promise<void> {
    try {
      await this.publisher.publish(channel, JSON.stringify(message));
    } catch (error) {
      console.error('Failed to publish message:', error);
      throw error;
    }
  }

  /**
   * Store user session in Redis
   */
  public async storeUserSession(
    userId: string,
    sessionData: any
  ): Promise<void> {
    try {
      const key = `user:session:${userId}`;
      await this.publisher.setex(key, 3600, JSON.stringify(sessionData)); // 1 hour TTL
    } catch (error) {
      console.error('Failed to store user session:', error);
    }
  }

  /**
   * Get user session from Redis
   */
  public async getUserSession(userId: string): Promise<any | null> {
    try {
      const key = `user:session:${userId}`;
      const data = await this.publisher.get(key);
      return data ? JSON.parse(data) : null;
    } catch (error) {
      console.error('Failed to get user session:', error);
      return null;
    }
  }

  /**
   * Remove user session from Redis
   */
  public async removeUserSession(userId: string): Promise<void> {
    try {
      const key = `user:session:${userId}`;
      await this.publisher.del(key);
    } catch (error) {
      console.error('Failed to remove user session:', error);
    }
  }

  /**
   * Check if Redis is connected
   */
  public isRedisConnected(): boolean {
    return this.isConnected && this.subscriber.status === 'ready';
  }

  /**
   * Gracefully disconnect from Redis
   */
  public async disconnect(): Promise<void> {
    try {
      await Promise.all([
        this.subscriber.disconnect(),
        this.publisher.disconnect(),
      ]);
      console.log('Redis connections closed');
    } catch (error) {
      console.error('Error disconnecting from Redis:', error);
    }
  }
}
