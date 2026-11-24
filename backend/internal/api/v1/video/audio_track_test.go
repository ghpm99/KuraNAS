package video

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"strings"
	"testing"
	"time"

	files "nas-go/api/internal/api/v1/files"
	"nas-go/api/pkg/i18n"

	"github.com/gin-gonic/gin"
)

const twoAudioTracksProbeJSON = `{"streams":[
 {"index":1,"codec_name":"aac","channels":2,"disposition":{"default":1,"forced":0},"tags":{"language":"jpn","title":"Japanese Stereo"}},
 {"index":2,"codec_name":"ac3","channels":6,"disposition":{"default":0},"tags":{"LANGUAGE":"por","TITLE":"Dublagem 5.1"}}
]}`

type audioTrackServiceFake struct {
	tracks []AudioTrack
	err    error
}

func (fake *audioTrackServiceFake) ListAudioTracks(videoFile files.FileDto) ([]AudioTrack, error) {
	return fake.tracks, fake.err
}

func (fake *audioTrackServiceFake) HasAudioStream(videoFile files.FileDto, streamIndex int) (bool, error) {
	if fake.err != nil {
		return false, fake.err
	}
	return containsAudioStreamIndex(fake.tracks, streamIndex), nil
}

func twoAudioTracksFake() *audioTrackServiceFake {
	return &audioTrackServiceFake{tracks: []AudioTrack{{StreamIndex: 1, Language: "ja"}, {StreamIndex: 2, Language: "pt"}}}
}

func TestParseAudioTracksReadsLanguageTitleCodecChannelsAndDefault(t *testing.T) {
	tracks, err := ParseAudioTracks([]byte(twoAudioTracksProbeJSON))

	if err != nil || len(tracks) != 2 {
		t.Fatalf("unexpected tracks %v err %v", tracks, err)
	}
	expectedFirst := AudioTrack{StreamIndex: 1, Language: "ja", Title: "Japanese Stereo", Codec: "aac", Channels: 2, IsDefault: true}
	if tracks[0] != expectedFirst {
		t.Fatalf("first track = %+v, want %+v", tracks[0], expectedFirst)
	}
	if tracks[1].StreamIndex != 2 || tracks[1].Language != "pt" || tracks[1].Title != "Dublagem 5.1" || tracks[1].Channels != 6 || tracks[1].IsDefault {
		t.Fatalf("unexpected second track %+v", tracks[1])
	}
}

func TestParseAudioTracksReturnsEmptyListWithoutStreams(t *testing.T) {
	tracks, err := ParseAudioTracks([]byte(`{"streams":[]}`))
	if err != nil || tracks == nil || len(tracks) != 0 {
		t.Fatalf("unexpected tracks %#v err %v", tracks, err)
	}
	if _, err := ParseAudioTracks([]byte("not json")); err == nil {
		t.Fatal("expected parse error")
	}
}

func TestAudioTrackServiceProbesOnlyAudioStreamsAndCachesByFileVersion(t *testing.T) {
	probeCalls := 0
	var probeArguments []string
	service := NewAudioTrackServiceWithRunner(func(ctx context.Context, arguments []string) ([]byte, error) {
		probeCalls++
		probeArguments = arguments
		return []byte(twoAudioTracksProbeJSON), nil
	})
	videoFile := files.FileDto{ID: 4, Path: "/data", Name: "a.mkv", Size: 10, UpdatedAt: time.Unix(1700000000, 0)}

	for range 2 {
		hasAudioStream, err := service.HasAudioStream(videoFile, 2)
		if err != nil || !hasAudioStream {
			t.Fatalf("expected stream 2, got %v err %v", hasAudioStream, err)
		}
	}
	if hasAudioStream, _ := service.HasAudioStream(videoFile, 0); hasAudioStream {
		t.Fatal("stream 0 must not be an audio stream")
	}
	if probeCalls != 1 || !strings.Contains(strings.Join(probeArguments, " "), "-select_streams a") {
		t.Fatalf("probeCalls=%d arguments=%v", probeCalls, probeArguments)
	}
	videoFile.Size = 11
	_, _ = service.ListAudioTracks(videoFile)
	if probeCalls != 2 {
		t.Fatalf("expected re-probe after file change, probeCalls=%d", probeCalls)
	}
}

