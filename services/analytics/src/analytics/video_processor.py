"""Video processing utilities for metadata extraction and thumbnail generation"""
import os
import json
import logging
import subprocess
from typing import Dict, Any, List, Optional, Tuple
import cv2
import numpy as np
from PIL import Image

from .config import config

logger = logging.getLogger(__name__)


class VideoMetadataExtractor:
    """Extract comprehensive metadata from video files"""
    
    @staticmethod
    async def extract_metadata(video_path: str) -> Dict[str, Any]:
        """Extract comprehensive video metadata using OpenCV and FFmpeg"""
        try:
            # Validate file exists
            if not os.path.exists(video_path):
                raise FileNotFoundError(f"Video file not found: {video_path}")
            
            # Extract basic metadata with OpenCV
            opencv_metadata = VideoMetadataExtractor._extract_opencv_metadata(video_path)
            
            # Extract detailed metadata with FFprobe
            ffprobe_metadata = await VideoMetadataExtractor._extract_ffprobe_metadata(video_path)
            
            # Combine metadata
            combined_metadata = {
                **opencv_metadata,
                "detailed": ffprobe_metadata,
                "file_path": video_path,
                "file_name": os.path.basename(video_path),
                "file_size": os.path.getsize(video_path)
            }
            
            # Calculate additional derived properties
            combined_metadata.update(
                VideoMetadataExtractor._calculate_derived_properties(combined_metadata)
            )
            
            logger.info(f"Successfully extracted metadata for: {video_path}")
            return combined_metadata
            
        except Exception as e:
            logger.error(f"Metadata extraction failed for {video_path}: {e}")
            return VideoMetadataExtractor._get_error_metadata(video_path, str(e))
    
    @staticmethod
    def _extract_opencv_metadata(video_path: str) -> Dict[str, Any]:
        """Extract basic metadata using OpenCV"""
        metadata = {}
        cap = None
        
        try:
            cap = cv2.VideoCapture(video_path)
            
            if not cap.isOpened():
                raise ValueError(f"Could not open video file with OpenCV: {video_path}")
            
            # Get basic video properties
            fps = cap.get(cv2.CAP_PROP_FPS)
            frame_count = int(cap.get(cv2.CAP_PROP_FRAME_COUNT))
            width = int(cap.get(cv2.CAP_PROP_FRAME_WIDTH))
            height = int(cap.get(cv2.CAP_PROP_FRAME_HEIGHT))
            
            # Calculate duration
            duration = frame_count / fps if fps > 0 else 0
            
            # Get codec information
            fourcc = int(cap.get(cv2.CAP_PROP_FOURCC))
            codec = "".join([chr((fourcc >> 8 * i) & 0xFF) for i in range(4)])
            
            metadata = {
                "duration": duration,
                "fps": fps,
                "width": width,
                "height": height,
                "frame_count": frame_count,
                "resolution": f"{width}x{height}",
                "codec": codec.strip(),
                "opencv_version": cv2.__version__
            }
            
        except Exception as e:
            logger.warning(f"OpenCV metadata extraction failed: {e}")
            metadata = {
                "duration": 0,
                "fps": 0,
                "width": 0,
                "height": 0,
                "frame_count": 0,
                "resolution": "unknown",
                "codec": "unknown",
                "opencv_error": str(e)
            }
        
        finally:
            if cap:
                cap.release()
        
        return metadata
    
    @staticmethod
    async def _extract_ffprobe_metadata(video_path: str) -> Dict[str, Any]:
        """Extract detailed metadata using FFprobe"""
        try:
            # Build FFprobe command
            ffprobe_cmd = [
                "ffprobe",
                "-v", "quiet",
                "-print_format", "json",
                "-show_format",
                "-show_streams",
                "-show_chapters",
                video_path
            ]
            
            # Execute FFprobe
            result = subprocess.run(
                ffprobe_cmd,
                capture_output=True,
                text=True,
                timeout=30  # 30 second timeout
            )
            
            if result.returncode == 0:
                return json.loads(result.stdout)
            else:
                logger.warning(f"FFprobe failed with return code {result.returncode}: {result.stderr}")
                return {"error": result.stderr}
                
        except subprocess.TimeoutExpired:
            logger.error("FFprobe timed out")
            return {"error": "FFprobe timeout"}
        except json.JSONDecodeError as e:
            logger.error(f"Failed to parse FFprobe JSON output: {e}")
            return {"error": f"JSON parse error: {e}"}
        except Exception as e:
            logger.error(f"FFprobe execution failed: {e}")
            return {"error": str(e)}
    
    @staticmethod
    def _calculate_derived_properties(metadata: Dict[str, Any]) -> Dict[str, Any]:
        """Calculate additional derived properties from metadata"""
        derived = {}
        
        try:
            width = metadata.get("width", 0)
            height = metadata.get("height", 0)
            
            # Aspect ratio
            if height > 0:
                derived["aspect_ratio"] = width / height
            else:
                derived["aspect_ratio"] = 1.0
            
            # Resolution category
            if height >= 2160:
                derived["resolution_category"] = "4K"
            elif height >= 1440:
                derived["resolution_category"] = "2K"
            elif height >= 1080:
                derived["resolution_category"] = "Full HD"
            elif height >= 720:
                derived["resolution_category"] = "HD"
            elif height >= 480:
                derived["resolution_category"] = "SD"
            else:
                derived["resolution_category"] = "Low"
            
            # Bitrate estimation from file size and duration
            file_size = metadata.get("file_size", 0)
            duration = metadata.get("duration", 0)
            
            if duration > 0 and file_size > 0:
                # Bitrate in kbps
                derived["estimated_bitrate"] = (file_size * 8) / (duration * 1000)
            else:
                derived["estimated_bitrate"] = 0
            
            # Frame rate category
            fps = metadata.get("fps", 0)
            if fps >= 60:
                derived["fps_category"] = "High"
            elif fps >= 30:
                derived["fps_category"] = "Standard"
            elif fps >= 24:
                derived["fps_category"] = "Cinema"
            else:
                derived["fps_category"] = "Low"
            
        except Exception as e:
            logger.warning(f"Failed to calculate derived properties: {e}")
        
        return derived
    
    @staticmethod
    def _get_error_metadata(video_path: str, error: str) -> Dict[str, Any]:
        """Return error metadata when extraction fails"""
        return {
            "duration": 0,
            "fps": 0,
            "width": 0,
            "height": 0,
            "frame_count": 0,
            "resolution": "unknown",
            "aspect_ratio": 1.0,
            "codec": "unknown",
            "file_path": video_path,
            "file_name": os.path.basename(video_path) if video_path else "unknown",
            "file_size": 0,
            "resolution_category": "unknown",
            "estimated_bitrate": 0,
            "fps_category": "unknown",
            "error": error,
            "extraction_failed": True
        }


