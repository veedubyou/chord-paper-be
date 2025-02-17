package track_split_test

import (
	"embed"
	"fmt"
	"github.com/fsouza/fake-gcs-server/fakestorage"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	server_app "github.com/veedubyou/chord-paper-be/src/server/application"
	"github.com/veedubyou/chord-paper-be/src/shared/config"
	. "github.com/veedubyou/chord-paper-be/src/shared/testing"
	worker_app "github.com/veedubyou/chord-paper-be/src/worker/application"
	"io"
	"net/http"
	"time"
)

//go:embed original_song.mp3
var originalSongMP3 embed.FS

var _ = Describe("TrackSplit", func() {
	var (
		server          server_app.App
		worker          worker_app.App
		cloudStorage    *fakestorage.Server
		originalSongURL string
		notSongURL      string
	)

	ServerHealthCheck := func() (int, error) {
		response, err := RequestFactory{
			Method:  "GET",
			Target:  ServerEndpoint("/health-check"),
			JSONObj: nil,
			Mods:    nil,
		}.Do()

		if err != nil {
			return 0, err
		}

		return response.StatusCode, nil
	}

	GetFileURL := func(bucket string, fileName string) string {
		return fmt.Sprintf("%s/%s/%s", cloudStorage.PublicURL(), bucket, fileName)
	}

	ExpectFileExists := func(fileURL string) {
		response := ExpectSuccess(http.Get(fileURL))
		Expect(response.StatusCode).To(Equal(http.StatusOK))
		bodyBytes := ExpectSuccess(io.ReadAll(response.Body))
		Expect(bodyBytes).NotTo(BeEmpty())
	}

	GetTrackList := func(songID string) map[string]any {
		factory := RequestFactory{
			Method: "GET",
			Target: ServerEndpoint(fmt.Sprintf("/songs/%s/tracklist", songID)),
		}

		response := ExpectSuccess(factory.Do())
		Expect(response.StatusCode).To(Equal(http.StatusOK))
		return DecodeJSON[map[string]any](response.Body)
	}

	GetFirstTrack := func(tracklist map[string]any) map[string]any {
		tracks := ExpectType[[]any](tracklist["tracks"])
		Expect(tracks).NotTo(BeEmpty())
		return ExpectType[map[string]any](tracks[0])
	}

	PutTrackList := func(songID string, tracklist map[string]any) {
		putTracklistResponse := ExpectSuccess(RequestFactory{
			Method:  "PUT",
			Target:  ServerEndpoint(fmt.Sprintf("/songs/%s/tracklist", songID)),
			JSONObj: tracklist,
			Mods:    RequestModifiers{WithUserCred(PrimaryUser)},
		}.Do())

		Expect(putTracklistResponse.StatusCode).To(Equal(http.StatusOK))
		Expect(GetTrackList(songID)).NotTo(BeZero())
	}

	BeforeEach(func() {
		ResetDB(db)
	})

	BeforeEach(func() {
		userBucket := "user-upload"
		songFileName := "original.mp3"
		textFileName := "text.mp3"

		By("Initializing Fake Cloud Storage Server", func() {
			cloudStorage = ExpectSuccess(fakestorage.NewServerWithOptions(fakestorage.Options{
				Scheme:     "http",
				PublicHost: "localhost:4443",
				Host:       "localhost",
				Port:       4443,
				InitialObjects: []fakestorage.Object{
					{
						ObjectAttrs: fakestorage.ObjectAttrs{
							BucketName: userBucket,
							Name:       songFileName,
						},
						Content: ExpectSuccess(originalSongMP3.ReadFile("original_song.mp3")),
					},
					{
						ObjectAttrs: fakestorage.ObjectAttrs{
							BucketName: userBucket,
							Name:       textFileName,
						},
						Content: []byte("some stuff in here"),
					},
				},
			}))

			cloudStorage.CreateBucket(bucketName)
		})

		By("Checking that the original song is in the bucket", func() {
			originalSongURL = GetFileURL(userBucket, songFileName)
			ExpectFileExists(originalSongURL)

			notSongURL = GetFileURL(userBucket, textFileName)
			ExpectFileExists(notSongURL)
		})
	})

	AfterEach(func() {
		cloudStorage.Stop()
	})

	BeforeEach(func() {
		By("Initializing Worker", func() {
			worker = worker_app.NewApp(
				WorkerConfig(region, config.LocalCloudStorage{
					StorageHost:  cloudStorage.PublicURL(),
					HostEndpoint: fmt.Sprintf("%s/storage/v1", cloudStorage.PublicURL()),
					BucketName:   bucketName,
				}),
			)

			go func() {
				defer GinkgoRecover()

				err := worker.Start()
				Expect(err).NotTo(HaveOccurred())
			}()
		})
	})

	AfterEach(func() {
		worker.Stop()
	})

	BeforeEach(func() {
		By("Initializing Server", func() {
			server = server_app.NewApp(ServerConfig(region))

			go func() {
				defer GinkgoRecover()

				err := server.Start()
				Expect(err).NotTo(HaveOccurred())
			}()

			Eventually(ServerHealthCheck).Should(Equal(http.StatusOK))
		})
	})

	AfterEach(func() {
		err := server.Stop()
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("Track splitting", func() {
		var (
			songID string
		)

		stemTests := []struct {
			EngineType            string
			SplitType             string
			ExpectedCompletedType string
			ExpectedNumberOfStems int
		}{
			{
				SplitType:             "split_2stems",
				EngineType:            "spleeter",
				ExpectedCompletedType: "2stems",
				ExpectedNumberOfStems: 2,
			},
			{
				SplitType:             "split_4stems",
				EngineType:            "spleeter",
				ExpectedCompletedType: "4stems",
				ExpectedNumberOfStems: 4,
			},
			{
				SplitType:             "split_5stems",
				EngineType:            "spleeter",
				ExpectedCompletedType: "5stems",
				ExpectedNumberOfStems: 5,
			},
			//{
			//	SplitType:             "split_2stems",
			//	EngineType:            "demucs",
			//	ExpectedCompletedType: "2stems",
			//	ExpectedNumberOfStems: 2,
			//},
			{
				SplitType:             "split_4stems",
				EngineType:            "demucs",
				ExpectedCompletedType: "4stems",
				ExpectedNumberOfStems: 4,
			},
		}

		BeforeEach(func() {
			By("first creating a song", func() {
				demoSong := LoadDemoSong()

				response := ExpectSuccess(RequestFactory{
					Method:  "POST",
					Target:  ServerEndpoint("/songs"),
					JSONObj: demoSong,
					Mods:    RequestModifiers{WithUserCred(PrimaryUser)},
				}.Do())

				Expect(response.StatusCode).To(Equal(http.StatusOK))
				song := DecodeJSON[map[string]any](response.Body)
				songID = ExpectType[string](song["id"])
				Expect(songID).NotTo(BeEmpty())
			})
		})

		for _, test := range stemTests {
			test := test
			requestType := test.SplitType
			engineType := test.EngineType

			Describe(fmt.Sprintf("A valid split request for %s type", requestType), func() {
				BeforeEach(func() {
					By("Putting a tracklist with a split request", func() {
						splitTracklist := map[string]any{
							"song_id": songID,
							"tracks": []map[string]any{
								{
									"id":           "",
									"track_type":   requestType,
									"engine_type":  engineType,
									"label":        "test split",
									"original_url": originalSongURL,
								},
							},
						}
						PutTrackList(songID, splitTracklist)
					})
				})

				It("splits the track", func() {
					GetFirstTrackType := func() string {
						tracklist := GetTrackList(songID)
						firstTrack := GetFirstTrack(tracklist)

						status := firstTrack["job_status"]
						if status != nil {
							if ExpectType[string](status) == "error" {
								fmt.Println(firstTrack)
								Fail("Split Track Job has errored")
							}
						}

						return ExpectType[string](firstTrack["track_type"])
					}

					By("detecting that the track type is changed", func() {
						Eventually(GetFirstTrackType, time.Minute).Should(Equal(test.ExpectedCompletedType))
					})

					tracklist := GetTrackList(songID)
					firstTrack := GetFirstTrack(tracklist)
					stemUrls := ExpectType[map[string]any](firstTrack["stem_urls"])

					By("checking the amount of stems", func() {
						Expect(stemUrls).To(HaveLen(test.ExpectedNumberOfStems))
					})

					By("verifying each stem URL points to a file", func() {
						for _, urlIface := range stemUrls {
							url := ExpectType[string](urlIface)
							ExpectFileExists(url)
						}
					})

				})
			})

			Describe(fmt.Sprintf("An absent original URL split request for %s type", requestType), func() {
				BeforeEach(func() {
					By("Putting a tracklist with a split request", func() {
						splitTracklist := map[string]any{
							"song_id": songID,
							"tracks": []map[string]any{
								{
									"id":           "",
									"track_type":   requestType,
									"engine_type":  engineType,
									"label":        "test split",
									"original_url": GetFileURL(bucketName, "no.mp3"),
								},
							},
						}

						PutTrackList(songID, splitTracklist)
					})
				})

				It("sets an error", func() {
					GetFirstTrackStatus := func() any {
						tracklist := GetTrackList(songID)
						firstTrack := GetFirstTrack(tracklist)

						return firstTrack["job_status"]
					}

					Eventually(GetFirstTrackStatus, time.Minute).Should(Equal("error"))
				})
			})

			Describe(fmt.Sprintf("A not mp3 file for split request for %s type", requestType), func() {
				BeforeEach(func() {
					By("Putting a tracklist with a split request", func() {
						splitTracklist := map[string]any{
							"song_id": songID,
							"tracks": []map[string]any{
								{
									"id":           "",
									"track_type":   requestType,
									"engine_type":  engineType,
									"label":        "test split",
									"original_url": notSongURL,
								},
							},
						}

						PutTrackList(songID, splitTracklist)
					})
				})

				It("sets an error", func() {
					GetFirstTrackStatus := func() any {
						tracklist := GetTrackList(songID)
						firstTrack := GetFirstTrack(tracklist)

						return firstTrack["job_status"]
					}

					Eventually(GetFirstTrackStatus, time.Minute).Should(Equal("error"))
				})
			})
		}
	})
})
