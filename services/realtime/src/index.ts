import { RealtimeService } from './services/RealtimeService';

// Create and start the realtime service
const realtimeService = new RealtimeService();

// Graceful shutdown handling
process.on('SIGTERM', async () => {
  console.log('SIGTERM received, shutting down gracefully...');
  await realtimeService.stop();
  process.exit(0);
});

process.on('SIGINT', async () => {
  console.log('SIGINT received, shutting down gracefully...');
  await realtimeService.stop();
  process.exit(0);
});

// Handle uncaught exceptions
process.on('uncaughtException', error => {
  console.error('Uncaught Exception:', error);
  process.exit(1);
});

process.on('unhandledRejection', (reason, promise) => {
  console.error('Unhandled Rejection at:', promise, 'reason:', reason);
  process.exit(1);
});

// Start the service
realtimeService.start().catch(error => {
  console.error('Failed to start realtime service:', error);
  process.exit(1);
});
