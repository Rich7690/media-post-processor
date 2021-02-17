package transcode

import (
	"media-web/internal/utils"
	"path"

	"github.com/pkg/errors"
)

var (
	ErrFileNotExists = errors.New("input file does not exist")
)

type VideoFile interface {
	GetFilePath() string
	GetContainerFormat() string
	GetVideoCodec() string
}

type VideoFileImpl struct {
	FilePath        string `json:"filePath,omitempty"`
	ContainerFormat string `json:"containerFormat,omitempty"`
	VideoCodec      string `json:"videoCodec,omitempty"`
}

func (v VideoFileImpl) GetFilePath() string {
	return v.FilePath
}

func (v VideoFileImpl) GetContainerFormat() string {
	return v.ContainerFormat
}

func (v VideoFileImpl) GetVideoCodec() string {
	return v.VideoCodec
}

func ShouldTranscode(input VideoFile) (should bool, reason string, err error) {
	if !utils.FileExists(input.GetFilePath()) {
		return false, "", ErrFileNotExists
	}

	if path.Ext(input.GetFilePath()) != ".mp4" {
		return true, "file does not have .mp4 extension", nil
	}

	if input.GetContainerFormat() != "" && input.GetContainerFormat() != "MPEG-4" {
		return true, "file not in mp4 format", nil
	}

	switch input.GetVideoCodec() {
	case "AVC":
	case "h264":
	case "x264":
	case "": // for now just skip unknown codecs
		return false, "", nil
	default:
		return true, "file has codec " + input.GetVideoCodec(), nil
	}

	return false, "", nil
}
