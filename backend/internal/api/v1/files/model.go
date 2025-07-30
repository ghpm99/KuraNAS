package files

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"io"
	"os"
	"time"
)

type FileModel struct {
	ID              int
	Name            string
	Path            string
	ParentPath      string
	Type            FileType
	Format          string
	Size            int64
	UpdatedAt       time.Time
	CreatedAt       time.Time
	DeletedAt       sql.NullTime
	LastInteraction sql.NullTime
	LastBackup      sql.NullTime
	CheckSum        string
	Starred         bool
}

type RecentFileModel struct {
	ID         int
	IPAddress  string
	FileID     int
	AccessedAt time.Time
}

func (i *FileDto) ToModel() (FileModel, error) {

	fileModel := FileModel{
		ID:         i.ID,
		Name:       i.Name,
		Path:       i.Path,
		ParentPath: i.ParentPath,
		Type:       i.Type,
		Format:     i.Format,
		Size:       i.Size,
		UpdatedAt:  i.UpdatedAt,
		CreatedAt:  i.CreatedAt,
		CheckSum:   i.CheckSum,
		Starred:    i.Starred,
	}

	deletedAt, err := i.DeletedAt.ParseToNullTime()
	if err != nil {
		return fileModel, err
	}
	fileModel.DeletedAt = deletedAt

	lastInteraction, err := i.LastInteraction.ParseToNullTime()
	if err != nil {
		return fileModel, err
	}
	fileModel.LastInteraction = lastInteraction

	lastBackup, err := i.LastBackup.ParseToNullTime()

	if err != nil {
		return fileModel, err
	}
	fileModel.LastBackup = lastBackup

	return fileModel, nil
}

func (fileModel *FileModel) GetCheckSumFromFile() error {
	file, err := os.Open(fileModel.Path)

	if err != nil {
		return err
	}

	defer file.Close()

	h := sha256.New()

	if _, err := io.Copy(h, file); err != nil {
		return err
	}

	checkSumBytes := h.Sum(nil)
	checkSumString := fmt.Sprintf("%x", checkSumBytes)

	fmt.Printf("Check sum %s, tamanho %d\n", checkSumString, len(checkSumString))

	return nil
}

type SizeReportModel struct {
	Format string
	Total  int
	Size   int64
}

type DuplicateFilesModel struct {
	Name   string
	Size   int64
	Copies int
	Paths  string
}

type ImageMetadataModel struct {
	ID           int
	FileId       int
	Path         string
	Format       string  `json:"format"`
	Mode         string  `json:"mode"`
	Width        int     `json:"width"`
	Height       int     `json:"height"`
	CaptureDate  string  `json:"capture_date"`
	Software     string  `json:"software"`
	Make         string  `json:"make"`
	Model        string  `json:"model"`
	LensModel    string  `json:"lens_model"`
	ISO          int     `json:"iso"`
	ExposureTime string  `json:"exposure_time"`
	DPIX         float64 `json:"dpi_x"`
	DPIY         float64 `json:"dpi_y"`
	ICCProfile   string  `json:"icc_profile"`
	GPSLatitude  float64 `json:"gps_latitude"`
	GPSLongitude float64 `json:"gps_longitude"`
	CreatedAt    time.Time
}

type AudioMetadataModel struct {
	ID                  int
	FileId              int
	Path                string
	Mime                string  `json:"mime"`
	Length              float64 `json:"length"`
	Bitrate             int     `json:"bitrate"`
	SampleRate          int     `json:"sample_rate"`
	Channels            int     `json:"channels"`
	BitrateMode         int     `json:"bitrate_mode"`
	EncoderInfo         string  `json:"encoder_info"`
	BitDepth            int     `json:"bit_depth"`
	Title               string  `json:"title"`
	Artist              string  `json:"artist"`
	Album               string  `json:"album"`
	AlbumArtist         string  `json:"album_artist"`
	TrackNumber         string  `json:"track_number"`
	Genre               string  `json:"genre"`
	Composer            string  `json:"composer"`
	Year                string  `json:"year"`
	RecordingDate       string  `json:"recording_date"`
	Encoder             string  `json:"encoder"`
	Publisher           string  `json:"publisher"`
	OriginalReleaseDate string  `json:"original_release_date"`
	OriginalArtist      string  `json:"original_artist"`
	Lyricist            string  `json:"lyricist"`
	Lyrics              string  `json:"lyrics"`
	CreatedAt           time.Time
}

type VideoMetadataModel struct {
	ID              int
	FileId          int
	Path            string
	FormatName      string  `json:"format_name"`
	Size            string  `json:"size"`
	Duration        string  `json:"duration"`
	Width           int     `json:"width"`
	Height          int     `json:"height"`
	FrameRate       float64 `json:"frame_rate"`
	NbFrames        int     `json:"nb_frames"`
	BitRate         string  `json:"bit_rate"`
	CodecName       string  `json:"codec_name"`
	CodecLongName   string  `json:"codec_long_name"`
	PixFmt          string  `json:"pix_fmt"`
	Level           int     `json:"level"`
	Profile         string  `json:"profile"`
	AspectRatio     string  `json:"aspect_ratio"`
	AudioCodec      string  `json:"audio_codec"`
	AudioChannels   int     `json:"audio_channels"`
	AudioSampleRate string  `json:"audio_sample_rate"`
	AudioBitRate    string  `json:"audio_bit_rate"`
	CreatedAt       time.Time
}
