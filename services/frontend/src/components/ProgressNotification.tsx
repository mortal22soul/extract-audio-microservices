'use client';

import { useEffect, useState } from 'react';
import { useSocket } from '@/hooks/useSocket';
import { ProgressEvent } from '@/types';

interface ProgressNotificationProps {
  videoId?: string;
  className?: string;
}

interface ActiveProgress {
  [videoId: string]: {
    progress: number;
    filename: string;
    estimatedTime?: number | undefined;
  };
}

export default function ProgressNotification({
  videoId,
  className = '',
}: ProgressNotificationProps) {
  const { socket, isConnected } = useSocket();
  const [activeProgress, setActiveProgress] = useState<ActiveProgress>({});

  useEffect(() => {
    if (!socket) return;

    const handleProgress = (data: ProgressEvent) => {
      // Only show progress for specific video if videoId is provided
      if (videoId && data.videoId !== videoId) return;

      setActiveProgress(prev => ({
        ...prev,
        [data.videoId]: {
          progress: data.progress,
          filename: data.videoId, // This would ideally come from the video store
          estimatedTime: data.estimatedTime,
        },
      }));
    };

    const handleComplete = (data: any) => {
      // Remove from active progress
      setActiveProgress(prev => {
        const updated = { ...prev };
        delete updated[data.videoId];
        return updated;
      });
    };

    const handleError = (data: any) => {
      // Remove from active progress
      setActiveProgress(prev => {
        const updated = { ...prev };
        delete updated[data.videoId];
        return updated;
      });
    };

    socket.on('progress', handleProgress);
    socket.on('complete', handleComplete);
    socket.on('error', handleError);

    return () => {
      socket.off('progress', handleProgress);
      socket.off('complete', handleComplete);
      socket.off('error', handleError);
    };
  }, [socket, videoId]);

  const progressEntries = Object.entries(activeProgress);

  if (!isConnected || progressEntries.length === 0) {
    return null;
  }

  return (
    <div className={`fixed bottom-4 right-4 space-y-2 z-50 ${className}`}>
      {progressEntries.map(([id, progress]) => (
        <div
          key={id}
          className="bg-white rounded-lg shadow-lg border border-gray-200 p-4 min-w-80 max-w-sm animate-slide-up"
        >
          <div className="flex items-center justify-between mb-2">
            <div className="flex items-center space-x-2">
              <div className="w-3 h-3 bg-blue-600 rounded-full animate-pulse"></div>
              <span className="text-sm font-medium text-gray-900">
                Converting...
              </span>
            </div>
            <button
              onClick={() => {
                setActiveProgress(prev => {
                  const updated = { ...prev };
                  delete updated[id];
                  return updated;
                });
              }}
              className="text-gray-400 hover:text-gray-600"
            >
              <svg
                className="w-4 h-4"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M6 18L18 6M6 6l12 12"
                />
              </svg>
            </button>
          </div>

          <div className="mb-2">
            <p
              className="text-sm text-gray-600 truncate"
              title={progress.filename}
            >
              {progress.filename}
            </p>
          </div>

          <div className="mb-2">
            <div className="flex items-center justify-between text-xs text-gray-500 mb-1">
              <span>{progress.progress}% complete</span>
              {progress.estimatedTime && (
                <span>
                  ~{Math.ceil(progress.estimatedTime / 60)}m remaining
                </span>
              )}
            </div>
            <div className="progress-bar">
              <div
                className="progress-fill"
                style={{ width: `${progress.progress}%` }}
              />
            </div>
          </div>

          <div className="flex items-center space-x-2 text-xs text-gray-500">
            <svg
              className="w-4 h-4"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
              />
            </svg>
            <span>Keep this tab open during conversion</span>
          </div>
        </div>
      ))}
    </div>
  );
}
