package music

import (
	"bytes"
	"encoding/binary"
	"errors"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"nas-go/api/internal/api/v1/files"
	"nas-go/api/pkg/utils"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func buildTestPNG(t *testing.T, width, height int) []byte {
	t.Helper()
	canvas := image.NewRGBA(image.Rect(0, 0, width, height))
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			canvas.Set(x, y, color.RGBA{R: 200, G: 30, B: 30, A: 255})
		}
	}
	var encoded bytes.Buffer
	if err := png.Encode(&encoded, canvas); err != nil {
		t.Fatal(err)
	}
	return encoded.Bytes()
}

func synchsafeBytes(value int) []byte {
	return []byte{byte(value >> 21 & 0x7f), byte(value >> 14 & 0x7f), byte(value >> 7 & 0x7f), byte(value & 0x7f)}
}

func buildAPICFrameBody(pictureType byte, imageBytes []byte) []byte {
	frameBody := []byte{0}
	frameBody = append(frameBody, []byte("image/png")...)
	frameBody = append(frameBody, 0, pictureType)
	frameBody = append(frameBody, []byte("desc")...)
	frameBody = append(frameBody, 0)
	return append(frameBody, imageBytes...)
}

func buildID3v2Tag(t *testing.T, majorVersion byte, frames ...[]byte) []byte {
	t.Helper()
	var tagBody []byte
	for _, frame := range frames {
		tagBody = append(tagBody, frame...)
	}
	tag := append([]byte("ID3"), majorVersion, 0, 0)
	tag = append(tag, synchsafeBytes(len(tagBody))...)
	return append(tag, tagBody...)
}

func buildID3Frame(majorVersion byte, frameID string, frameBody []byte) []byte {
	frame := []byte(frameID)
	if majorVersion == 4 {
		frame = append(frame, synchsafeBytes(len(frameBody))...)
	} else {
		sizeBytes := make([]byte, 4)
		binary.BigEndian.PutUint32(sizeBytes, uint32(len(frameBody)))
		frame = append(frame, sizeBytes...)
	}
	frame = append(frame, 0, 0)
	return append(frame, frameBody...)
}

func TestExtractID3PicturePrefersFrontCover(t *testing.T) {
	backCover := []byte("back-cover-bytes")
	frontCover := []byte("front-cover-bytes")
	for _, majorVersion := range []byte{3, 4} {
		tag := buildID3v2Tag(t, majorVersion,
			buildID3Frame(majorVersion, "TIT2", []byte{0, 'x'}),
			buildID3Frame(majorVersion, "APIC", buildAPICFrameBody(4, backCover)),
			buildID3Frame(majorVersion, "APIC", buildAPICFrameBody(3, frontCover)),
		)
		extracted, isFound := extractID3Picture(bytes.NewReader(tag))
		if !isFound || !bytes.Equal(extracted, frontCover) {
			t.Fatalf("v2.%d: expected front cover, got %q found=%v", majorVersion, extracted, isFound)
		}
	}
}

func TestExtractID3PictureLegacyPICFrame(t *testing.T) {
	pictureBytes := []byte("legacy-picture")
	frameBody := append([]byte{0, 'P', 'N', 'G', 3, 'd', 0}, pictureBytes...)
	frame := append([]byte("PIC"), byte(len(frameBody)>>16), byte(len(frameBody)>>8), byte(len(frameBody)))
	frame = append(frame, frameBody...)
	extracted, isFound := extractID3Picture(bytes.NewReader(buildID3v2Tag(t, 2, frame)))
	if !isFound || !bytes.Equal(extracted, pictureBytes) {
		t.Fatalf("unexpected legacy extraction %q %v", extracted, isFound)
	}
}

func TestExtractID3PictureUtf16Description(t *testing.T) {
	pictureBytes := []byte("utf16-picture")
	frameBody := []byte{1}
	frameBody = append(frameBody, []byte("image/jpeg")...)
	frameBody = append(frameBody, 0, 3, 0xff, 0xfe, 'a', 0, 0, 0)
	frameBody = append(frameBody, pictureBytes...)
	tag := buildID3v2Tag(t, 3, buildID3Frame(3, "APIC", frameBody))
	extracted, isFound := extractID3Picture(bytes.NewReader(tag))
	if !isFound || !bytes.Equal(extracted, pictureBytes) {
		t.Fatalf("unexpected utf16 extraction %q %v", extracted, isFound)
	}
}

