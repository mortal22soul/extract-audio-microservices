import { ConversionProgressData } from '../types';
import { ConnectionManager } from './ConnectionManager';
import { RedisService } from './RedisService';

export interface RetryConfig {
  maxRetries: number;
  retryDelay: number;
  backoffMultiplier: number;
}

export interface EventMetrics {
  totalEvents: number;
  successfulEvents: number;
  failedEvents: number;
  retryAttempts: number;
}

export class EventHandler {
  private retryConfig: RetryConfig = {
    maxRetries: 3,
    retryDelay: 1000, // 1 second
    backoffMultiplier: 2,
  };

  private metrics: EventMetrics = {
    totalEvents: 0,
    successfulEvents: 0,
    failedEvents: 0,
    retryAttempts: 0,
  };

  private retryQueue: Map<
    string,
    { data: any; attempts: number; nextRetry: number }
  > = new Map();
  private retryInterval?: NodeJS.Timeout;

  constructor(
    private connectionManager: ConnectionManager,
    private redisService: RedisService
  ) {
    this.startRetryProcessor();
  }

  /**
   * Handle conversion progress events with retry logic
   */
  public async handleConversionProgress(
    data: ConversionProgressData
  ): Promise<void> {
    this.metrics.totalEvents++;

    try {
      await this.processConversionProgress(data);
      this.metrics.successfulEvents++;
    } catch (error) {
      console.error('Failed to handle conversion progress:', error);
      this.metrics.failedEvents++;
      await this.scheduleRetry('conversion:progress', data);
    }
  }

  /**
   * Handle conversion completion events with retry logic
   */
  public async handleConversionComplete(
    data: ConversionProgressData
  ): Promise<void> {
    this.metrics.totalEvents++;

    try {
      await this.processConversionComplete(data);
      this.metrics.successfulEvents++;
    } catch (error) {
      console.error('Failed to handle conversion completion:', error);
      this.metrics.failedEvents++;
      await this.scheduleRetry('conversion:complete', data);
    }
  }

  /**
   * Handle conversion error events with retry logic
   */
  public async handleConversionError(
    data: ConversionProgressData
  ): Promise<void> {
    this.metrics.totalEvents++;

    try {
      await this.processConversionError(data);
      this.metrics.successfulEvents++;
    } catch (error) {
      console.error('Failed to handle conversion error:', error);
      this.metrics.failedEvents++;
      await this.scheduleRetry('conversion:error', data);
    }
  }

  /**
   * Process conversion progress updates
   */
  private async processConversionProgress(
    data: ConversionProgressData
  ): Promise<void> {
    if (!data.userId) {
      throw new Error('Missing userId in conversion progress data');
    }

    // Validate progress data
    if (
      typeof data.progress !== 'number' ||
      data.progress < 0 ||
      data.progress > 100
    ) {
      throw new Error('Invalid progress value');
    }

    // Check if user is connected
    if (!this.connectionManager.isUserConnected(data.userId)) {
      console.log(
        `User ${data.userId} not connected, storing progress for later delivery`
      );
      await this.storeProgressForLaterDelivery(data);
      return;
    }

    console.log(
      `Sending progress update to user ${data.userId}: ${data.progress}%`
    );

    const progressEvent = {
      videoId: data.videoId,
      jobId: data.jobId,
      progress: data.progress,
      status: data.status,
      estimatedTime: data.estimatedTime,
      timestamp: new Date().toISOString(),
      eventType: 'progress',
    };

    this.connectionManager.sendToUser(
      data.userId,
      'conversion:progress',
      progressEvent
    );

    // Store latest progress in Redis for reconnecting users
    await this.redisService.publishMessage('progress:store', {
      userId: data.userId,
      data: progressEvent,
      timestamp: Date.now(),
    });
  }

