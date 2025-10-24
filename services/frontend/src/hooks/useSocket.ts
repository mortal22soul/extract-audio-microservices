'use client';

import { useEffect, useState, useRef } from 'react';
import { io, Socket } from 'socket.io-client';
import { useAuthStore } from '@/stores/authStore';
import { useVideoStore } from '@/stores/videoStore';
import { ProgressEvent, CompletionEvent, ErrorEvent } from '@/types';
import toast from 'react-hot-toast';

interface UseSocketOptions {
  autoConnect?: boolean;
  onConnect?: () => void;
  onDisconnect?: () => void;
  onError?: (error: any) => void;
}

export const useSocket = (
  options: UseSocketOptions = {}
): {
  socket: Socket | null;
  isConnected: boolean;
  connectionError: string | null;
  connect: () => void;
  disconnect: () => void;
  emit: (event: string, data?: any) => void;
  on: (event: string, callback: (...args: any[]) => void) => void;
  off: (event: string, callback?: (...args: any[]) => void) => void;
} => {
  const { autoConnect = true, onConnect, onDisconnect, onError } = options;
  const [socket, setSocket] = useState<Socket | null>(null);
  const [isConnected, setIsConnected] = useState(false);
  const [connectionError, setConnectionError] = useState<string | null>(null);
  const { isAuthenticated, getAuthToken } = useAuthStore();
  const { updateVideoStatus } = useVideoStore();
  const reconnectTimeoutRef = useRef<NodeJS.Timeout>();
  const reconnectAttemptsRef = useRef(0);
  const maxReconnectAttempts = 5;

  const connect = () => {
    if (!isAuthenticated) {
      console.log('Not authenticated, skipping socket connection');
      return;
    }

    const token = getAuthToken();
    if (!token) {
      console.log('No auth token available');
      return;
    }

    console.log('Connecting to realtime service...');

    const newSocket = io(
      process.env.NEXT_PUBLIC_REALTIME_URL || 'http://localhost:3001',
      {
        auth: { token },
        transports: ['websocket', 'polling'],
        timeout: 10000,
        reconnection: true,
        reconnectionAttempts: maxReconnectAttempts,
        reconnectionDelay: 1000,
        reconnectionDelayMax: 5000,
      }
    );

    // Connection events
    newSocket.on('connect', () => {
      console.log('Connected to realtime service');
      setIsConnected(true);
      setConnectionError(null);
      reconnectAttemptsRef.current = 0;
      onConnect?.();
    });

    newSocket.on('disconnect', reason => {
      console.log('Disconnected from realtime service:', reason);
      setIsConnected(false);
      onDisconnect?.();

      // Handle reconnection for certain disconnect reasons
      if (reason === 'io server disconnect') {
        // Server initiated disconnect, don't reconnect automatically
        setConnectionError('Server disconnected');
      } else {
        // Client or network issue, attempt reconnection
        attemptReconnect();
      }
    });

    newSocket.on('connect_error', error => {
      console.error('Socket connection error:', error);
      setConnectionError(error.message);
      setIsConnected(false);
      onError?.(error);
      attemptReconnect();
    });

    // Video conversion events
    newSocket.on('progress', (data: ProgressEvent) => {
      console.log('Conversion progress:', data);

      // Update video status in store
      updateVideoStatus(data.videoId, 'processing');

      // Show progress notification
      toast.loading(`Converting video... ${data.progress}%`, {
        id: `progress-${data.videoId}`,
        duration: 1000,
      });
    });

    newSocket.on('complete', (data: CompletionEvent) => {
      console.log('Conversion completed:', data);

      // Update video status in store
      updateVideoStatus(data.videoId, 'completed');

      // Dismiss progress toast and show success
      toast.dismiss(`progress-${data.videoId}`);
      toast.success(
        'Video conversion completed! You can now download your MP3.',
        {
          duration: 5000,
          icon: '🎵',
        }
      );
    });

    newSocket.on('error', (data: ErrorEvent) => {
      console.error('Conversion error:', data);

      // Update video status in store
      updateVideoStatus(data.videoId, 'failed');

      // Dismiss progress toast and show error
      toast.dismiss(`progress-${data.videoId}`);
      toast.error(`Conversion failed: ${data.error}`, {
        duration: 8000,
      });
    });

    // System status events
    newSocket.on('system_status', data => {
      console.log('System status update:', data);
      // Handle system status updates if needed
    });

    setSocket(newSocket);
  };

  const disconnect = () => {
    if (socket) {
      console.log('Disconnecting from realtime service...');
      socket.disconnect();
      setSocket(null);
      setIsConnected(false);
    }

    // Clear reconnection timeout
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
    }
  };

  const attemptReconnect = () => {
    if (reconnectAttemptsRef.current >= maxReconnectAttempts) {
      console.log('Max reconnection attempts reached');
      setConnectionError('Failed to reconnect after multiple attempts');
      return;
    }

    const delay = Math.min(
      1000 * Math.pow(2, reconnectAttemptsRef.current),
      30000
    );
    console.log(
      `Attempting to reconnect in ${delay}ms (attempt ${reconnectAttemptsRef.current + 1})`
    );

    reconnectTimeoutRef.current = setTimeout(() => {
      reconnectAttemptsRef.current++;
      connect();
    }, delay);
  };

  const emit = (event: string, data?: any) => {
    if (socket && isConnected) {
      socket.emit(event, data);
    } else {
      console.warn('Socket not connected, cannot emit event:', event);
    }
  };

  const on = (event: string, callback: (...args: any[]) => void) => {
    if (socket) {
      socket.on(event, callback);
    }
  };

  const off = (event: string, callback?: (...args: any[]) => void) => {
    if (socket) {
      socket.off(event, callback);
    }
  };

  // Auto-connect when authenticated
  useEffect(() => {
    if (autoConnect && isAuthenticated) {
      connect();
    } else if (!isAuthenticated) {
      disconnect();
    }

    return () => {
      disconnect();
    };
  }, [isAuthenticated, autoConnect]);

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      disconnect();
    };
  }, []);

  return {
    socket,
    isConnected,
    connectionError,
    connect,
    disconnect,
    emit,
    on,
    off,
  };
};
