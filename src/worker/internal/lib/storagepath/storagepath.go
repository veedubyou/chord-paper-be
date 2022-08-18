package storagepath

import "fmt"

type Generator struct {
	Host   string
	Bucket string
}

func (g Generator) GeneratePath(tracklistID string, trackID string, leafPath string) string {
	return fmt.Sprintf("%s/%s/%s/%s/%s", g.Host, g.Bucket, tracklistID, trackID, leafPath)
}
