package tracer

import (
	"fmt"
	"testing"

	"github.com/epsagon/epsagon-go/protocol"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestMaskIgnoredKeys(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Mask Ignored Keys")
}

var _ = Describe("mask_ignored_keys", func() {
	Describe("maskEventIgnoredKeys", func() {
		config := &Config{
			Disable: true,
		}
		testTracer := CreateTracer(config).(*epsagonTracer)
		Context("sanity", func() {
			It("masks sanity test", func() {
				event := &protocol.Event{
					Resource: &protocol.Resource{
						Metadata: map[string]string{
							"not-ignored": "hello",
							"ignored":     "bye",
						},
					},
				}
				ignoredKeys := []string{"ignored"}
				testTracer.Config.IgnoredKeys = ignoredKeys
				testTracer.maskEventIgnoredKeys(event, ignoredKeys)
				Expect(event.Resource.Metadata["not-ignored"]).To(Equal("hello"))
				Expect(event.Resource.Metadata["ignored"]).To(Equal(maskedValue))
			})
		})
		Context("test matrix", func() {
			type testCase struct {
				metadata         map[string]string
				ignoredKeys      []string
				expectedMetadata map[string]string
			}
			testMatrix := map[string]testCase{
				"handles empty metadata": {
					metadata:         map[string]string{},
					ignoredKeys:      []string{"ignored"},
					expectedMetadata: map[string]string{},
				},
				"handles empty ignoredKeys": {
					metadata:         map[string]string{"ignored": "bye"},
					ignoredKeys:      []string{},
					expectedMetadata: map[string]string{"ignored": "bye"},
				},
				"passes matrix sanity": {
					metadata:         map[string]string{"to-be-ignored": "bye", "not-ignored": "hello"},
					ignoredKeys:      []string{"to-be-ignored"},
					expectedMetadata: map[string]string{"to-be-ignored": maskedValue, "not-ignored": "hello"},
				},
				"handles json map without ignored keys": {
					metadata:         map[string]string{"not-ignored": "hello", "other-not-ignored": "{\"hello\":\"world\"}"},
					ignoredKeys:      []string{"to-be-ignored"},
					expectedMetadata: map[string]string{"not-ignored": "hello", "other-not-ignored": "{\"hello\":\"world\"}"},
				},
				"handles json map *with* ignored keys": {
					metadata:         map[string]string{"not-ignored": "hello", "other-not-ignored": "{\"to-be-ignored\":\"world\"}"},
					ignoredKeys:      []string{"to-be-ignored"},
					expectedMetadata: map[string]string{"not-ignored": "hello", "other-not-ignored": fmt.Sprintf("{\"to-be-ignored\":\"%s\"}", maskedValue)},
				},
				"handles json nested array and map without ignored keys": {
					metadata:         map[string]string{"not-ignored": "hello", "other-not-ignored": "[{\"hello\":\"world\"},{\"erez\":\"is-cool\"}]"},
					ignoredKeys:      []string{"to-be-ignored"},
					expectedMetadata: map[string]string{"not-ignored": "hello", "other-not-ignored": "[{\"hello\":\"world\"},{\"erez\":\"is-cool\"}]"},
				},
				"handles json nested array and map *with* ignored keys": {
					metadata: map[string]string{
						"not-ignored":       "hello",
						"other-not-ignored": "[{\"to-be-ignored\":\"world\"},{\"erez\":\"is-cool\"}]",
					},
					ignoredKeys: []string{"to-be-ignored"},
					expectedMetadata: map[string]string{
						"not-ignored": "hello",
						"other-not-ignored": fmt.Sprintf(
							"[{\"to-be-ignored\":\"%s\"},{\"erez\":\"is-cool\"}]",
							maskedValue),
					},
				},
				"handles json nested map *with* ignored keys": {
					metadata: map[string]string{
						"not-ignored":       "hello",
						"other-not-ignored": "{\"wait\":{\"for\":{\"it\":{\"to-be-ignored\":\"boom\"},\"not\":[\"I\",\"are\",\"baboon\"]},\"not\":\"it\"}}",
					},
					ignoredKeys: []string{"to-be-ignored"},
					expectedMetadata: map[string]string{
						"not-ignored": "hello",
						"other-not-ignored": fmt.Sprintf(
							"{\"wait\":{\"for\":{\"it\":{\"to-be-ignored\":\"%s\"},\"not\":[\"I\",\"are\",\"baboon\"]},\"not\":\"it\"}}",
							maskedValue,
						),
					},
				},
			}
			for testName, value := range testMatrix {
				value := value
				It(testName, func() {
					testMetadata := make(map[string]string)
					for k, v := range value.metadata {
						testMetadata[k] = v
					}
					event := &protocol.Event{
						Resource: &protocol.Resource{
							Metadata: testMetadata,
						},
					}
					testTracer.maskEventIgnoredKeys(event, value.ignoredKeys)
					Expect(event.Resource.Metadata).To(Equal(value.expectedMetadata))
				})
			}
		})
	})
})
