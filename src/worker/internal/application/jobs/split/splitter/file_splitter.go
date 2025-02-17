package splitter

import (
	"context"
	trackentity "github.com/veedubyou/chord-paper-be/src/shared/track/entity"
)

type StemFilePaths = map[string]string

type FileSplitter interface {
	SplitFile(ctx context.Context, originalFilePath string, stemOutputDir string, splitType SplitType, engineType trackentity.SplitEngineType) (StemFilePaths, error)
}