func TestExtractID3PictureRejectsInvalidInput(t *testing.T) {
	invalidInputs := map[string][]byte{
		"empty":            {},
		"not id3":          []byte("RIFFxxxxxxxxxxxx"),
		"unsupported":      append([]byte("ID3"), 9, 0, 0, 0, 0, 0, 5),
		"truncated body":   append([]byte("ID3"), 3, 0, 0, 0, 0, 1, 0),
		"no picture":       buildID3v2Tag(t, 3, buildID3Frame(3, "TIT2", []byte{0, 'x'})),
		"frame overflows":  buildID3v2Tag(t, 3, append([]byte("APIC"), 0, 0, 1, 0, 0, 0, 1)),
		"missing mime end": buildID3v2Tag(t, 3, buildID3Frame(3, "APIC", []byte{0, 'i', 'm'})),
		"empty picture":    buildID3v2Tag(t, 3, buildID3Frame(3, "APIC", buildAPICFrameBody(3, nil))),
	}
	for caseName, input := range invalidInputs {
		if extracted, isFound := extractID3Picture(bytes.NewReader(input)); isFound {
			t.Fatalf("%s: expected no picture, got %q", caseName, extracted)
		}
	}
}

func buildFLACPictureBlock(pictureType uint32, imageBytes []byte, isLast bool) []byte {
	blockBody := make([]byte, 0)
	appendUint32 := func(value uint32) {
		encoded := make([]byte, 4)
		binary.BigEndian.PutUint32(encoded, value)
		blockBody = append(blockBody, encoded...)
	}
	appendUint32(pictureType)
	appendUint32(uint32(len("image/png")))
	blockBody = append(blockBody, []byte("image/png")...)
	appendUint32(0)
	for i := 0; i < 4; i++ {
		appendUint32(0)
	}
	appendUint32(uint32(len(imageBytes)))
	blockBody = append(blockBody, imageBytes...)

	blockType := byte(flacPictureBlockType)
	if isLast {
		blockType |= flacLastBlockFlag
	}
	header := []byte{blockType, byte(len(blockBody) >> 16), byte(len(blockBody) >> 8), byte(len(blockBody))}
	return append(header, blockBody...)
}

func TestExtractFLACPictureSkipsOtherBlocksAndPrefersFront(t *testing.T) {
	streamInfo := append([]byte{0, 0, 0, 4}, 1, 2, 3, 4)
	flacStream := append([]byte(flacMarker), streamInfo...)
	flacStream = append(flacStream, buildFLACPictureBlock(0, []byte("other"), false)...)
	flacStream = append(flacStream, buildFLACPictureBlock(3, []byte("front"), true)...)
	extracted, isFound := extractFLACPicture(bytes.NewReader(flacStream))
	if !isFound || string(extracted) != "front" {
		t.Fatalf("unexpected flac extraction %q %v", extracted, isFound)
	}
}

func TestExtractFLACPictureRejectsInvalidInput(t *testing.T) {
	validBlock := buildFLACPictureBlock(3, []byte("front"), true)
	invalidInputs := map[string][]byte{
		"wrong marker":     []byte("OggS...."),
		"no blocks":        []byte(flacMarker),
		"truncated body":   append([]byte(flacMarker), validBlock[:12]...),
		"truncated length": append([]byte(flacMarker), validBlock[:4+10]...),
		"data overflow": append([]byte(flacMarker), func() []byte {
			block := append([]byte{}, validBlock...)
			binary.BigEndian.PutUint32(block[len(block)-len("front")-4:], 9999)
			return block
		}()...),
		"skip beyond end": append([]byte(flacMarker), 0, 0, 1, 0),
	}
	for caseName, input := range invalidInputs {
		if extracted, isFound := extractFLACPicture(bytes.NewReader(input)); isFound {
			t.Fatalf("%s: expected no picture, got %q", caseName, extracted)
		}
	}
}