func TestBuildAudioMapArguments(t *testing.T) {
	if got := strings.Join(buildAudioMapArguments(defaultAudioStreamIndex), " "); got != "-map 0:a:0?" {
		t.Fatalf("default map = %q", got)
	}
	if got := strings.Join(buildAudioMapArguments(2), " "); got != "-map 0:2" {
		t.Fatalf("indexed map = %q", got)
	}
}

func TestBuildRemuxArgumentsMapsRequestedAudioStream(t *testing.T) {
	withIndex := strings.Join(BuildRemuxArguments("/a.mkv", 0, 2), " ")
	if !strings.Contains(withIndex, "-map 0:v:0 -map 0:2 -c copy") || strings.Contains(withIndex, "0:a:0?") {
		t.Fatalf("unexpected arguments %s", withIndex)
	}
	withoutIndex := strings.Join(BuildRemuxArguments("/a.mkv", 0, defaultAudioStreamIndex), " ")
	if !strings.Contains(withoutIndex, "-map 0:v:0 -map 0:a:0? -c copy") {
		t.Fatalf("unexpected arguments %s", withoutIndex)
	}
}

func TestBuildTranscodeArgumentsMapsRequestedAudioStream(t *testing.T) {
	withIndex := strings.Join(BuildTranscodeArguments("/a.mkv", 0, 0, 720, 3), " ")
	if !strings.Contains(withIndex, "-map 0:v:0 -map 0:3 -c:v libx264") || strings.Contains(withIndex, "0:a:0?") {
		t.Fatalf("unexpected arguments %s", withIndex)
	}
	withoutIndex := strings.Join(BuildTranscodeArguments("/a.mkv", 0, 0, 720, defaultAudioStreamIndex), " ")
	if !strings.Contains(withoutIndex, "-map 0:v:0 -map 0:a:0? -c:v libx264") {
		t.Fatalf("unexpected arguments %s", withoutIndex)
	}
}

func TestRemuxHandlerPassesValidatedAudioStreamToRunner(t *testing.T) {
	var receivedArguments []string
	handler := NewRemuxHandler(newRemuxSourceFile(t), h264Codec(), capturingRunner(&receivedArguments), isPresent, 2, &videoLoggerMock{})
	handler.UseAudioTracks(twoAudioTracksFake())

	recorder := performRemuxRequest(handler, "/video/stream/1/remux?audio_stream_index=2")

	if recorder.Code != http.StatusOK || !strings.Contains(strings.Join(receivedArguments, " "), "-map 0:v:0 -map 0:2 ") {
		t.Fatalf("status %d arguments %v", recorder.Code, receivedArguments)
	}
}

func TestTranscodeHandlerPassesValidatedAudioStreamToRunner(t *testing.T) {
	var receivedArguments []string
	handler := NewTranscodeHandler(newRemuxSourceFile(t), h264Codec(), capturingRunner(&receivedArguments), isPresent, 2, &videoLoggerMock{})
	handler.UseAudioTracks(twoAudioTracksFake())

	recorder := performTranscodeRequest(handler, "/video/stream/1/transcode?audio_stream_index=1")

	if recorder.Code != http.StatusOK || !strings.Contains(strings.Join(receivedArguments, " "), "-map 0:v:0 -map 0:1 ") {
		t.Fatalf("status %d arguments %v", recorder.Code, receivedArguments)
	}
}

