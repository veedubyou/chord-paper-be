package jsonlib_test

import (
	"encoding/json"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/veedubyou/chord-paper-be/src/shared/lib/jsonlib"
	. "github.com/veedubyou/chord-paper-be/src/shared/testing"
)

type user struct {
	ID   string         `json:"identifier"`
	Name string         `json:"name"`
	Map  map[string]any `json:"map"`
}

func newFlattenUser(u user, m map[string]any) jsonlib.Flatten[user] {
	return jsonlib.Flatten[user]{
		Defined: u,
		Extra:   m,
	}
}

var _ = Describe("Flatten", func() {
	var (
		flatUser    jsonlib.Flatten[user]
		mapContents map[string]any
	)

	ItTransformsToMap := func() {
		It("transforms to map correctly", Offset(1), func() {
			toMap := ExpectSuccess(flatUser.ToMap())
			Expect(toMap).To(Equal(mapContents))
		})
	}

	ItMarshals := func() {
		It("marshals correctly", Offset(1), func() {
			flattenJSON := ExpectSuccess(flatUser.MarshalJSON())
			expectedJSON := ExpectSuccess(json.Marshal(mapContents))
			Expect(flattenJSON).To(Equal(expectedJSON))
		})
	}

	ItTransformsFromMap := func() {
		It("transforms from map correctly", Offset(1), func() {
			actual := jsonlib.Flatten[user]{}
			err := actual.FromMap(mapContents)
			Expect(err).NotTo(HaveOccurred())

			Expect(actual).To(Equal(flatUser))
		})
	}

	ItUnmarshals := func() {
		It("unmarshals correctly", Offset(1), func() {
			jsonContents := ExpectSuccess(json.Marshal(mapContents))

			actual := jsonlib.Flatten[user]{}
			err := actual.UnmarshalJSON(jsonContents)
			Expect(err).NotTo(HaveOccurred())

			Expect(actual).To(Equal(flatUser))
		})
	}

	ItFlattens := func() {
		ItTransformsToMap()
		ItMarshals()
		ItTransformsFromMap()
		ItUnmarshals()
	}

	Describe("Empty cases", func() {
		BeforeEach(func() {
			flatUser = newFlattenUser(user{}, map[string]any{})
			mapContents = map[string]any{
				"identifier": "",
				"name":       "",
				"map":        nil,
			}
		})

		ItFlattens()
	})

	Describe("Non-empty defined fields", func() {
		BeforeEach(func() {
			flatUser = newFlattenUser(user{
				ID:   "12-34-56",
				Name: "Chord Paper-san",
				Map: map[string]any{
					"Array":  []any{"3", float64(4)},
					"Number": float64(3),
					"Map": map[string]any{
						"Five": float64(5),
					},
				},
			}, map[string]any{})

			mapContents = map[string]any{
				"identifier": "12-34-56",
				"name":       "Chord Paper-san",
				"map": map[string]any{
					"Array":  []any{"3", float64(4)},
					"Number": float64(3),
					"Map": map[string]any{
						"Five": float64(5),
					},
				},
			}
		})

		ItFlattens()
	})

	Describe("Non-empty extra fields", func() {
		BeforeEach(func() {
			flatUser = newFlattenUser(user{}, map[string]any{
				"id2": "12-34-56",
				"map2": map[string]any{
					"arr2": []any{"b"},
				},
			})
			mapContents = map[string]any{
				"identifier": "",
				"name":       "",
				"map":        nil,
				"id2":        "12-34-56",
				"map2": map[string]any{
					"arr2": []any{"b"},
				},
			}
		})

		ItFlattens()
	})

	Describe("Mix of non-empty defined and extra fields", func() {
		BeforeEach(func() {
			flatUser = newFlattenUser(user{
				ID:   "12-34-56",
				Name: "Chord Paper-san",
				Map: map[string]any{
					"Array":  []any{"3", float64(4)},
					"Number": float64(3),
					"Map": map[string]any{
						"Five": float64(5),
					},
				},
			}, map[string]any{
				"id2": "12-34-56",
				"map2": map[string]any{
					"arr2": []any{"b"},
				},
			})
			mapContents = map[string]any{
				"identifier": "12-34-56",
				"name":       "Chord Paper-san",
				"map": map[string]any{
					"Array":  []any{"3", float64(4)},
					"Number": float64(3),
					"Map": map[string]any{
						"Five": float64(5),
					},
				},
				"id2": "12-34-56",
				"map2": map[string]any{
					"arr2": []any{"b"},
				},
			}
		})

		ItFlattens()
	})

	Describe("Overlapping fields in defined and extra fields", func() {
		BeforeEach(func() {
			flatUser = newFlattenUser(user{
				ID:   "id3",
				Name: "real-name",
				Map:  nil,
			}, map[string]any{
				"name": "extra-name",
			})

			mapContents = map[string]any{
				"identifier": "id3",
				"name":       "real-name",
				"map":        nil,
			}
		})

		ItMarshals()
	})
})