func buildMP4WithCover(coverBytes []byte) []byte {
	buildAtom := func(atomType string, body ...[]byte) []byte {
		joined := bytes.Join(body, nil)
		header := make([]byte, 8)
		binary.BigEndian.PutUint32(header, uint32(8+len(joined)))
		copy(header[4:], atomType)
		return append(header, joined...)
	}
	dataAtom := buildAtom("data", make([]byte, 8), coverBytes)
	ilst := buildAtom("ilst", buildAtom("covr", dataAtom))
	meta := buildAtom("meta", make([]byte, 4), ilst)
	moov := buildAtom("moov", buildAtom("udta", meta))
	return append(buildAtom("ftyp", []byte("M4A \x00\x00\x00\x00")), moov...)
}

func TestExtractEmbeddedCoverDispatchesByContainer(t *testing.T) {
	directory := t.TempDir()
	cases := map[string][]byte{
		"track.mp3":  buildID3v2Tag(t, 3, buildID3Frame(3, "APIC", buildAPICFrameBody(3, []byte("mp3-art")))),
		"track.flac": append([]byte(flacMarker), buildFLACPictureBlock(3, []byte("flac-art"), true)...),
		"track.m4a":  buildMP4WithCover([]byte("mp4-art")),
	}
	expected := map[string]string{"track.mp3": "mp3-art", "track.flac": "flac-art", "track.m4a": "mp4-art"}
	for fileName, content := range cases {
		path := filepath.Join(directory, fileName)
		if err := os.WriteFile(path, content, 0644); err != nil {
			t.Fatal(err)
		}
		extracted, isFound := extractEmbeddedCover(path)
		if !isFound || string(extracted) != expected[fileName] {
			t.Fatalf("%s: unexpected extraction %q %v", fileName, extracted, isFound)
		}
	}

	if _, isFound := extractEmbeddedCover(filepath.Join(directory, "missing.mp3")); isFound {
		t.Fatal("missing file must not yield a cover")
	}
	plainPath := filepath.Join(directory, "plain.wav")
	_ = os.WriteFile(plainPath, []byte("RIFF....WAVEfmt "), 0644)
	if _, isFound := extractEmbeddedCover(plainPath); isFound {
		t.Fatal("unknown container must not yield a cover")
	}
}

func TestExtractMP4PictureRejectsInvalidInput(t *testing.T) {
	for caseName, input := range map[string][]byte{
		"no moov":     []byte("\x00\x00\x00\x08ftyp"),
		"zero size":   {0, 0, 0, 0, 'f', 't', 'y', 'p'},
		"tiny":        {0, 0, 0},
		"empty cover": buildMP4WithCover(nil),
	} {
		if extracted, isFound := extractMP4Picture(bytes.NewReader(input), int64(len(input))); isFound {
			t.Fatalf("%s: expected no picture, got %q", caseName, extracted)
		}
	}
}

func TestFindFolderCoverIsCaseInsensitive(t *testing.T) {
	directory := t.TempDir()
	audioPath := filepath.Join(directory, "song.mp3")
	if _, isFound := findFolderCover(audioPath); isFound {
		t.Fatal("empty folder must not yield a cover")
	}
	_ = os.WriteFile(filepath.Join(directory, "FRONT.JPG"), []byte("front"), 0644)
	extracted, isFound := findFolderCover(audioPath)
	if !isFound || string(extracted) != "front" {
		t.Fatalf("unexpected folder cover %q %v", extracted, isFound)
	}
	_ = os.WriteFile(filepath.Join(directory, "Cover.jpg"), []byte("cover"), 0644)
	extracted, _ = findFolderCover(audioPath)
	if string(extracted) != "cover" {
		t.Fatalf("cover.jpg must outrank front.jpg, got %q", extracted)
	}
	if _, isFound := findFolderCover(filepath.Join(directory, "nope", "song.mp3")); isFound {
		t.Fatal("unreadable folder must not yield a cover")
	}
}

