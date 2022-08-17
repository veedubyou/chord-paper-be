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

	GetServerURL := func(path string) string {
		return fmt.Sprintf("http://localhost%s%s", ServerPort, path)
	}

	GetTrackList := func(songID string) map[string]interface{} {
		factory := RequestFactory{
			Method: "GET",
			Target: GetServerURL(fmt.Sprintf("/songs/%s/tracklist", songID)),
		}

		response := ExpectSuccess(factory.Do())
		Expect(response.StatusCode).To(Equal(http.StatusOK))
		return DecodeJSON[map[string]interface{}](response.Body)
	}

	BeforeEach(func() {
		ResetDB(db)
	})

	BeforeEach(func() {
		userBucket := "user-upload"
		songFileName := "original.mp3"

		By("Initializing Fake Cloud Storage Server")
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
			},
		}))

		cloudStorage.CreateBucket(bucketName)

		By("Checking that the original song is in the bucket")
		originalSongURL = GetFileURL(userBucket, songFileName)
		ExpectFileExists(originalSongURL)
	})

	AfterEach(func() {
		cloudStorage.Stop()
	})

	BeforeEach(func() {
		By("Initializing Worker")
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

	AfterEach(func() {
		worker.Stop()
	})

	BeforeEach(func() {
		By("Initializing Server")
		server = server_app.NewApp(ServerConfig(region))

		go func() {
			defer GinkgoRecover()

			err := server.Start()
			Expect(err).NotTo(HaveOccurred())
		}()

		Eventually(ServerHealthCheck).Should(Equal(http.StatusOK))
	})

	AfterEach(func() {
		err := server.Stop()
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("Track splitting", func() {
		var (
			songID string
		)

		stemTests := map[string]struct {
			CompletedType string
			NumberOfStems int
		}{
			"split_2stems": {
				CompletedType: "2stems",
				NumberOfStems: 2,
			},
			"split_4stems": {
				CompletedType: "4stems",
				NumberOfStems: 4,
			},
			"split_5stems": {
				CompletedType: "5stems",
				NumberOfStems: 5,
			},
		}

		BeforeEach(func() {
			By("first creating a song")
			demoSong := LoadDemoSong()

			response := ExpectSuccess(RequestFactory{
				Method:  "POST",
				Target:  ServerEndpoint("/songs"),
				JSONObj: demoSong,
				Mods:    RequestModifiers{WithUserCred(PrimaryUser)},
			}.Do())

			Expect(response.StatusCode).To(Equal(http.StatusOK))
			song := DecodeJSON[map[string]interface{}](response.Body)
			songID = ExpectType[string](song["id"])
			Expect(songID).NotTo(BeEmpty())
		})

		for requestType, expected := range stemTests {
			requestType := requestType
			expected := expected

			Describe(fmt.Sprintf("A valid split request for %s tyoe", requestType), func() {
				var GetFirstTrack = func(tracklist map[string]interface{}) map[string]interface{} {
					tracks := ExpectType[[]interface{}](tracklist["tracks"])
					Expect(tracks).NotTo(BeEmpty())
					return ExpectType[map[string]interface{}](tracks[0])
				}

				BeforeEach(func() {
					By("Putting a tracklist with a split request")
					splitTracklist := map[string]interface{}{
						"song_id": songID,
						"tracks": []map[string]interface{}{
							{
								"id":           "",
								"track_type":   requestType,
								"label":        "test split",
								"original_url": originalSongURL,
							},
						},
					}

					putTracklistResponse := ExpectSuccess(RequestFactory{
						Method:  "PUT",
						Target:  GetServerURL(fmt.Sprintf("/songs/%s/tracklist", songID)),
						JSONObj: splitTracklist,
						Mods:    RequestModifiers{WithUserCred(PrimaryUser)},
					}.Do())

					Expect(putTracklistResponse.StatusCode).To(Equal(http.StatusOK))
					Expect(GetTrackList(songID)).NotTo(BeZero())
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

					By("detecting that the track type is changed")
					Eventually(GetFirstTrackType, time.Minute).Should(Equal(expected.CompletedType))

					By("checking the amount of stems")
					tracklist := GetTrackList(songID)
					firstTrack := GetFirstTrack(tracklist)
					stemUrls := ExpectType[map[string]interface{}](firstTrack["stem_urls"])
					Expect(stemUrls).To(HaveLen(expected.NumberOfStems))

					By("verifying each stem URL points to a file")
					for _, urlIface := range stemUrls {
						url := ExpectType[string](urlIface)
						ExpectFileExists(url)
					}
				})
			})
		}
	})
})
