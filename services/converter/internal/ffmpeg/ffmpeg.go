package ffmpeg

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type FFmpeg struct {
	path            string
	audioBitrate    string
	audioSampleRate string
}

type ConversionOptions struct {
	InputPath       string
	OutputPath      string
	AudioBitrate    string
	AudioSampleRate string
	OnProgress      func(progress int)
}

type VideoInfo struct {
	Duration   time.Duration
	Format     string
	Resolution string
	Bitrate    int64
}

func New(ffmpegPath, audioBitrate, audioSampleRate string) *FFmpeg {
	return &FFmpeg{
		path:            ffmpegPath,
		audioBitrate:    audioBitrate,
		audioSampleRate: audioSampleRate,
	}
}

func (f *FFmpeg) GetVideoInfo(ctx context.Context, inputPath string) (*VideoInfo, error) {
	cmd := exec.CommandContext(ctx, f.path,
		"-i", inputPath,
		"-hide_banner",
		"-f", "null", "-")

	output, err := cmd.CombinedOutput()
	if err != nil {
		// FFmpeg returns non-zero exit code for info queries, but that's expected
		if !strings.Contains(string(output), "Duration:") {
			return nil, fmt.Errorf("failed to get video info: %w", err)
		}
	}

	info := &VideoInfo{}
	outputStr := string(output)

	// Parse duration
	durationRegex := regexp.MustCompile(`Duration: (\d{2}):(\d{2}):(\d{2}\.\d{2})`)
	if matches := durationRegex.FindStringSubmatch(outputStr); len(matches) == 4 {
		hours, _ := strconv.Atoi(matches[1])
		minutes, _ := strconv.Atoi(matches[2])
		seconds, _ := strconv.ParseFloat(matches[3], 64)
		
		info.Duration = time.Duration(hours)*time.Hour +
			time.Duration(minutes)*time.Minute +
			time.Duration(seconds*float64(time.Second))
	}

	// Parse resolution
	resolutionRegex := regexp.MustCompile(`(\d{3,4}x\d{3,4})`)
	if matches := resolutionRegex.FindStringSubmatch(outputStr); len(matches) > 0 {
		info.Resolution = matches[0]
	}

	// Parse bitrate
	bitrateRegex := regexp.MustCompile(`bitrate: (\d+) kb/s`)
	if matches := bitrateRegex.FindStringSubmatch(outputStr); len(matches) == 2 {
		if bitrate, err := strconv.ParseInt(matches[1], 10, 64); err == nil {
			info.Bitrate = bitrate * 1000 // Convert to bits per second
		}
	}

	log.Printf("Video info for %s: duration=%v, resolution=%s, bitrate=%d",
		inputPath, info.Duration, info.Resolution, info.Bitrate)

	return info, nil
}

func (f *FFmpeg) ConvertToMP3(ctx context.Context, options ConversionOptions) error {
	// Get video info for progress tracking
	videoInfo, err := f.GetVideoInfo(ctx, options.InputPath)
	if err != nil {
		return fmt.Errorf("failed to get video info: %w", err)
	}

	// Use provided options or defaults
	bitrate := options.AudioBitrate
	if bitrate == "" {
		bitrate = f.audioBitrate
	}
	
	sampleRate := options.AudioSampleRate
	if sampleRate == "" {
		sampleRate = f.audioSampleRate
	}

	// Build FFmpeg command
	args := []string{
		"-i", options.InputPath,
		"-vn",                    // No video
		"-acodec", "libmp3lame",  // MP3 codec
		"-ab", bitrate,           // Audio bitrate
		"-ar", sampleRate,        // Audio sample rate
		"-f", "mp3",              // Output format
		"-y",                     // Overwrite output file
		"-progress", "pipe:1",    // Progress to stdout
		options.OutputPath,
	}

	cmd := exec.CommandContext(ctx, f.path, args...)

	// Create pipes for progress monitoring
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start FFmpeg: %w", err)
	}

	// Monitor progress
	go f.monitorProgress(stdout, videoInfo.Duration, options.OnProgress)

	// Monitor errors
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, "error") || strings.Contains(line, "Error") {
				log.Printf("FFmpeg error: %s", line)
			}
		}
	}()

	// Wait for completion
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("FFmpeg conversion failed: %w", err)
	}

	log.Printf("Successfully converted %s to %s", options.InputPath, options.OutputPath)
	return nil
}

func (f *FFmpeg) monitorProgress(stdout io.ReadCloser, totalDuration time.Duration, onProgress func(int)) {
	scanner := bufio.NewScanner(stdout)
	
	for scanner.Scan() {
		line := scanner.Text()
		
		// Parse progress from FFmpeg output
		if strings.HasPrefix(line, "out_time_ms=") {
			timeStr := strings.TrimPrefix(line, "out_time_ms=")
			if timeMicros, err := strconv.ParseInt(timeStr, 10, 64); err == nil {
				currentTime := time.Duration(timeMicros) * time.Microsecond
				
				if totalDuration > 0 {
					progress := int((currentTime * 100) / totalDuration)
					if progress > 100 {
						progress = 100
					}
					if progress >= 0 && onProgress != nil {
						onProgress(progress)
					}
				}
			}
		}
	}
}

func (f *FFmpeg) ValidateInstallation(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, f.path, "-version")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("FFmpeg not found or not executable: %w", err)
	}

	if !strings.Contains(string(output), "ffmpeg version") {
		return fmt.Errorf("invalid FFmpeg installation")
	}

	log.Printf("FFmpeg validation successful: %s", strings.Split(string(output), "\n")[0])
	return nil
}