type coverTrackSourceFake struct {
	tracksById map[int]files.FileDto
}

func (fake *coverTrackSourceFake) GetFileById(id int) (files.FileDto, error) {
	track, isKnown := fake.tracksById[id]
	if !isKnown {
		return files.FileDto{}, errors.New("unknown track")
	}
	return track, nil
}

type coverAlbumTracksFake struct {
	trackIDs []int
	err      error
}

func (fake *coverAlbumTracksFake) GetLibraryTrackIDsByAlbum(albumKey string, page int, pageSize int) (utils.PaginationResponse[int], error) {
	return utils.PaginationResponse[int]{Items: fake.trackIDs}, fake.err
}

func writeTrackWithEmbeddedPNG(t *testing.T, directory string, fileName string, fileID int) files.FileDto {
	t.Helper()
	path := filepath.Join(directory, fileName)
	tag := buildID3v2Tag(t, 3, buildID3Frame(3, "APIC", buildAPICFrameBody(3, buildTestPNG(t, 600, 400))))
	if err := os.WriteFile(path, tag, 0644); err != nil {
		t.Fatal(err)
	}
	return files.FileDto{ID: fileID, Path: path, UpdatedAt: time.Unix(1700000000, 0)}
}

func newCoverTestSetup(t *testing.T) (*CoverService, string, *coverTrackSourceFake) {
	t.Helper()
	mediaDir := t.TempDir()
	trackSource := &coverTrackSourceFake{tracksById: map[int]files.FileDto{}}
	embeddedTrack := writeTrackWithEmbeddedPNG(t, mediaDir, "with-art.mp3", 1)
	trackSource.tracksById[1] = embeddedTrack

	bareDir := t.TempDir()
	barePath := filepath.Join(bareDir, "bare.mp3")
	_ = os.WriteFile(barePath, []byte("no tag"), 0644)
	trackSource.tracksById[2] = files.FileDto{ID: 2, Path: barePath, UpdatedAt: time.Unix(1700000000, 0)}

	service := NewCoverService(trackSource, &coverAlbumTracksFake{trackIDs: []int{2, 1, 99}}, t.TempDir())
	return service, bareDir, trackSource
}

func TestCoverServiceRendersResizedJPEGAndCachesIt(t *testing.T) {
	service, _, _ := newCoverTestSetup(t)

	cover, err := service.GetTrackCover(1, 96)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := jpeg.Decode(bytes.NewReader(cover.Data))
	if err != nil {
		t.Fatalf("cover must be a JPEG: %v", err)
	}
	if decoded.Bounds().Dx() != 96 || decoded.Bounds().Dy() != 64 {
		t.Fatalf("expected aspect-preserving 96x64, got %v", decoded.Bounds())
	}
	if cover.ETag != `"cover-1-96-1700000000000000000"` {
		t.Fatalf("unexpected etag %s", cover.ETag)
	}

	cachedEntries, _ := os.ReadDir(service.cacheDir)
	if len(cachedEntries) != 1 {
		t.Fatalf("expected one cached rendition, got %d", len(cachedEntries))
	}
	cachedCover, err := service.GetTrackCover(1, 96)
	if err != nil || !bytes.Equal(cachedCover.Data, cover.Data) {
		t.Fatal("second request must be served from cache")
	}
}

func TestCoverServiceFallsBackToFolderImage(t *testing.T) {
	service, bareDir, _ := newCoverTestSetup(t)
	if _, err := service.GetTrackCover(2, 0); !errors.Is(err, ErrCoverNotFound) {
		t.Fatalf("expected ErrCoverNotFound, got %v", err)
	}
	_ = os.WriteFile(filepath.Join(bareDir, "Folder.png"), buildTestPNG(t, 50, 50), 0644)
	service.cacheDir = t.TempDir()
	cover, err := service.GetTrackCover(2, 0)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(cover.Data, []byte{0xff, 0xd8}) {
		t.Fatal("folder cover must be re-encoded as JPEG")
	}
}

