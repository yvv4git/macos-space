package fs

import (
	"context"
	"os"

	setupports "github.com/yvv4git/macos-space/internal/ports/setup"
)

type Writer struct{}

func NewWriter() *Writer {
	return &Writer{}
}

func (w *Writer) WriteFile(_ context.Context, params *setupports.WriteFileParams) error {
	if err := os.WriteFile(params.FilePath, []byte(params.Content), fileMode); err != nil {
		return err
	}

	return os.Chmod(params.FilePath, fileMode)
}

func (w *Writer) MkdirAll(_ context.Context, params *setupports.MkdirAllParams) error {
	if err := os.MkdirAll(params.DirPath, dirMode); err != nil {
		return err
	}

	return os.Chmod(params.DirPath, dirMode)
}

func (w *Writer) FileExists(_ context.Context, params *setupports.FileExistsParams) (bool, error) {
	_, err := os.Stat(params.FilePath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}

	return false, err
}
