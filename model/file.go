package model

import (
	"fmt"
	"time"
)

type File struct {
	ID           int64     `json:"id"`
	Filename     string    `json:"filename"`
	Hash         string    `json:"hash"`
	Size         int64     `json:"size"`
	SizeReadable string    `json:"size_readable"`
	TxId         string    `json:"tx_id"`
	IsUploaded   bool      `json:"is_uploaded"`
	CreatedAt    time.Time `json:"created_at"`
}

func (f *File) SetSizeReadable() {
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
