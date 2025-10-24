'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import Layout from '@/components/Layout';
import ProtectedRoute from '@/components/ProtectedRoute';
import VideoUpload from '@/components/VideoUpload';
import VideoList from '@/components/VideoList';
import ProgressNotification from '@/components/ProgressNotification';

export default function UploadPage() {
  const router = useRouter();
  const [showRecentUploads, setShowRecentUploads] = useState(true);

  const handleUploadComplete = (videoId: string) => {
    // Optionally navigate to the video details or dashboard
    console.log('Upload completed for video:', videoId);

    // Refresh the recent uploads
    setShowRecentUploads(false);
    setTimeout(() => setShowRecentUploads(true), 100);
  };

  return (
    <ProtectedRoute>
      <Layout>
        <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          {/* Header */}
          <div className="mb-8">
            <h1 className="text-3xl font-bold text-gray-900">Upload Video</h1>
            <p className="text-gray-600 mt-2">
              Upload your video files and convert them to MP3 format with
              real-time progress tracking.
            </p>
          </div>

          {/* Upload Section */}
          <div className="mb-12">
            <VideoUpload
              onUploadComplete={handleUploadComplete}
              className="max-w-2xl mx-auto"
            />
          </div>

          {/* Instructions */}
          <div className="bg-blue-50 border border-blue-200 rounded-lg p-6 mb-8">
            <h2 className="text-lg font-semibold text-blue-900 mb-3">
              How it works
            </h2>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4 text-sm text-blue-800">
              <div className="flex items-start space-x-3">
                <div className="w-6 h-6 bg-blue-600 text-white rounded-full flex items-center justify-center text-xs font-bold flex-shrink-0 mt-0.5">
                  1
                </div>
                <div>
                  <h3 className="font-medium">Upload Video</h3>
                  <p className="text-blue-700">
                    Select or drag & drop your video file. We support MP4, AVI,
                    MOV, MKV, and WebM formats.
                  </p>
                </div>
              </div>

              <div className="flex items-start space-x-3">
                <div className="w-6 h-6 bg-blue-600 text-white rounded-full flex items-center justify-center text-xs font-bold flex-shrink-0 mt-0.5">
                  2
                </div>
                <div>
                  <h3 className="font-medium">Processing</h3>
                  <p className="text-blue-700">
                    Our AI analyzes your video and converts it to high-quality
                    MP3. Track progress in real-time.
                  </p>
                </div>
              </div>

              <div className="flex items-start space-x-3">
                <div className="w-6 h-6 bg-blue-600 text-white rounded-full flex items-center justify-center text-xs font-bold flex-shrink-0 mt-0.5">
                  3
                </div>
                <div>
                  <h3 className="font-medium">Download</h3>
                  <p className="text-blue-700">
                    Once complete, download your MP3 file and enjoy your
                    converted audio.
                  </p>
                </div>
              </div>
            </div>
          </div>

          {/* Recent Uploads */}
          {showRecentUploads && (
            <div>
              <div className="flex items-center justify-between mb-6">
                <h2 className="text-xl font-semibold text-gray-900">
                  Recent Uploads
                </h2>
                <button
                  onClick={() => router.push('/history')}
                  className="text-blue-600 hover:text-blue-700 text-sm font-medium"
                >
                  View all →
                </button>
              </div>

              <VideoList limit={5} />
            </div>
          )}

          {/* Tips */}
          <div className="mt-12 bg-gray-50 rounded-lg p-6">
            <h2 className="text-lg font-semibold text-gray-900 mb-4">
              Tips for best results
            </h2>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm text-gray-700">
              <div className="flex items-start space-x-2">
                <svg
                  className="w-5 h-5 text-green-600 flex-shrink-0 mt-0.5"
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
                <span>
                  Use high-quality source videos for better audio output
                </span>
              </div>
              <div className="flex items-start space-x-2">
                <svg
                  className="w-5 h-5 text-green-600 flex-shrink-0 mt-0.5"
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
                <span>Ensure stable internet connection during upload</span>
              </div>
              <div className="flex items-start space-x-2">
                <svg
                  className="w-5 h-5 text-green-600 flex-shrink-0 mt-0.5"
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
                <span>Keep the browser tab open during processing</span>
              </div>
              <div className="flex items-start space-x-2">
                <svg
                  className="w-5 h-5 text-green-600 flex-shrink-0 mt-0.5"
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
                <span>Check file format compatibility before upload</span>
              </div>
            </div>
          </div>

          {/* Progress Notifications */}
          <ProgressNotification />
        </div>
      </Layout>
    </ProtectedRoute>
  );
}