  /**
   * Process conversion completion
   */
  private async processConversionComplete(
    data: ConversionProgressData
  ): Promise<void> {
    if (!data.userId) {
      throw new Error('Missing userId in conversion complete data');
    }

    console.log(`Sending completion notification to user ${data.userId}`);

    const completeEvent = {
      videoId: data.videoId,
      jobId: data.jobId,
      status: 'completed',
      downloadUrl: (data as any).downloadUrl,
      fileSize: (data as any).fileSize,
      duration: (data as any).duration,
      timestamp: new Date().toISOString(),
      eventType: 'complete',
    };

    // Send to user if connected
    if (this.connectionManager.isUserConnected(data.userId)) {
      this.connectionManager.sendToUser(
        data.userId,
        'conversion:complete',
        completeEvent
      );
    }

    // Always store completion event for later retrieval
    await this.storeCompletionEvent(data.userId, completeEvent);

    // Send push notification if user is not connected
    if (!this.connectionManager.isUserConnected(data.userId)) {
      await this.sendPushNotification(
        data.userId,
        'Conversion Complete',
        `Your video conversion is ready for download!`
      );
    }
  }

  /**
   * Process conversion errors
   */
  private async processConversionError(
    data: ConversionProgressData
  ): Promise<void> {
    if (!data.userId) {
      throw new Error('Missing userId in conversion error data');
    }

    console.log(`Sending error notification to user ${data.userId}`);

    const errorEvent = {
      videoId: data.videoId,
      jobId: data.jobId,
      status: 'failed',
      error: data.errorMessage || 'Unknown error occurred',
      timestamp: new Date().toISOString(),
      eventType: 'error',
      retryable: this.isRetryableError(data.errorMessage),
    };

    // Send to user if connected
    if (this.connectionManager.isUserConnected(data.userId)) {
      this.connectionManager.sendToUser(
        data.userId,
        'conversion:error',
        errorEvent
      );
    }

    // Store error event
    await this.storeErrorEvent(data.userId, errorEvent);

    // Send push notification for critical errors
    if (!this.connectionManager.isUserConnected(data.userId)) {
      await this.sendPushNotification(
        data.userId,
        'Conversion Failed',
        `There was an error processing your video: ${errorEvent.error}`
      );
    }
  }

  /**
   * Store progress for later delivery when user reconnects
   */
  private async storeProgressForLaterDelivery(
    data: ConversionProgressData
  ): Promise<void> {
    const key = `user:${data.userId}:pending_progress`;
    const progressData = {
      videoId: data.videoId,
      jobId: data.jobId,
      progress: data.progress,
      status: data.status,
      timestamp: new Date().toISOString(),
    };

    try {
      await this.redisService.publishMessage('store:progress', {
        userId: data.userId,
        data: progressData,
        timestamp: Date.now(),
      });
    } catch (error) {
      console.error('Failed to store progress for later delivery:', error);
    }
  }

  /**
   * Store completion event for user history
   */
  private async storeCompletionEvent(
    userId: string,
    event: any
  ): Promise<void> {
    try {
      await this.redisService.publishMessage('store:completion', {
        userId,
        data: event,
        timestamp: Date.now(),
      });
    } catch (error) {
      console.error('Failed to store completion event:', error);
    }
  }

  /**
   * Store error event for user history
   */
  private async storeErrorEvent(userId: string, event: any): Promise<void> {
    try {
      await this.redisService.publishMessage('store:error', {
        userId,
        data: event,
        timestamp: Date.now(),
      });
    } catch (error) {
      console.error('Failed to store error event:', error);
    }
  }

  /**
   * Send push notification (placeholder for actual implementation)
   */
  private async sendPushNotification(
    userId: string,
    title: string,
    message: string
  ): Promise<void> {
    try {
      await this.redisService.publishMessage('notification:push', {
        userId,
        data: { title, message, type: 'conversion' },
        timestamp: Date.now(),
      });
    } catch (error) {
      console.error('Failed to send push notification:', error);
    }
  }

