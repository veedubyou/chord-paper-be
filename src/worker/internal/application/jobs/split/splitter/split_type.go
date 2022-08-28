package splitter

import (
	trackentity "github.com/veedubyou/chord-paper-be/src/shared/track/entity"
	"github.com/veedubyou/chord-paper-be/src/worker/internal/lib/cerr"
)

type SplitType string

const (
	InvalidSplitType   SplitType = ""
	SplitTwoStemsType  SplitType = "2stems"
	SplitFourStemsType SplitType = "4stems"
	SplitFiveStemsType SplitType = "5stems"
)

func ConvertToSplitType(trackType trackentity.SplitRequestType) (SplitType, error) {
	switch trackType {
	case trackentity.SplitTwoStemsType:
		return SplitTwoStemsType, nil
	case trackentity.SplitFourStemsType:
		return SplitFourStemsType, nil
	case trackentity.SplitFiveStemsType:
		return SplitFiveStemsType, nil
	default:
		return InvalidSplitType,
			cerr.Field("track_type", trackType).Error("Value does not match any split type")
	}
}