func TestCoverServiceRemembersMissingCover(t *testing.T) {
	service, bareDir, _ := newCoverTestSetup(t)
	if _, err := service.GetTrackCover(2, 64); !errors.Is(err, ErrCoverNotFound) {
		t.Fatalf("expected ErrCoverNotFound, got %v", err)
	}
	_ = os.WriteFile(filepath.Join(bareDir, "cover.png"), buildTestPNG(t, 40, 40), 0644)
	if _, err := service.GetTrackCover(2, 64); !errors.Is(err, ErrCoverNotFound) {
		t.Fatal("fresh no-cover marker must short-circuit resolution")
	}
	markerPaths, _ := filepath.Glob(filepath.Join(service.cacheDir, "*.none"))
	if len(markerPaths) != 1 {
		t.Fatalf("expected one marker, got %d", len(markerPaths))
	}
	staleTime := time.Now().Add(-time.Hour)
	_ = os.Chtimes(markerPaths[0], staleTime, staleTime)
	if _, err := service.GetTrackCover(2, 64); err != nil {
		t.Fatalf("stale marker must allow re-resolution: %v", err)
	}
}

func TestCoverServiceRejectsUndecodableCoverBytes(t *testing.T) {
	mediaDir := t.TempDir()
	path := filepath.Join(mediaDir, "bad.mp3")
	tag := buildID3v2Tag(t, 3, buildID3Frame(3, "APIC", buildAPICFrameBody(3, []byte("not-an-image"))))
	_ = os.WriteFile(path, tag, 0644)
	_ = os.WriteFile(filepath.Join(mediaDir, "cover.jpg"), []byte("also-not-an-image"), 0644)
	service := NewCoverService(&coverTrackSourceFake{tracksById: map[int]files.FileDto{7: {ID: 7, Path: path}}}, &coverAlbumTracksFake{}, t.TempDir())
	if _, err := service.GetTrackCover(7, 0); !errors.Is(err, ErrCoverNotFound) {
		t.Fatalf("expected ErrCoverNotFound, got %v", err)
	}
}

func TestCoverServiceAlbumUsesFirstTrackWithCover(t *testing.T) {
	service, _, _ := newCoverTestSetup(t)
	cover, err := service.GetAlbumCover("album-key", 128)
	if err != nil {
		t.Fatal(err)
	}
	if cover.ETag != `"cover-1-128-1700000000000000000"` {
		t.Fatalf("album cover must come from track 1, got %s", cover.ETag)
	}

	emptyAlbum := NewCoverService(&coverTrackSourceFake{}, &coverAlbumTracksFake{trackIDs: []int{5}}, t.TempDir())
	if _, err := emptyAlbum.GetAlbumCover("k", 0); !errors.Is(err, ErrCoverNotFound) {
		t.Fatalf("expected ErrCoverNotFound, got %v", err)
	}
	failingAlbum := NewCoverService(&coverTrackSourceFake{}, &coverAlbumTracksFake{err: errors.New("db")}, t.TempDir())
	if _, err := failingAlbum.GetAlbumCover("k", 0); err == nil {
		t.Fatal("repository failure must surface")
	}
	if _, err := service.GetTrackCover(404, 0); err == nil {
		t.Fatal("unknown track must fail")
	}
}

func TestNormalizeCoverSize(t *testing.T) {
	for requested, expected := range map[int]int{-5: defaultCoverSize, 0: defaultCoverSize, 1: minimumCoverSize, 96: 96, 99999: maximumCoverSize} {
		if NormalizeCoverSize(requested) != expected {
			t.Fatalf("size %d expected %d got %d", requested, expected, NormalizeCoverSize(requested))
		}
	}
}

type coverServiceFake struct {
	trackCover Cover
	err        error
}

func (fake *coverServiceFake) GetTrackCover(fileID int, size int) (Cover, error) {
	return fake.trackCover, fake.err
}
func (fake *coverServiceFake) GetAlbumCover(albumKey string, size int) (Cover, error) {
	return fake.trackCover, fake.err
}

func performCoverRequest(handler *CoverHandler, target string, ifNoneMatch string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/music/tracks/:file_id/cover", handler.GetTrackCoverHandler)
	router.GET("/music/library/albums/:key/cover", handler.GetAlbumCoverHandler)
	request := httptest.NewRequest(http.MethodGet, target, nil)
	if ifNoneMatch != "" {
		request.Header.Set("If-None-Match", ifNoneMatch)
	}
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}

