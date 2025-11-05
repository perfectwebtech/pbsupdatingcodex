package service

import (
	"fmt"
	"strings"

	"github.com/iptv-platform/streaming-gateway/internal/config"
	"github.com/iptv-platform/streaming-gateway/internal/model"
	"go.uber.org/zap"
)

type HLSService struct {
	config config.StreamingConfig
	logger *zap.Logger
}

func NewHLSService(config config.StreamingConfig, logger *zap.Logger) *HLSService {
	return &HLSService{
		config: config,
		logger: logger,
	}
}

type HLSPlaylist struct {
	Content     string
	ContentType string
}

func (s *HLSService) GenerateMasterPlaylist(stream *model.Stream, baseURL string) (*HLSPlaylist, error) {
	var builder strings.Builder

	// M3U8 header
	builder.WriteString("#EXTM3U\n")
	builder.WriteString("#EXT-X-VERSION:3\n")

	// Multiple bitrate variants
	variants := []struct {
		bandwidth  int
		resolution string
		name       string
	}{
		{8000000, "1920x1080", "1080p"},
		{5000000, "1280x720", "720p"},
		{3000000, "854x480", "480p"},
		{1500000, "640x360", "360p"},
	}

	for _, variant := range variants {
		if variant.bandwidth <= s.config.MaxBitrate {
			builder.WriteString(fmt.Sprintf("#EXT-X-STREAM-INF:BANDWIDTH=%d,RESOLUTION=%s,NAME=\"%s\"\n",
				variant.bandwidth,
				variant.resolution,
				variant.name,
			))
			builder.WriteString(fmt.Sprintf("%s/%s/playlist.m3u8\n", baseURL, variant.name))
		}
	}

	return &HLSPlaylist{
		Content:     builder.String(),
		ContentType: "application/vnd.apple.mpegurl",
	}, nil
}

func (s *HLSService) GenerateMediaPlaylist(
	stream *model.Stream,
	quality string,
	segmentBaseURL string,
	sessionID string,
) (*HLSPlaylist, error) {
	var builder strings.Builder

	// M3U8 header
	builder.WriteString("#EXTM3U\n")
	builder.WriteString("#EXT-X-VERSION:3\n")
	builder.WriteString(fmt.Sprintf("#EXT-X-TARGETDURATION:%d\n", s.config.SegmentDuration))
	builder.WriteString("#EXT-X-MEDIA-SEQUENCE:0\n")

	// For live streams
	if stream.Type == "live" {
		// Generate recent segments (last 10 segments for a rolling window)
		numSegments := 10
		for i := 0; i < numSegments; i++ {
			builder.WriteString(fmt.Sprintf("#EXTINF:%d.0,\n", s.config.SegmentDuration))
			builder.WriteString(fmt.Sprintf("%s/segment_%d.ts?session=%s\n",
				segmentBaseURL,
				i,
				sessionID,
			))
		}
	} else {
		// For VOD, generate all segments
		// This would typically come from pre-segmented files
		builder.WriteString("#EXT-X-PLAYLIST-TYPE:VOD\n")

		// Calculate number of segments based on duration
		durationSeconds := stream.Duration // Assuming duration is in seconds
		numSegments := (durationSeconds + s.config.SegmentDuration - 1) / s.config.SegmentDuration

		for i := 0; i < numSegments; i++ {
			builder.WriteString(fmt.Sprintf("#EXTINF:%d.0,\n", s.config.SegmentDuration))
			builder.WriteString(fmt.Sprintf("%s/segment_%d.ts?session=%s\n",
				segmentBaseURL,
				i,
				sessionID,
			))
		}

		builder.WriteString("#EXT-X-ENDLIST\n")
	}

	return &HLSPlaylist{
		Content:     builder.String(),
		ContentType: "application/vnd.apple.mpegurl",
	}, nil
}

func (s *HLSService) GenerateSimplePlaylist(
	segmentBaseURL string,
	sessionID string,
	isLive bool,
) (*HLSPlaylist, error) {
	var builder strings.Builder

	builder.WriteString("#EXTM3U\n")
	builder.WriteString("#EXT-X-VERSION:3\n")
	builder.WriteString(fmt.Sprintf("#EXT-X-TARGETDURATION:%d\n", s.config.SegmentDuration))
	builder.WriteString("#EXT-X-MEDIA-SEQUENCE:0\n")

	if !isLive {
		builder.WriteString("#EXT-X-PLAYLIST-TYPE:VOD\n")
	}

	// Generate segments
	numSegments := 10
	if !isLive {
		numSegments = 100 // More segments for VOD
	}

	for i := 0; i < numSegments; i++ {
		builder.WriteString(fmt.Sprintf("#EXTINF:%d.0,\n", s.config.SegmentDuration))
		builder.WriteString(fmt.Sprintf("%s/segment_%d.ts?session=%s\n",
			segmentBaseURL,
			i,
			sessionID,
		))
	}

	if !isLive {
		builder.WriteString("#EXT-X-ENDLIST\n")
	}

	return &HLSPlaylist{
		Content:     builder.String(),
		ContentType: "application/vnd.apple.mpegurl",
	}, nil
}
