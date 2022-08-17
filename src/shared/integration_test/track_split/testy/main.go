package main

import (
	"embed"
	"fmt"
	"github.com/fsouza/fake-gcs-server/fakestorage"
	"time"
)

//go:embed original_song.mp3
var originalSongMP3 embed.FS

func ExpectSuccess[T any](t T, err error) T {
	if err != nil {
		panic(err)
	}
	return t
}

func main() {
	cloudStorage := ExpectSuccess(fakestorage.NewServerWithOptions(fakestorage.Options{
		Scheme:     "http",
		PublicHost: "127.0.0.1.nip.io:4443",
		Host:       "localhost",
		Port:       4443,
		InitialObjects: []fakestorage.Object{
			{
				ObjectAttrs: fakestorage.ObjectAttrs{
					BucketName: "chord-paper-tracks-test",
					Name:       "original.mp3",
				},
				Content: ExpectSuccess(originalSongMP3.ReadFile("original_song.mp3")),
			},
		},
	}))

	fmt.Println(cloudStorage.URL())
	time.Sleep(time.Hour)
}
