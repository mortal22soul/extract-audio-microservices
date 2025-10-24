import { create } from 'zustand';
import { Video } from '@/types';
import apiClient from '@/lib/api';
import toast from 'react-hot-toast';

interface VideoState {
  videos: Video[];
  currentVideo: Video | null;
  isLoading: boolean;
  isUploading: boolean;
  uploadProgress: number;

  // Actions
  fetchVideos: () => Promise<void>;
  uploadVideo: (file: File) => Promise<Video | null>;
  getVideo: (id: string) => Promise<void>;
  deleteVideo: (id: string) => Promise<void>;
  downloadVideo: (id: string, filename: string) => Promise<void>;
  setCurrentVideo: (video: Video | null) => void;
  setUploadProgress: (progress: number) => void;
  updateVideoStatus: (videoId: string, status: Video['status']) => void;
}

export const useVideoStore = create<VideoState>((set, get) => ({
  videos: [],
  currentVideo: null,
  isLoading: false,
  isUploading: false,
  uploadProgress: 0,

  fetchVideos: async () => {
    try {
      set({ isLoading: true });

      const videos = await apiClient.getVideos();

      set({
        videos,
        isLoading: false,
      });
    } catch (error: any) {
      set({ isLoading: false });
      console.error('Failed to fetch videos:', error);
      toast.error('Failed to load videos');
    }
  },

  uploadVideo: async (file: File) => {
    try {
      set({ isUploading: true, uploadProgress: 0 });

      const video = await apiClient.uploadVideo(file, progress => {
        set({ uploadProgress: progress });
      });

      // Add the new video to the list
      const { videos } = get();
      set({
        videos: [video, ...videos],
        isUploading: false,
        uploadProgress: 0,
      });

      toast.success('Video uploaded successfully!');
      return video;
    } catch (error: any) {
      set({ isUploading: false, uploadProgress: 0 });
      const message = error.response?.data?.error || 'Upload failed';
      toast.error(message);
      return null;
    }
  },

  getVideo: async (id: string) => {
    try {
      set({ isLoading: true });

      const video = await apiClient.getVideo(id);

      set({
        currentVideo: video,
        isLoading: false,
      });
    } catch (error: any) {
      set({ isLoading: false });
      console.error('Failed to fetch video:', error);
      toast.error('Failed to load video');
    }
  },

  deleteVideo: async (id: string) => {
    try {
      await apiClient.deleteVideo(id);

      // Remove the video from the list
      const { videos } = get();
      set({
        videos: videos.filter(video => video.id !== id),
      });

      toast.success('Video deleted successfully!');
    } catch (error: any) {
      console.error('Failed to delete video:', error);
      toast.error('Failed to delete video');
    }
  },

  downloadVideo: async (id: string, filename: string) => {
    try {
      const blob = await apiClient.downloadVideo(id);

      // Create download link
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = filename;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      window.URL.revokeObjectURL(url);

      toast.success('Download started!');
    } catch (error: any) {
      console.error('Failed to download video:', error);
      toast.error('Failed to download video');
    }
  },

  setCurrentVideo: (video: Video | null) => {
    set({ currentVideo: video });
  },

  setUploadProgress: (progress: number) => {
    set({ uploadProgress: progress });
  },

  updateVideoStatus: (videoId: string, status: Video['status']) => {
    const { videos } = get();
    const updatedVideos = videos.map(video =>
      video.id === videoId ? { ...video, status } : video
    );
    set({ videos: updatedVideos });
  },
}));
