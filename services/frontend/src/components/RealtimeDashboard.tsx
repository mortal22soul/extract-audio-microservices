'use client';

import { useEffect, useState } from 'react';
import { useSocket } from '@/hooks/useSocket';
import { useVideoStore } from '@/stores/videoStore';

interface SystemStatus {
  activeConnections: number;
  processingJobs: number;
  queueLength: number;
  systemLoad: number;
  uptime: number;
}

export default function RealtimeDashboard() {
  const { isConnected, connectionError, socket } = useSocket();
  const { videos } = useVideoStore();
  const [systemStatus, setSystemStatus] = useState<SystemStatus | null>(null);

  useEffect(() => {
    if (socket) {
      socket.on('system_status', (status: SystemStatus) => {
        setSystemStatus(status);
      });

      // Request initial system status
      socket.emit('get_system_status');

      return () => {
        socket.off('system_status');
      };
    }

    return undefined;
  }, [socket]);

  const processingVideos = videos.filter(
    video => video.status === 'processing'
  );
  const completedVideos = videos.filter(video => video.status === 'completed');
  const failedVideos = videos.filter(video => video.status === 'failed');

  const formatUptime = (seconds: number) => {
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    return `${hours}h ${minutes}m`;
  };

  return (
    <div className="bg-white rounded-lg shadow p-6">
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-xl font-semibold text-gray-900">
          Real-time Status
        </h2>

        {/* Connection Status */}
        <div className="flex items-center space-x-2">
          <div
            className={`w-3 h-3 rounded-full ${
              isConnected ? 'bg-green-500' : 'bg-red-500'
            }`}
          ></div>
          <span
            className={`text-sm font-medium ${
              isConnected ? 'text-green-700' : 'text-red-700'
            }`}
          >
            {isConnected ? 'Connected' : 'Disconnected'}
          </span>
        </div>
      </div>

      {connectionError && (
        <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded-md">
          <p className="text-sm text-red-700">
            Connection Error: {connectionError}
          </p>
        </div>
      )}

      {/* User Statistics */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
        <div className="bg-blue-50 rounded-lg p-4">
          <div className="flex items-center">
            <div className="flex-shrink-0">
              <svg
                className="w-6 h-6 text-blue-600"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
                />
              </svg>
            </div>
            <div className="ml-3">
              <p className="text-sm font-medium text-blue-900">Processing</p>
              <p className="text-lg font-semibold text-blue-600">
                {processingVideos.length}
              </p>
            </div>
          </div>
        </div>

        <div className="bg-green-50 rounded-lg p-4">
          <div className="flex items-center">
            <div className="flex-shrink-0">
              <svg
                className="w-6 h-6 text-green-600"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M5 13l4 4L19 7"
                />
              </svg>
            </div>
            <div className="ml-3">
              <p className="text-sm font-medium text-green-900">Completed</p>
              <p className="text-lg font-semibold text-green-600">
                {completedVideos.length}
              </p>
            </div>
          </div>
        </div>

        <div className="bg-red-50 rounded-lg p-4">
          <div className="flex items-center">
            <div className="flex-shrink-0">
              <svg
                className="w-6 h-6 text-red-600"
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
            </div>
            <div className="ml-3">
              <p className="text-sm font-medium text-red-900">Failed</p>
              <p className="text-lg font-semibold text-red-600">
                {failedVideos.length}
              </p>
            </div>
          </div>
        </div>
      </div>

      {/* System Status */}
      {systemStatus && (
        <div>
          <h3 className="text-lg font-medium text-gray-900 mb-4">
            System Status
          </h3>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            <div className="bg-gray-50 rounded-lg p-4">
              <div className="text-center">
                <p className="text-2xl font-bold text-gray-900">
                  {systemStatus.activeConnections}
                </p>
                <p className="text-sm text-gray-600">Active Connections</p>
              </div>
            </div>

            <div className="bg-gray-50 rounded-lg p-4">
              <div className="text-center">
                <p className="text-2xl font-bold text-gray-900">
                  {systemStatus.processingJobs}
                </p>
                <p className="text-sm text-gray-600">Processing Jobs</p>
              </div>
            </div>

            <div className="bg-gray-50 rounded-lg p-4">
              <div className="text-center">
                <p className="text-2xl font-bold text-gray-900">
                  {systemStatus.queueLength}
                </p>
                <p className="text-sm text-gray-600">Queue Length</p>
              </div>
            </div>

            <div className="bg-gray-50 rounded-lg p-4">
              <div className="text-center">
                <p className="text-2xl font-bold text-gray-900">
                  {formatUptime(systemStatus.uptime)}
                </p>
                <p className="text-sm text-gray-600">System Uptime</p>
              </div>
            </div>
          </div>

          {/* System Load Bar */}
          <div className="mt-4">
            <div className="flex items-center justify-between mb-2">
              <span className="text-sm font-medium text-gray-700">
                System Load
              </span>
              <span className="text-sm text-gray-600">
                {systemStatus.systemLoad}%
              </span>
            </div>
            <div className="progress-bar">
              <div
                className={`h-2 rounded-full transition-all duration-300 ${
                  systemStatus.systemLoad > 80
                    ? 'bg-red-600'
                    : systemStatus.systemLoad > 60
                      ? 'bg-yellow-600'
                      : 'bg-green-600'
                }`}
                style={{ width: `${systemStatus.systemLoad}%` }}
              />
            </div>
          </div>
        </div>
      )}

      {/* Processing Videos */}
      {processingVideos.length > 0 && (
        <div className="mt-6">
          <h3 className="text-lg font-medium text-gray-900 mb-4">
            Currently Processing
          </h3>
          <div className="space-y-3">
            {processingVideos.map(video => (
              <div key={video.id} className="bg-blue-50 rounded-lg p-4">
                <div className="flex items-center justify-between">
                  <div className="flex-1 min-w-0">
                    <p className="text-sm font-medium text-blue-900 truncate">
                      {video.originalFilename}
                    </p>
                    <p className="text-xs text-blue-700">
                      Uploaded {new Date(video.uploadedAt).toLocaleTimeString()}
                    </p>
                  </div>
                  <div className="flex items-center space-x-2">
                    <div className="w-4 h-4 bg-blue-600 rounded-full animate-pulse"></div>
                    <span className="text-sm font-medium text-blue-600">
                      Processing...
                    </span>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
