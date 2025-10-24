// MongoDB seed data for development and testing
// This script creates sample data for testing the video converter system

// Switch to the application database
db = db.getSiblingDB("video_converter");

// Sample user IDs (these should match PostgreSQL user IDs in a real system)
const sampleUserIds = [
  ObjectId("507f1f77bcf86cd799439011"),
  ObjectId("507f1f77bcf86cd799439012"),
  ObjectId("507f1f77bcf86cd799439013"),
];

// Sample video documents
const sampleVideos = [
  {
    _id: ObjectId("507f1f77bcf86cd799439021"),
    userId: sampleUserIds[0],
    originalFilename: "sample_video_1.mp4",
    mimeType: "video/mp4",
    size: NumberLong(15728640), // 15MB
    status: "completed",
    uploadedAt: new Date("2024-01-15T10:30:00Z"),
    conversionJobId: "job_001",
    gridfsFileId: ObjectId("507f1f77bcf86cd799439031"),
    mp3FileId: ObjectId("507f1f77bcf86cd799439041"),
    metadata: {
      duration: 120.5,
      resolution: "1920x1080",
      codec: "h264",
      bitrate: NumberLong(2500000),
      fps: 30.0,
    },
    analytics: {
      thumbnails: [
        "/thumbnails/507f1f77bcf86cd799439021_1.jpg",
        "/thumbnails/507f1f77bcf86cd799439021_2.jpg",
        "/thumbnails/507f1f77bcf86cd799439021_3.jpg",
      ],
      qualityScore: 85.5,
      safetyScore: 95.0,
      tags: ["music", "entertainment", "high-quality"],
    },
  },
  {
    _id: ObjectId("507f1f77bcf86cd799439022"),
    userId: sampleUserIds[1],
    originalFilename: "presentation_recording.mov",
    mimeType: "video/quicktime",
    size: NumberLong(52428800), // 50MB
    status: "processing",
    uploadedAt: new Date("2024-01-16T14:20:00Z"),
    conversionJobId: "job_002",
    gridfsFileId: ObjectId("507f1f77bcf86cd799439032"),
    metadata: {
      duration: 1800.0,
      resolution: "1280x720",
      codec: "h264",
      bitrate: NumberLong(1500000),
      fps: 24.0,
    },
  },
  {
    _id: ObjectId("507f1f77bcf86cd799439023"),
    userId: sampleUserIds[2],
    originalFilename: "tutorial_video.avi",
    mimeType: "video/x-msvideo",
    size: NumberLong(104857600), // 100MB
    status: "failed",
    uploadedAt: new Date("2024-01-17T09:15:00Z"),
    conversionJobId: "job_003",
    gridfsFileId: ObjectId("507f1f77bcf86cd799439033"),
    metadata: {
      duration: 3600.0,
      resolution: "854x480",
      codec: "xvid",
      bitrate: NumberLong(800000),
      fps: 25.0,
    },
  },
];

// Sample conversion jobs
const sampleJobs = [
  {
    _id: ObjectId("507f1f77bcf86cd799439051"),
    videoId: ObjectId("507f1f77bcf86cd799439021"),
    userId: sampleUserIds[0],
    status: "completed",
    progress: 100,
    startedAt: new Date("2024-01-15T10:31:00Z"),
    completedAt: new Date("2024-01-15T10:33:30Z"),
    processingNode: "converter-node-1",
    conversionSettings: {
      outputFormat: "mp3",
      bitrate: "192k",
      sampleRate: 44100,
    },
  },
  {
    _id: ObjectId("507f1f77bcf86cd799439052"),
    videoId: ObjectId("507f1f77bcf86cd799439022"),
    userId: sampleUserIds[1],
    status: "processing",
    progress: 65,
    startedAt: new Date("2024-01-16T14:21:00Z"),
    processingNode: "converter-node-2",
    conversionSettings: {
      outputFormat: "mp3",
      bitrate: "128k",
      sampleRate: 44100,
    },
  },
  {
    _id: ObjectId("507f1f77bcf86cd799439053"),
    videoId: ObjectId("507f1f77bcf86cd799439023"),
    userId: sampleUserIds[2],
    status: "failed",
    progress: 25,
    startedAt: new Date("2024-01-17T09:16:00Z"),
    completedAt: new Date("2024-01-17T09:18:45Z"),
    errorMessage: "Unsupported codec: xvid",
    processingNode: "converter-node-1",
    conversionSettings: {
      outputFormat: "mp3",
      bitrate: "320k",
      sampleRate: 48000,
    },
  },
];

// Sample analytics data
const sampleAnalytics = [
  {
    _id: ObjectId("507f1f77bcf86cd799439061"),
    videoId: ObjectId("507f1f77bcf86cd799439021"),
    analysisType: "quality",
    results: {
      sharpness: 0.85,
      brightness: 0.72,
      contrast: 0.68,
      overallScore: 85.5,
      recommendations: ["Good quality video", "Suitable for conversion"],
    },
    analyzedAt: new Date("2024-01-15T10:32:00Z"),
    modelVersion: "quality-analyzer-v1.2",
  },
  {
    _id: ObjectId("507f1f77bcf86cd799439062"),
    videoId: ObjectId("507f1f77bcf86cd799439021"),
    analysisType: "content_moderation",
    results: {
      safetyScore: 95.0,
      categories: {
        violence: 0.02,
        adult: 0.01,
        hate: 0.0,
        spam: 0.03,
      },
      approved: true,
    },
    analyzedAt: new Date("2024-01-15T10:32:30Z"),
    modelVersion: "content-mod-v2.1",
  },
  {
    _id: ObjectId("507f1f77bcf86cd799439063"),
    videoId: ObjectId("507f1f77bcf86cd799439022"),
    analysisType: "metadata",
    results: {
      extractedText: "Presentation about microservices architecture",
      detectedLanguage: "en",
      topics: ["technology", "software", "architecture"],
      speakers: 1,
      audioQuality: "good",
    },
    analyzedAt: new Date("2024-01-16T14:22:00Z"),
    modelVersion: "metadata-extractor-v1.5",
  },
];

// Insert sample data
try {
  // Insert videos
  db.videos.insertMany(sampleVideos);
  print("Inserted " + sampleVideos.length + " sample videos");

  // Insert conversion jobs
  db.conversion_jobs.insertMany(sampleJobs);
  print("Inserted " + sampleJobs.length + " sample conversion jobs");

  // Insert analytics data
  db.analytics_data.insertMany(sampleAnalytics);
  print("Inserted " + sampleAnalytics.length + " sample analytics records");

  print("MongoDB seed data inserted successfully");
} catch (error) {
  print("Error inserting seed data: " + error);
}