  /**
   * Check if error is retryable
   */
  private isRetryableError(errorMessage?: string): boolean {
    if (!errorMessage) return false;

    const retryableErrors = [
      'network timeout',
      'connection refused',
      'temporary failure',
      'service unavailable',
    ];

    return retryableErrors.some(error =>
      errorMessage.toLowerCase().includes(error)
    );
  }

  /**
   * Schedule event for retry
   */
  private async scheduleRetry(eventType: string, data: any): Promise<void> {
    const retryId = `${eventType}:${data.userId}:${data.jobId || Date.now()}`;
    const existingRetry = this.retryQueue.get(retryId);

    if (
      existingRetry &&
      existingRetry.attempts >= this.retryConfig.maxRetries
    ) {
      console.error(`Max retries exceeded for event ${retryId}`);
      return;
    }

    const attempts = existingRetry ? existingRetry.attempts + 1 : 1;
    const delay =
      this.retryConfig.retryDelay *
      Math.pow(this.retryConfig.backoffMultiplier, attempts - 1);
    const nextRetry = Date.now() + delay;

    this.retryQueue.set(retryId, {
      data: { eventType, ...data },
      attempts,
      nextRetry,
    });

    this.metrics.retryAttempts++;
    console.log(
      `Scheduled retry ${attempts}/${this.retryConfig.maxRetries} for ${retryId} in ${delay}ms`
    );
  }

  /**
   * Process retry queue
   */
  private startRetryProcessor(): void {
    this.retryInterval = setInterval(async () => {
      const now = Date.now();
      const retryPromises: Promise<void>[] = [];

      for (const [retryId, retryData] of this.retryQueue.entries()) {
        if (retryData.nextRetry <= now) {
          retryPromises.push(this.processRetry(retryId, retryData));
        }
      }

      if (retryPromises.length > 0) {
        await Promise.allSettled(retryPromises);
      }
    }, 1000); // Check every second
  }

  /**
   * Process individual retry
   */
  private async processRetry(retryId: string, retryData: any): Promise<void> {
    try {
      const { eventType, ...data } = retryData.data;

      switch (eventType) {
        case 'conversion:progress':
          await this.processConversionProgress(data);
          break;
        case 'conversion:complete':
          await this.processConversionComplete(data);
          break;
        case 'conversion:error':
          await this.processConversionError(data);
          break;
      }

      // Remove from retry queue on success
      this.retryQueue.delete(retryId);
      console.log(`Retry successful for ${retryId}`);
    } catch (error) {
      console.error(`Retry failed for ${retryId}:`, error);

      if (retryData.attempts >= this.retryConfig.maxRetries) {
        this.retryQueue.delete(retryId);
        console.error(
          `Giving up on ${retryId} after ${retryData.attempts} attempts`
        );
      } else {
        // Reschedule with exponential backoff
        await this.scheduleRetry(retryData.data.eventType, retryData.data);
      }
    }
  }

  /**
   * Get event handling metrics
   */
  public getMetrics(): EventMetrics {
    return { ...this.metrics };
  }

  /**
   * Reset metrics
   */
  public resetMetrics(): void {
    this.metrics = {
      totalEvents: 0,
      successfulEvents: 0,
      failedEvents: 0,
      retryAttempts: 0,
    };
  }

  /**
   * Get retry queue status
   */
  public getRetryQueueStatus(): {
    queueSize: number;
    pendingRetries: number;
  } {
    const now = Date.now();
    const pendingRetries = Array.from(this.retryQueue.values()).filter(
      retry => retry.nextRetry <= now
    ).length;

    return {
      queueSize: this.retryQueue.size,
      pendingRetries,
    };
  }

  /**
   * Cleanup retry processor
   */
  public cleanup(): void {
    if (this.retryInterval) {
      clearInterval(this.retryInterval);
    }
    this.retryQueue.clear();
  }
}
