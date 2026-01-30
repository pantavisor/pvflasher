package flash

import "time"

type Progress struct {
	Phase          string  `json:"phase"`
	BytesProcessed int64   `json:"processed"`
	BytesTotal     int64   `json:"total"`
	Percentage     float64 `json:"percentage"`
	Speed          float64 `json:"speed"`
}

type ProgressCallback func(Progress)

type FlashResult struct {
	BytesWritten     int64         `json:"bytes_written"`
	BlocksWritten    int64         `json:"blocks_written"`
	Duration         time.Duration `json:"duration"`
	AverageSpeed     float64       `json:"average_speed"`
	UsedBmap         bool          `json:"used_bmap"`
	VerificationDone bool          `json:"verification_done"`
	DeviceEjected    bool          `json:"device_ejected"`
}

type Options struct {
	ImagePath  string
	DevicePath string
	BmapPath   string // Optional
	NoVerify   bool
	NoEject    bool // Don't eject device after flash
	Force      bool // Allow writing to mounted devices
	ProgressCb ProgressCallback
}