class ThumbnailGenerator:
    """Generate video thumbnails at specified intervals"""
    
    def __init__(self):
        self.thumbnail_width = config.THUMBNAIL_WIDTH
        self.thumbnail_height = config.THUMBNAIL_HEIGHT
        self.thumbnail_count = config.THUMBNAIL_COUNT
        self.thumbnails_dir = config.THUMBNAILS_DIR
    
    async def generate_thumbnails(
        self,
        video_path: str,
        video_id: str,
        timestamps: Optional[List[float]] = None,
        count: Optional[int] = None
    ) -> List[Dict[str, Any]]:
        """Generate video thumbnails at specified timestamps or intervals"""
        try:
            # Validate input
            if not os.path.exists(video_path):
                raise FileNotFoundError(f"Video file not found: {video_path}")
            
            # Create thumbnails directory
            os.makedirs(self.thumbnails_dir, exist_ok=True)
            
            # Get video metadata for duration
            cap = cv2.VideoCapture(video_path)
            if not cap.isOpened():
                raise ValueError(f"Could not open video file: {video_path}")
            
            fps = cap.get(cv2.CAP_PROP_FPS)
            frame_count = int(cap.get(cv2.CAP_PROP_FRAME_COUNT))
            duration = frame_count / fps if fps > 0 else 0
            
            if duration <= 0:
                cap.release()
                raise ValueError("Video has zero duration")
            
            # Determine timestamps for thumbnails
            if timestamps is None:
                timestamps = self._calculate_thumbnail_timestamps(duration, count or self.thumbnail_count)
            
            thumbnails = []
            
            for i, timestamp in enumerate(timestamps):
                try:
                    thumbnail_data = await self._generate_single_thumbnail(
                        cap, video_id, timestamp, i, fps
                    )
                    if thumbnail_data:
                        thumbnails.append(thumbnail_data)
                        
                except Exception as e:
                    logger.warning(f"Failed to generate thumbnail at {timestamp}s: {e}")
                    continue
            
            cap.release()
            
            logger.info(f"Generated {len(thumbnails)} thumbnails for video {video_id}")
            return thumbnails
            
        except Exception as e:
            logger.error(f"Thumbnail generation failed for {video_path}: {e}")
            return []
    
    def _calculate_thumbnail_timestamps(self, duration: float, count: int) -> List[float]:
        """Calculate evenly distributed timestamps for thumbnails"""
        if count <= 0:
            return []
        
        if count == 1:
            return [duration / 2]  # Middle of video
        
        # Avoid very beginning and end of video
        start_offset = min(5.0, duration * 0.05)  # 5 seconds or 5% of duration
        end_offset = min(5.0, duration * 0.05)
        
        effective_duration = duration - start_offset - end_offset
        
        if effective_duration <= 0:
            return [duration / 2]
        
        timestamps = []
        for i in range(count):
            timestamp = start_offset + (i * effective_duration / (count - 1))
            timestamps.append(timestamp)
        
        return timestamps
    
    async def _generate_single_thumbnail(
        self,
        cap: cv2.VideoCapture,
        video_id: str,
        timestamp: float,
        index: int,
        fps: float
    ) -> Optional[Dict[str, Any]]:
        """Generate a single thumbnail at the specified timestamp"""
        try:
            # Seek to timestamp
            frame_number = int(timestamp * fps)
            cap.set(cv2.CAP_PROP_POS_FRAMES, frame_number)
            
            ret, frame = cap.read()
            if not ret:
                logger.warning(f"Could not read frame at timestamp {timestamp}")
                return None
            
            # Get original dimensions
            original_height, original_width = frame.shape[:2]
            
            # Resize frame while maintaining aspect ratio
            resized_frame = self._resize_frame_with_aspect_ratio(
                frame, self.thumbnail_width, self.thumbnail_height
            )
            
            # Convert BGR to RGB for PIL
            frame_rgb = cv2.cvtColor(resized_frame, cv2.COLOR_BGR2RGB)
            
            # Generate filename
            thumbnail_filename = f"{video_id}_thumb_{index}_{int(timestamp)}.jpg"
            thumbnail_path = os.path.join(self.thumbnails_dir, thumbnail_filename)
            
            # Save thumbnail using PIL for better quality control
            pil_image = Image.fromarray(frame_rgb)
            pil_image.save(thumbnail_path, "JPEG", quality=85, optimize=True)
            
            # Get final dimensions
            final_height, final_width = resized_frame.shape[:2]
            
            return {
                "url": f"/thumbnails/{thumbnail_filename}",
                "timestamp_seconds": int(timestamp),
                "width": final_width,
                "height": final_height,
                "file_path": thumbnail_path,
                "file_size": os.path.getsize(thumbnail_path)
            }
            
        except Exception as e:
            logger.error(f"Failed to generate thumbnail at {timestamp}s: {e}")
            return None
    
    def _resize_frame_with_aspect_ratio(
        self,
        frame: np.ndarray,
        target_width: int,
        target_height: int
    ) -> np.ndarray:
        """Resize frame while maintaining aspect ratio"""
        height, width = frame.shape[:2]
        
        # Calculate scaling factor
        scale_w = target_width / width
        scale_h = target_height / height
        scale = min(scale_w, scale_h)
        
        # Calculate new dimensions
        new_width = int(width * scale)
        new_height = int(height * scale)
        
        # Resize frame
        resized = cv2.resize(frame, (new_width, new_height), interpolation=cv2.INTER_AREA)
        
        # If the resized frame doesn't match target dimensions, pad with black
        if new_width != target_width or new_height != target_height:
            # Create black canvas
            canvas = np.zeros((target_height, target_width, 3), dtype=np.uint8)
            
            # Calculate position to center the resized frame
            y_offset = (target_height - new_height) // 2
            x_offset = (target_width - new_width) // 2
            
            # Place resized frame on canvas
            canvas[y_offset:y_offset + new_height, x_offset:x_offset + new_width] = resized
            
            return canvas
        
        return resized
    
    async def generate_smart_thumbnails(
        self,
        video_path: str,
        video_id: str,
        count: int = None
    ) -> List[Dict[str, Any]]:
        """Generate thumbnails using smart scene detection"""
        try:
            count = count or self.thumbnail_count
            
            # For now, use evenly distributed timestamps
            # In a more advanced implementation, you could use scene detection
            # to find the most representative frames
            
            return await self.generate_thumbnails(video_path, video_id, count=count)
            
        except Exception as e:
            logger.error(f"Smart thumbnail generation failed: {e}")
            return await self.generate_thumbnails(video_path, video_id, count=count)