func TestCoverHandlerServesJPEGWithETagAndHonoursIfNoneMatch(t *testing.T) {
	handler := NewCoverHandler(&coverServiceFake{trackCover: Cover{Data: []byte{0xff, 0xd8, 1}, ETag: `"cover-1"`}}, &musicLoggerMock{})
	for _, target := range []string{"/music/tracks/1/cover?size=96", "/music/library/albums/abc/cover?size=96"} {
		recorder := performCoverRequest(handler, target, "")
		if recorder.Code != http.StatusOK || recorder.Header().Get("Content-Type") != "image/jpeg" {
			t.Fatalf("%s: unexpected response %d %s", target, recorder.Code, recorder.Header().Get("Content-Type"))
		}
		if recorder.Header().Get("ETag") != `"cover-1"` || recorder.Header().Get("Cache-Control") != "public, max-age=3600" {
			t.Fatalf("%s: missing cache headers %v", target, recorder.Header())
		}
		revalidated := performCoverRequest(handler, target, `"cover-1"`)
		if revalidated.Code != http.StatusNotModified || revalidated.Body.Len() != 0 {
			t.Fatalf("%s: expected empty 304, got %d", target, revalidated.Code)
		}
	}
}

func TestCoverHandlerErrors(t *testing.T) {
	missing := NewCoverHandler(&coverServiceFake{err: ErrCoverNotFound}, &musicLoggerMock{})
	for _, target := range []string{"/music/tracks/1/cover", "/music/library/albums/abc/cover"} {
		if code := performCoverRequest(missing, target, "").Code; code != http.StatusNotFound {
			t.Fatalf("%s: expected 404, got %d", target, code)
		}
	}
	failing := NewCoverHandler(&coverServiceFake{err: errors.New("boom")}, &musicLoggerMock{})
	if code := performCoverRequest(failing, "/music/tracks/1/cover", "").Code; code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", code)
	}
	for _, target := range []string{"/music/tracks/abc/cover", "/music/tracks/1/cover?size=x", "/music/library/albums/k/cover?size=x"} {
		if code := performCoverRequest(missing, target, "").Code; code != http.StatusBadRequest {
			t.Fatalf("%s: expected 400, got %d", target, code)
		}
	}
}

func TestExtractID3PictureHandlesUnsynchronisationAndExtendedHeader(t *testing.T) {
	pictureBytes := []byte{0xff, 0xd8, 0xff, 0xe0, 'j'}
	frame := buildID3Frame(3, "APIC", buildAPICFrameBody(3, pictureBytes))
	unsynchronisedFrame := bytes.ReplaceAll(frame, []byte{0xff}, []byte{0xff, 0x00})
	extendedHeader := []byte{0, 0, 0, 6, 0, 0, 0, 0, 0, 0}
	tagBody := append(extendedHeader, unsynchronisedFrame...)
	tag := append([]byte("ID3"), 3, 0, id3UnsynchronisationOn|id3ExtendedHeaderOn)
	tag = append(tag, synchsafeBytes(len(tagBody))...)
	tag = append(tag, tagBody...)
	extracted, isFound := extractID3Picture(bytes.NewReader(tag))
	if !isFound || !bytes.Equal(extracted, pictureBytes) {
		t.Fatalf("unexpected extraction %v %v", extracted, isFound)
	}

	v24Extended := append([]byte{}, synchsafeBytes(6)...)
	v24Extended = append(v24Extended, 1, 0)
	if !bytes.Equal(skipExtendedHeader(append(v24Extended, 'x'), 4), []byte{'x'}) {
		t.Fatal("v2.4 extended header must be skipped")
	}
	if len(skipExtendedHeader([]byte{1}, 3)) != 1 || len(skipExtendedHeader([]byte{0xff, 0xff, 0xff, 0xff, 1}, 3)) != 5 {
		t.Fatal("implausible extended header must be left untouched")
	}
}
