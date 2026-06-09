// MongoDB initialization script for video converter microservices
// This script sets up the database, collections, indexes, and users

// Switch to the application database
db = db.getSiblingDB("video_converter");

// Create application user
db.createUser({
  user: "app_user",
  pwd: "dev_password_123",
  roles: [
    {
      role: "readWrite",
      db: "video_converter",
    },
  ],
});

// Create collections with validation schemas

// Videos collection - stores video metadata and processing status
db.createCollection("videos", {
  validator: {
    $jsonSchema: {
      bsonType: "object",
      required: ["userId", "originalFilename", "mimeType", "size", "status"],
      properties: {
        userId: {
          bsonType: "objectId",
          description: "User ID who uploaded the video",
        },
        originalFilename: {
          bsonType: "string",
          description: "Original filename of the uploaded video",
        },
        mimeType: {
          bsonType: "string",
          description: "MIME type of the video file",
        },
        size: {
          bsonType: "long",
          minimum: 0,
          description: "File size in bytes",
        },
        status: {
          bsonType: "string",
          enum: ["uploaded", "processing", "completed", "failed"],
          description: "Current processing status",
        },
        uploadedAt: {
          bsonType: "date",
          description: "Upload timestamp",
        },
        conversionJobId: {
          bsonType: "string",
          description: "ID of the conversion job",
        },
        gridfsFileId: {
          bsonType: "objectId",
          description: "GridFS file ID for the original video",
        },
        mp3FileId: {
          bsonType: "objectId",
          description: "GridFS file ID for the converted MP3",
        },
        metadata: {
          bsonType: "object",
          properties: {
            duration: {
              bsonType: "double",
              minimum: 0,
              description: "Video duration in seconds",
            },
            resolution: {
              bsonType: "string",
              description: "Video resolution (e.g., 1920x1080)",
            },
            codec: {
              bsonType: "string",
              description: "Video codec",
            },
            bitrate: {
              bsonType: "long",
              minimum: 0,
              description: "Video bitrate",
            },
            fps: {
              bsonType: "double",
              minimum: 0,
              description: "Frames per second",
            },
          },
        },
        analytics: {
          bsonType: "object",
          properties: {
            thumbnails: {
              bsonType: "array",
              items: {
                bsonType: "string",
              },
              description: "Array of thumbnail URLs",
            },
            qualityScore: {
              bsonType: "double",
              minimum: 0,
              maximum: 100,
              description: "Video quality score (0-100)",
            },
            safetyScore: {
              bsonType: "double",
              minimum: 0,
              maximum: 100,
              description: "Content safety score (0-100)",
            },
            tags: {
              bsonType: "array",
              items: {
                bsonType: "string",
              },
              description: "Content tags",
            },
          },
        },
      },
    },
  },
});

// Conversion jobs collection - tracks video processing jobs
db.createCollection("conversion_jobs", {
  validator: {
    $jsonSchema: {
      bsonType: "object",
      required: ["videoId", "userId", "status"],
      properties: {
        videoId: {
          bsonType: "objectId",
          description: "Reference to the video being processed",
        },
        userId: {
          bsonType: "objectId",
          description: "User ID who owns the video",
        },
        status: {
          bsonType: "string",
          enum: ["pending", "processing", "completed", "failed"],
          description: "Job status",
        },
        progress: {
          bsonType: "int",
          minimum: 0,
          maximum: 100,
          description: "Processing progress percentage",
        },
        startedAt: {
          bsonType: "date",
          description: "Job start timestamp",
        },
        completedAt: {
          bsonType: "date",
          description: "Job completion timestamp",
        },
        errorMessage: {
          bsonType: "string",
          description: "Error message if job failed",
        },
        processingNode: {
          bsonType: "string",
          description: "ID of the processing node handling this job",
        },
        conversionSettings: {
          bsonType: "object",
          properties: {
            outputFormat: {
              bsonType: "string",
              description: "Output format (mp3, wav, etc.)",
            },
            bitrate: {
              bsonType: "string",
              description: "Output bitrate (128k, 192k, 320k)",
            },
            sampleRate: {
              bsonType: "int",
              description: "Sample rate in Hz",
            },
          },
        },
      },
    },
  },
});

// Analytics data collection - stores ML analysis results
db.createCollection("analytics_data", {
  validator: {
    $jsonSchema: {
      bsonType: "object",
      required: ["videoId", "analysisType", "results"],
      properties: {
        videoId: {
          bsonType: "objectId",
          description: "Reference to the analyzed video",
        },
        analysisType: {
          bsonType: "string",
          enum: [
            "metadata",
            "quality",
            "content_moderation",
            "recommendations",
          ],
          description: "Type of analysis performed",
        },
        results: {
          bsonType: "object",
          description: "Analysis results data",
        },
        analyzedAt: {
          bsonType: "date",
          description: "Analysis timestamp",
        },
        modelVersion: {
          bsonType: "string",
          description: "Version of the ML model used",
        },
      },
    },
  },
});

// Create indexes for efficient queries

// Videos collection indexes
db.videos.createIndex({ userId: 1 });
db.videos.createIndex({ status: 1 });
db.videos.createIndex({ uploadedAt: -1 });
db.videos.createIndex({ userId: 1, status: 1 });
db.videos.createIndex({ userId: 1, uploadedAt: -1 });
db.videos.createIndex({ conversionJobId: 1 });
db.videos.createIndex({ gridfsFileId: 1 });
db.videos.createIndex({ mp3FileId: 1 });

// Conversion jobs collection indexes
db.conversion_jobs.createIndex({ videoId: 1 });
db.conversion_jobs.createIndex({ userId: 1 });
db.conversion_jobs.createIndex({ status: 1 });
db.conversion_jobs.createIndex({ startedAt: -1 });
db.conversion_jobs.createIndex({ userId: 1, status: 1 });
db.conversion_jobs.createIndex({ processingNode: 1, status: 1 });

// Analytics data collection indexes
db.analytics_data.createIndex({ videoId: 1 });
db.analytics_data.createIndex({ analysisType: 1 });
db.analytics_data.createIndex({ analyzedAt: -1 });
db.analytics_data.createIndex({ videoId: 1, analysisType: 1 });

// GridFS collections are created automatically, but we can create indexes
// for better performance on file metadata queries
db.fs.files.createIndex({ filename: 1 });
db.fs.files.createIndex({ uploadDate: -1 });
db.fs.files.createIndex({ "metadata.userId": 1 });
db.fs.files.createIndex({ "metadata.contentType": 1 });

// Switch to Auth database and configure collections
authDb = db.getSiblingDB("auth");

authDb.createCollection("users", {
  validator: {
    $jsonSchema: {
      bsonType: "object",
      required: ["email", "password_hash"],
      properties: {
        email: { bsonType: "string" },
        password_hash: { bsonType: "string" },
      }
    }
  }
});

authDb.createCollection("user_sessions", {
  validator: {
    $jsonSchema: {
      bsonType: "object",
      required: ["user_id", "token_hash", "expires_at"],
    }
  }
});

authDb.users.createIndex({ email: 1 }, { unique: true });
authDb.user_sessions.createIndex({ user_id: 1 });
authDb.user_sessions.createIndex({ token_hash: 1 }, { unique: true });

print("MongoDB initialization completed successfully");
print("Created databases: video_converter, auth");
print("Created indexes for optimal query performance");
