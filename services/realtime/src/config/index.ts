import dotenv from 'dotenv';

dotenv.config();

export const config = {
  port: parseInt(process.env.PORT || '3001', 10),
  jwtSecret: process.env.JWT_SECRET || 'your-secret-key',
  redisUrl: process.env.REDIS_URL || 'redis://localhost:6379',
  corsOrigin: process.env.CORS_ORIGIN || '*',
  nodeEnv: process.env.NODE_ENV || 'development',

  // Connection settings
  connectionTimeout: 30000, // 30 seconds
  heartbeatInterval: 25000, // 25 seconds
  maxConnections: 10000,

  // Redis channels
  redisChannels: {
    conversionProgress: 'conversion:progress',
    conversionComplete: 'conversion:complete',
    conversionError: 'conversion:error',
    notification: 'notification:send',
    analytics: 'analytics:update',
  },
};
