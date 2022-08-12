package jsonlib_test

import (
	"encoding/json"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/veedubyou/chord-paper-be/src/server/internal/lib/jsonlib"
	. "github.com/veedubyou/chord-paper-be/src/server/internal/lib/testing"
)

type user struct {
	ID   string                 `json:"identifier"`
	Name string                 `json:"name"`
	Map  map[string]interface{} `json:"map"`
}

func newFlattenUser(u user, m map[string]interface{}) jsonlib.Flatten[user] {
	return jsonlib.Flatten[user]{
		Defined: u,
		Extra:   m,
	}
}

var _ = Describe("Flatten", func() {
	var (
		flatUser    jsonlib.Flatten[user]
		mapContents map[string]interface{}
	)

	testToMap := func() {
		It("transforms to map correctly", Offset(1), func() {
			toMap := ExpectSuccess(flatUser.ToMap())
			Expect(toMap).To(Equal(mapContents))
		})
	}

	testMarshal := func() {
		It("marshals correctly", Offset(1), func() {
			flattenJSON := ExpectSuccess(flatUser.MarshalJSON())
			expectedJSON := ExpectSuccess(json.Marshal(mapContents))
			Expect(flattenJSON).To(Equal(expectedJSON))
		})
	}

	testFromMap := func() {
		It("transforms from map correctly", Offset(1), func() {
			actual := jsonlib.Flatten[user]{}
			err := actual.FromMap(mapContents)
			Expect(err).NotTo(HaveOccurred())

			Expect(actual).To(Equal(flatUser))
		})
	}

	testUnmarshal := func() {
		It("unmarshals correctly", Offset(1), func() {
			jsonContents := ExpectSuccess(json.Marshal(mapContents))

			actual := jsonlib.Flatten[user]{}
			err := actual.UnmarshalJSON(jsonContents)
			Expect(err).NotTo(HaveOccurred())

			Expect(actual).To(Equal(flatUser))
		})
	}

	TestFlatten := func() {
		testToMap()
		testMarshal()
		testFromMap()
		testUnmarshal()
	}

	Describe("Empty cases", func() {
		BeforeEach(func() {
			flatUser = newFlattenUser(user{}, map[string]interface{}{})
			mapContents = map[string]interface{}{
				"identifier": "",
				"name":       "",
				"map":        nil,
			}
		})

		TestFlatten()
	})

	Describe("Non-empty defined fields", func() {
		BeforeEach(func() {
			flatUser = newFlattenUser(user{
				ID:   "12-34-56",
				Name: "Chord Paper-san",
				Map: map[string]interface{}{
					"Array":  []interface{}{"3", float64(4)},
					"Number": float64(3),
					"Map": map[string]interface{}{
						"Five": float64(5),
					},
				},
			}, map[string]interface{}{})

			mapContents = map[string]interface{}{
				"identifier": "12-34-56",
				"name":       "Chord Paper-san",
				"map": map[string]interface{}{
					"Array":  []interface{}{"3", float64(4)},
					"Number": float64(3),
					"Map": map[string]interface{}{
						"Five": float64(5),
					},
				},
			}
		})

		TestFlatten()
	})

	Describe("Non-empty extra fields", func() {
		BeforeEach(func() {
			flatUser = newFlattenUser(user{}, map[string]interface{}{
				"id2": "12-34-56",
				"map2": map[string]interface{}{
					"arr2": []interface{}{"b"},
				},
			})
			mapContents = map[string]interface{}{
				"identifier": "",
				"name":       "",
				"map":        nil,
				"id2":        "12-34-56",
				"map2": map[string]interface{}{
					"arr2": []interface{}{"b"},
				},
			}
		})

		TestFlatten()
	})

	Describe("Mix of non-empty defined and extra fields", func() {
		BeforeEach(func() {
			flatUser = newFlattenUser(user{
				ID:   "12-34-56",
				Name: "Chord Paper-san",
				Map: map[string]interface{}{
					"Array":  []interface{}{"3", float64(4)},
					"Number": float64(3),
					"Map": map[string]interface{}{
						"Five": float64(5),
					},
				},
			}, map[string]interface{}{
				"id2": "12-34-56",
				"map2": map[string]interface{}{
					"arr2": []interface{}{"b"},
				},
			})
			mapContents = map[string]interface{}{
				"identifier": "12-34-56",
				"name":       "Chord Paper-san",
				"map": map[string]interface{}{
					"Array":  []interface{}{"3", float64(4)},
					"Number": float64(3),
					"Map": map[string]interface{}{
						"Five": float64(5),
					},
				},
				"id2": "12-34-56",
				"map2": map[string]interface{}{
					"arr2": []interface{}{"b"},
				},
			}
		})

		TestFlatten()
	})
})
