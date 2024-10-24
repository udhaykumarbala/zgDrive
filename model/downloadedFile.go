package model

import (
	"fmt"
	"time"
)

type DownloadedFile struct {
	ID            int64     `json:"id"`
	FileId        int64     `json:"file_id"`
	Filename      string    `json:"filename"`
	Size          int64     `json:"size"`
	SizeReadable  string    `json:"size_readable"`
	IsDownloading bool      `json:"is_downloading"`
	DownloadedAt  time.Time `json:"downloaded_at"`
}

func (f *DownloadedFile) SetSizeReadable() {
	if f.Size < 1024 {
		f.SizeReadable = fmt.Sprintf("%d B", f.Size)
	} else if f.Size < 1024*1024 {
		f.SizeReadable = fmt.Sprintf("%.2f KB", float64(f.Size)/1024)
	} else if f.Size < 1024*1024*1024 {
		f.SizeReadable = fmt.Sprintf("%.2f MB", float64(f.Size)/(1024*1024))
	} else {
		f.SizeReadable = fmt.Sprintf("%.2f GB", float64(f.Size)/(1024*1024*1024))
	}
}
