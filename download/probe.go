package download

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"
)

type FFProbeOutput struct {
	Streams []Streams `json:"streams"`
	Format  Format    `json:"format"`
}
type Disposition struct {
	Default         int `json:"default"`
	Dub             int `json:"dub"`
	Original        int `json:"original"`
	Comment         int `json:"comment"`
	Lyrics          int `json:"lyrics"`
	Karaoke         int `json:"karaoke"`
	Forced          int `json:"forced"`
	HearingImpaired int `json:"hearing_impaired"`
	VisualImpaired  int `json:"visual_impaired"`
	CleanEffects    int `json:"clean_effects"`
	AttachedPic     int `json:"attached_pic"`
	TimedThumbnails int `json:"timed_thumbnails"`
}
type Streams struct {
	Index              int         `json:"index"`
	CodecName          string      `json:"codec_name"`
	CodecLongName      string      `json:"codec_long_name"`
	Profile            string      `json:"profile"`
	CodecType          string      `json:"codec_type"`
	CodecTimeBase      string      `json:"codec_time_base"`
	CodecTagString     string      `json:"codec_tag_string"`
	CodecTag           string      `json:"codec_tag"`
	Width              int         `json:"width,omitempty"`
	Height             int         `json:"height,omitempty"`
	CodedWidth         int         `json:"coded_width,omitempty"`
	CodedHeight        int         `json:"coded_height,omitempty"`
	HasBFrames         int         `json:"has_b_frames,omitempty"`
	SampleAspectRatio  string      `json:"sample_aspect_ratio,omitempty"`
	DisplayAspectRatio string      `json:"display_aspect_ratio,omitempty"`
	PixFmt             string      `json:"pix_fmt,omitempty"`
	Level              int         `json:"level,omitempty"`
	ColorRange         string      `json:"color_range,omitempty"`
	ColorSpace         string      `json:"color_space,omitempty"`
	ColorTransfer      string      `json:"color_transfer,omitempty"`
	ColorPrimaries     string      `json:"color_primaries,omitempty"`
	ChromaLocation     string      `json:"chroma_location,omitempty"`
	Refs               int         `json:"refs,omitempty"`
	IsAvc              string      `json:"is_avc,omitempty"`
	NalLengthSize      string      `json:"nal_length_size,omitempty"`
	RFrameRate         string      `json:"r_frame_rate"`
	AvgFrameRate       string      `json:"avg_frame_rate"`
	TimeBase           string      `json:"time_base"`
	StartPts           int         `json:"start_pts"`
	StartTime          string      `json:"start_time"`
	BitRate            string      `json:"bit_rate"`
	BitsPerRawSample   string      `json:"bits_per_raw_sample,omitempty"`
	Disposition        Disposition `json:"disposition"`
	SampleFmt          string      `json:"sample_fmt,omitempty"`
	SampleRate         string      `json:"sample_rate,omitempty"`
	Channels           int         `json:"channels,omitempty"`
	ChannelLayout      string      `json:"channel_layout,omitempty"`
	BitsPerSample      int         `json:"bits_per_sample,omitempty"`
}
type Format struct {
	Filename       string  `json:"filename"`
	NbStreams      int     `json:"nb_streams"`
	NbPrograms     int     `json:"nb_programs"`
	FormatName     string  `json:"format_name"`
	FormatLongName string  `json:"format_long_name"`
	StartTime      seconds `json:"start_time"`
	Duration       seconds `json:"duration"`
	Size           string  `json:"size"`
	BitRate        string  `json:"bit_rate"`
	ProbeScore     int     `json:"probe_score"`
}

type seconds time.Duration

func (d *seconds) UnmarshalJSON(b []byte) error {
	var f float64
	if len(b) < 2 {
		*d = 0
		return nil
	}
	err := json.Unmarshal(b[1:len(b)-1], &f)
	if err != nil {
		return fmt.Errorf("Can't get duration: %w", err)
	}
	*d = seconds(int64(f * float64(time.Second)))
	return nil
}
func (d seconds) Duration() time.Duration {
	return time.Duration(d)
}

func ProbeStream(ctx context.Context, url string) (*FFProbeOutput, error) {

	params := []string{
		"-v", "quiet",
		"-show_streams",
		"-show_format",
		"-print_format", "json",
		"-i", url,
	}

	out, err := exec.Command("ffprobe", params...).Output()
	if err != nil {
		return nil, fmt.Errorf("Can't probe stream %q: %w", url, err)
	}

	probe := FFProbeOutput{}
	err = json.Unmarshal(out, &probe)
	if err != nil {
		return nil, fmt.Errorf("Can't decode probe result %q: %w", url, err)
	}

	return &probe, nil
}
