'use client';

import { useCallback, useState } from 'react';
import { useDropzone } from 'react-dropzone';
import { useVideoStore } from '@/stores/videoStore';
import { isValidVideoFile, isFileSizeValid, formatFileSize } from '@/lib/utils';

interface VideoUploadProps {
  onUploadComplete?: (videoId: string) => void;
  className?: string;
}

export default function VideoUpload({
  onUploadComplete,
  className = '',
}: VideoUploadProps) {
  const { uploadVideo, isUploading, uploadProgress } = useVideoStore();
  const [dragActive, setDragActive] = useState(false);

  const onDrop = useCallback(
    async (acceptedFiles: File[], rejectedFiles: any[]) => {
      // Handle rejected files
      if (rejectedFiles.length > 0) {
        const reasons = rejectedFiles
          .map(file => file.errors[0]?.message)
          .join(', ');
        console.error('File rejected:', reasons);
        return;
      }

      const file = acceptedFiles[0];
      if (!file) return;

      // Validate file type
      if (!isValidVideoFile(file)) {
        console.error('Invalid file type. Please upload a video file.');
        return;
      }

      // Validate file size
      if (!isFileSizeValid(file)) {
        console.error(
          `File too large. Maximum size is ${formatFileSize(parseInt(process.env.NEXT_PUBLIC_MAX_FILE_SIZE || '100000000'))}`
        );
        return;
      }

      // Upload the video
      const video = await uploadVideo(file);
      if (video && onUploadComplete) {
        onUploadComplete(video.id);
      }
    },
    [uploadVideo, onUploadComplete]
  );

  const { getRootProps, getInputProps, isDragActive } = useDropzone({
    onDrop,
    accept: {
      'video/*': ['.mp4', '.avi', '.mov', '.mkv', '.webm'],
    },
    maxFiles: 1,
    disabled: isUploading,
    onDragEnter: () => setDragActive(true),
    onDragLeave: () => setDragActive(false),
  });

  return (
    <div className={`w-full ${className}`}>
      <div
        {...getRootProps()}
        className={`
          relative border-2 border-dashed rounded-lg p-8 text-center cursor-pointer transition-all duration-200
          ${
            isDragActive || dragActive
              ? 'border-blue-500 bg-blue-50'
              : 'border-gray-300 hover:border-gray-400'
          }
          ${isUploading ? 'cursor-not-allowed opacity-75' : 'hover:bg-gray-50'}
        `}
      >
        <input {...getInputProps()} />

        {isUploading ? (
          <div className="space-y-4">
            <div className="mx-auto w-16 h-16 bg-blue-100 rounded-full flex items-center justify-center">
              <svg
                className="w-8 h-8 text-blue-600 animate-pulse"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M9 19l3 3m0 0l3-3m-3 3V10"
                />
              </svg>
            </div>

            <div>
              <p className="text-lg font-medium text-gray-900">
                Uploading video...
              </p>
              <p className="text-sm text-gray-600">
                Please don't close this page
              </p>
            </div>

            <div className="max-w-xs mx-auto">
              <div className="progress-bar">
                <div
                  className="progress-fill"
                  style={{ width: `${uploadProgress}%` }}
                />
              </div>
              <p className="text-sm text-gray-600 mt-2">
                {uploadProgress}% complete
              </p>
            </div>
          </div>
        ) : (
          <div className="space-y-4">
            <div className="mx-auto w-16 h-16 bg-gray-100 rounded-full flex items-center justify-center">
              <svg
                className="w-8 h-8 text-gray-600"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M9 19l3 3m0 0l3-3m-3 3V10"
                />
              </svg>
            </div>

            <div>
              <p className="text-lg font-medium text-gray-900">
                {isDragActive ? 'Drop your video here' : 'Upload a video file'}
              </p>
              <p className="text-sm text-gray-600">
                Drag & drop or click to select • MP4, AVI, MOV, MKV, WebM
              </p>
              <p className="text-xs text-gray-500 mt-1">
                Maximum file size:{' '}
                {formatFileSize(
                  parseInt(process.env.NEXT_PUBLIC_MAX_FILE_SIZE || '100000000')
                )}
              </p>
            </div>

            <button
              type="button"
              className="btn-primary"
              disabled={isUploading}
            >
              Choose File
            </button>
          </div>
        )}
      </div>
    </div>
  );
}