func TestConversionHandlersRejectInvalidAudioStreamIndex(t *testing.T) {
	useTranslationLocale(t, "pt-BR")
	testCases := []struct {
		name       string
		query      string
		service    AudioTrackServiceInterface
		messageKey string
		statusCode int
	}{
		{"not a number", "audio_stream_index=abc", twoAudioTracksFake(), "ERROR_INVALID_REQUEST", http.StatusBadRequest},
		{"negative", "audio_stream_index=-1", twoAudioTracksFake(), "ERROR_INVALID_REQUEST", http.StatusBadRequest},
		{"empty", "audio_stream_index=", twoAudioTracksFake(), "ERROR_INVALID_REQUEST", http.StatusBadRequest},
		{"not an audio stream", "audio_stream_index=0", twoAudioTracksFake(), "ERROR_VIDEO_INVALID_AUDIO_STREAM", http.StatusBadRequest},
		{"no probe service", "audio_stream_index=1", nil, "ERROR_VIDEO_INVALID_AUDIO_STREAM", http.StatusBadRequest},
		{"ffprobe missing", "audio_stream_index=1", &audioTrackServiceFake{err: exec.ErrNotFound}, "ERROR_VIDEO_AUDIO_TRACKS_UNAVAILABLE", http.StatusNotImplemented},
		{"probe failure", "audio_stream_index=1", &audioTrackServiceFake{err: errors.New("boom")}, "ERROR_VIDEO_OPERATION_FAILED", http.StatusInternalServerError},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			runnerWasCalled := false
			runner := func(ctx context.Context, arguments []string, output io.Writer) error {
				runnerWasCalled = true
				return nil
			}
			remuxHandler := NewRemuxHandler(newRemuxSourceFile(t), h264Codec(), runner, isPresent, 2, &videoLoggerMock{})
			transcodeHandler := NewTranscodeHandler(newRemuxSourceFile(t), h264Codec(), runner, isPresent, 2, &videoLoggerMock{})
			if testCase.service != nil {
				remuxHandler.UseAudioTracks(testCase.service)
				transcodeHandler.UseAudioTracks(testCase.service)
			}
			responses := []*httptest.ResponseRecorder{
				performRemuxRequest(remuxHandler, "/video/stream/1/remux?"+testCase.query),
				performTranscodeRequest(transcodeHandler, "/video/stream/1/transcode?"+testCase.query),
			}
			for _, recorder := range responses {
				if recorder.Code != testCase.statusCode || !strings.Contains(recorder.Body.String(), i18n.GetMessage(testCase.messageKey)) {
					t.Fatalf("unexpected response %d %s", recorder.Code, recorder.Body.String())
				}
			}
			if runnerWasCalled {
				t.Fatal("runner must not run")
			}
		})
	}
}

func performAudioTracksRequest(handler *AudioTrackHandler, target string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/video/metadata/:file_id/audio-tracks", handler.ListAudioTracksHandler)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, target, nil))
	return recorder
}

func TestListAudioTracksHandlerReturnsItems(t *testing.T) {
	tracks := []AudioTrack{{StreamIndex: 1, Language: "ja", Title: "Japanese", Codec: "aac", Channels: 2, IsDefault: true}}
	handler := NewAudioTrackHandler(videoSourceForEmbeddedTest(), &audioTrackServiceFake{tracks: tracks})

	recorder := performAudioTracksRequest(handler, "/video/metadata/1/audio-tracks")

	var body struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil || recorder.Code != http.StatusOK || len(body.Items) != 1 {
		t.Fatalf("unexpected response %d %s", recorder.Code, recorder.Body.String())
	}
	for _, field := range []string{"stream_index", "language", "title", "codec", "channels", "is_default"} {
		if _, exists := body.Items[0][field]; !exists {
			t.Fatalf("missing field %s in %s", field, recorder.Body.String())
		}
	}
}

func TestListAudioTracksHandlerFailures(t *testing.T) {
	useTranslationLocale(t, "pt-BR")
	testCases := []struct {
		name       string
		target     string
		source     *subtitleFileSourceFake
		service    *audioTrackServiceFake
		messageKey string
		statusCode int
	}{
		{"unknown file", "/video/metadata/99/audio-tracks", videoSourceForEmbeddedTest(), &audioTrackServiceFake{}, "ERROR_VIDEO_NOT_FOUND", http.StatusNotFound},
		{"lookup row missing", "/video/metadata/1/audio-tracks", &subtitleFileSourceFake{err: sql.ErrNoRows}, &audioTrackServiceFake{}, "ERROR_VIDEO_NOT_FOUND", http.StatusNotFound},
		{"invalid id", "/video/metadata/abc/audio-tracks", videoSourceForEmbeddedTest(), &audioTrackServiceFake{}, "ERROR_INVALID_REQUEST", http.StatusBadRequest},
		{"ffprobe missing", "/video/metadata/1/audio-tracks", videoSourceForEmbeddedTest(), &audioTrackServiceFake{err: exec.ErrNotFound}, "ERROR_VIDEO_AUDIO_TRACKS_UNAVAILABLE", http.StatusNotImplemented},
		{"probe failure", "/video/metadata/1/audio-tracks", videoSourceForEmbeddedTest(), &audioTrackServiceFake{err: errors.New("boom")}, "ERROR_VIDEO_OPERATION_FAILED", http.StatusInternalServerError},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			recorder := performAudioTracksRequest(NewAudioTrackHandler(testCase.source, testCase.service), testCase.target)
			if recorder.Code != testCase.statusCode || !strings.Contains(recorder.Body.String(), i18n.GetMessage(testCase.messageKey)) {
				t.Fatalf("unexpected response %d %s", recorder.Code, recorder.Body.String())
			}
		})
	}
}
