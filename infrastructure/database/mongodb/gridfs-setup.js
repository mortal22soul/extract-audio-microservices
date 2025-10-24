// GridFS setup and configuration script
// This script configures GridFS collections and creates additional indexes

// Switch to the application database
db = db.getSiblingDB("video_converter");

// GridFS uses two collections: fs.files and fs.chunks
// These are created automatically, but we can optimize them with additional indexes

print("Setting up GridFS collections and indexes...");

// Additional indexes for fs.files collection (metadata queries)
db.fs.files.createIndex({ "metadata.userId": 1, uploadDate: -1 });
db.fs.files.createIndex({ "metadata.contentType": 1 });
db.fs.files.createIndex({ "metadata.originalName": 1 });
db.fs.files.createIndex({ "metadata.videoId": 1 });
db.fs.files.createIndex({ "metadata.fileType": 1 }); // 'original' or 'converted'
db.fs.files.createIndex({ length: 1 }); // File size queries

// Additional indexes for fs.chunks collection (streaming performance)
db.fs.chunks.createIndex({ files_id: 1, n: 1 }, { unique: true });

// Create a helper function to store file metadata consistently
db.system.js.save({
  _id: "createVideoFileMetadata",
  value: function (userId, videoId, originalName, contentType, fileType) {
    return {
      userId: userId,
      videoId: videoId,
      originalName: originalName,
      contentType: contentType,
      fileType: fileType, // 'original' or 'converted'
      uploadedBy: "video-converter-system",
      uploadDate: new Date(),
    };
  },
});

// Create a helper function to query files by user
db.system.js.save({
  _id: "getUserFiles",
  value: function (userId, fileType) {
    var query = { "metadata.userId": userId };
    if (fileType) {
      query["metadata.fileType"] = fileType;
    }
    return db.fs.files.find(query).sort({ uploadDate: -1 });
  },
});

// Create a helper function to get file by video ID
db.system.js.save({
  _id: "getVideoFile",
  value: function (videoId, fileType) {
    var query = { "metadata.videoId": videoId };
    if (fileType) {
      query["metadata.fileType"] = fileType;
    }
    return db.fs.files.findOne(query);
  },
});

// Create sample GridFS file metadata (without actual file data)
const sampleFileMetadata = [
  {
    _id: ObjectId("507f1f77bcf86cd799439031"),
    filename: "sample_video_1.mp4",
    contentType: "video/mp4",
    length: NumberLong(15728640),
    chunkSize: 261120,
    uploadDate: new Date("2024-01-15T10:30:00Z"),
    metadata: {
      userId: ObjectId("507f1f77bcf86cd799439011"),
      videoId: ObjectId("507f1f77bcf86cd799439021"),
      originalName: "sample_video_1.mp4",
      contentType: "video/mp4",
      fileType: "original",
      uploadedBy: "video-converter-system",
    },
  },
  {
    _id: ObjectId("507f1f77bcf86cd799439041"),
    filename: "sample_video_1.mp3",
    contentType: "audio/mpeg",
    length: NumberLong(2883584),
    chunkSize: 261120,
    uploadDate: new Date("2024-01-15T10:33:30Z"),
    metadata: {
      userId: ObjectId("507f1f77bcf86cd799439011"),
      videoId: ObjectId("507f1f77bcf86cd799439021"),
      originalName: "sample_video_1.mp3",
      contentType: "audio/mpeg",
      fileType: "converted",
      uploadedBy: "video-converter-system",
      conversionSettings: {
        bitrate: "192k",
        sampleRate: 44100,
        format: "mp3",
      },
    },
  },
];

// Insert sample file metadata (in a real system, this would be done by GridFS)
try {
  db.fs.files.insertMany(sampleFileMetadata);
  print("Inserted sample GridFS file metadata");
} catch (error) {
  print("Note: Sample file metadata may already exist");
}

print("GridFS setup completed successfully");
print("Created additional indexes for optimal file queries");
print(
  "Created helper functions: createVideoFileMetadata, getUserFiles, getVideoFile",
);
