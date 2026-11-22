package setup

import "context"

type WriteFileParams struct {
	FilePath string
	Content  string
}

type MkdirAllParams struct {
	DirPath string
}

type FileExistsParams struct {
	FilePath string
}

type FileWriter interface {
	WriteFile(ctx context.Context, params *WriteFileParams) error
	MkdirAll(ctx context.Context, params *MkdirAllParams) error
	FileExists(ctx context.Context, params *FileExistsParams) (bool, error)
}
