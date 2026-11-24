package video

import (
	"context"
	"time"

	files "nas-go/api/internal/api/v1/files"
)

type AudioTrackServiceInterface interface {
	ListAudioTracks(videoFile files.FileDto) ([]AudioTrack, error)
	HasAudioStream(videoFile files.FileDto, streamIndex int) (bool, error)
}

type AudioTrackService struct {
	probeRunner SubtitleProbeRunner
	timeout     func() time.Duration
	trackLists  *trackProbeCache[AudioTrack]
}

func NewAudioTrackService() *AudioTrackService {
	return NewAudioTrackServiceWithRunner(RunFFprobeSubtitleProbe)
}

func NewAudioTrackServiceWithRunner(probeRunner SubtitleProbeRunner) *AudioTrackService {
	return &AudioTrackService{
		probeRunner: probeRunner,
		timeout:     ffmpegTimeout,
		trackLists:  newTrackProbeCache[AudioTrack](),
	}
}

func (s *AudioTrackService) ListAudioTracks(videoFile files.FileDto) ([]AudioTrack, error) {
	return s.trackLists.load(videoFile, func() ([]AudioTrack, error) {
		return s.probeTracks(videoFile)
	})
}

func (s *AudioTrackService) HasAudioStream(videoFile files.FileDto, streamIndex int) (bool, error) {
	audioTracks, err := s.ListAudioTracks(videoFile)
	if err != nil {
		return false, err
	}
	return containsAudioStreamIndex(audioTracks, streamIndex), nil
}

func (s *AudioTrackService) probeTracks(videoFile files.FileDto) ([]AudioTrack, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout())
	defer cancel()
	probeJSON, err := s.probeRunner(ctx, buildAudioTrackProbeArguments(videoFile.ResolveContentPath()))
	if err != nil {
		return nil, err
	}
	return ParseAudioTracks(probeJSON)
}
