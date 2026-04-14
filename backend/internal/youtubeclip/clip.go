package youtubeclip

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"strings"
)

var (
	// ErrNotYouTube means the URL is not an allowed YouTube watch / shorts / youtu.be link.
	ErrNotYouTube = errors.New("not a supported YouTube URL")
	// ErrClip covers yt-dlp / ffmpeg failures or empty output.
	ErrClip = errors.New("youtube audio clip failed")
)

// Extractor runs yt-dlp and ffmpeg. Both binaries must be on PATH or set via env.
type Extractor struct {
	YtDlp  string
	Ffmpeg string
}

// NewFromEnv builds paths from YT_DLP_PATH and FFMPEG_PATH (defaults: yt-dlp, ffmpeg).
func NewFromEnv() *Extractor {
	y := os.Getenv("YT_DLP_PATH")
	if y == "" {
		y = "yt-dlp"
	}
	f := os.Getenv("FFMPEG_PATH")
	if f == "" {
		f = "ffmpeg"
	}
	return &Extractor{YtDlp: y, Ffmpeg: f}
}

// ValidateYouTubePageURL rejects non-HTTP(S) and hosts other than YouTube watch URLs.
func ValidateYouTubePageURL(raw string) error {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Host == "" {
		return ErrNotYouTube
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return ErrNotYouTube
	}
	host := strings.ToLower(u.Hostname())
	switch host {
	case "www.youtube.com", "youtube.com", "m.youtube.com", "music.youtube.com":
		if strings.HasPrefix(u.Path, "/watch") || strings.HasPrefix(u.Path, "/shorts/") {
			return nil
		}
	case "youtu.be":
		if len(u.Path) > 1 {
			return nil
		}
	}
	return ErrNotYouTube
}

// FirstSecondsMP3 streams best audio from a YouTube page URL and returns the first seconds as MP3 bytes.
func (e *Extractor) FirstSecondsMP3(parent context.Context, youTubePageURL string, seconds int) ([]byte, error) {
	if seconds < 1 {
		return nil, fmt.Errorf("%w: invalid duration", ErrClip)
	}
	if err := ValidateYouTubePageURL(youTubePageURL); err != nil {
		return nil, err
	}

	ctx, stop := context.WithCancel(parent)
	defer stop()

	pr, pw := io.Pipe()
	yt := exec.CommandContext(ctx, e.YtDlp,
		"-f", "bestaudio/best",
		"--no-playlist",
		"-o", "-",
		youTubePageURL,
	)
	yt.Stdout = pw
	yt.Stderr = io.Discard

	if err := yt.Start(); err != nil {
		_ = pw.Close()
		return nil, fmt.Errorf("%w: yt-dlp start: %v", ErrClip, err)
	}

	var buf bytes.Buffer
	ff := exec.CommandContext(ctx, e.Ffmpeg,
		"-hide_banner", "-loglevel", "error",
		"-i", "pipe:0",
		"-t", fmt.Sprintf("%d", seconds),
		"-vn",
		"-acodec", "libmp3lame",
		"-f", "mp3",
		"pipe:1",
	)
	ff.Stdin = pr
	ff.Stdout = &buf
	ff.Stderr = io.Discard

	ffErr := ff.Run()
	stop()
	_ = pw.Close()
	_ = pr.Close()
	_ = yt.Wait()

	if ffErr != nil {
		return nil, fmt.Errorf("%w: ffmpeg: %v", ErrClip, ffErr)
	}
	if buf.Len() == 0 {
		return nil, fmt.Errorf("%w: empty mp3 output (is yt-dlp installed and up to date?)", ErrClip)
	}
	return buf.Bytes(), nil
}
