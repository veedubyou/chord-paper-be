package songentity

import (
	"github.com/google/uuid"
	"github.com/veedubyou/chord-paper-be/go-rewrite/src/lib/jsonlib"
	"time"
)

type Song struct {
	jsonlib.Flatten[SongFields]
}

func (s Song) IsNew() bool {
	return s.Defined.ID == ""
}

func (s *Song) CreateID() {
	if !s.IsNew() {
		panic("CreateID is called without an IsNew check")
	}

	s.Defined.ID = uuid.New().String()
}

func (s *Song) SetSavedTime() {
	// truncate to seconds because this will be consumed by the browser
	// and browser dates have only millisecond resolution

	now := time.Now().UTC().Truncate(time.Second)
	s.Defined.LastSavedAt = &now
}

type SongFields struct {
	ID          string     `json:"id"`
	Owner       string     `json:"owner"`
	LastSavedAt *time.Time `json:"lastSavedAt"`
}