class VideoIndexer:
    """Create searchable indexes from video metadata"""
    
    @staticmethod
    def create_search_index(metadata: Dict[str, Any]) -> Dict[str, Any]:
        """Create searchable index from video metadata"""
        index = {}
        
        try:
            # Basic properties for search
            index["duration_category"] = VideoIndexer._categorize_duration(
                metadata.get("duration", 0)
            )
            index["resolution_category"] = metadata.get("resolution_category", "unknown")
            index["fps_category"] = metadata.get("fps_category", "unknown")
            index["aspect_ratio_category"] = VideoIndexer._categorize_aspect_ratio(
                metadata.get("aspect_ratio", 1.0)
            )
            
            # File properties
            index["file_size_category"] = VideoIndexer._categorize_file_size(
                metadata.get("file_size", 0)
            )
            
            # Codec information
            index["codec"] = metadata.get("codec", "unknown").lower()
            
            # Extract searchable text from detailed metadata
            detailed = metadata.get("detailed", {})
            index["searchable_text"] = VideoIndexer._extract_searchable_text(detailed)
            
        except Exception as e:
            logger.error(f"Failed to create search index: {e}")
        
        return index
    
    @staticmethod
    def _categorize_duration(duration: float) -> str:
        """Categorize video duration"""
        if duration < 30:
            return "very_short"
        elif duration < 300:  # 5 minutes
            return "short"
        elif duration < 1800:  # 30 minutes
            return "medium"
        elif duration < 3600:  # 1 hour
            return "long"
        else:
            return "very_long"
    
    @staticmethod
    def _categorize_aspect_ratio(aspect_ratio: float) -> str:
        """Categorize aspect ratio"""
        if abs(aspect_ratio - 16/9) < 0.1:
            return "widescreen"
        elif abs(aspect_ratio - 4/3) < 0.1:
            return "standard"
        elif aspect_ratio > 2.0:
            return "ultra_wide"
        elif aspect_ratio < 1.0:
            return "portrait"
        else:
            return "other"
    
    @staticmethod
    def _categorize_file_size(file_size: int) -> str:
        """Categorize file size"""
        mb = file_size / (1024 * 1024)
        
        if mb < 10:
            return "small"
        elif mb < 100:
            return "medium"
        elif mb < 1000:
            return "large"
        else:
            return "very_large"
    
    @staticmethod
    def _extract_searchable_text(detailed_metadata: Dict[str, Any]) -> str:
        """Extract searchable text from detailed metadata"""
        text_parts = []
        
        try:
            # Extract from format tags
            format_info = detailed_metadata.get("format", {})
            tags = format_info.get("tags", {})
            
            for key, value in tags.items():
                if isinstance(value, str) and key.lower() in [
                    "title", "artist", "album", "comment", "description", "genre"
                ]:
                    text_parts.append(value)
            
            # Extract from stream information
            streams = detailed_metadata.get("streams", [])
            for stream in streams:
                stream_tags = stream.get("tags", {})
                for key, value in stream_tags.items():
                    if isinstance(value, str) and key.lower() in [
                        "title", "language", "handler_name"
                    ]:
                        text_parts.append(value)
        
        except Exception as e:
            logger.warning(f"Failed to extract searchable text: {e}")
        
        return " ".join(text_parts).lower()


# Global instances
metadata_extractor = VideoMetadataExtractor()
thumbnail_generator = ThumbnailGenerator()
video_indexer = VideoIndexer()