package exec_test

import (
	"errors"

	"github.com/cloudfoundry-incubator/candiedyaml"
	"github.com/concourse/atc"
	. "github.com/concourse/atc/exec"
	"github.com/concourse/atc/exec/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("ConfigSource", func() {
	var (
		someConfig = atc.TaskConfig{
			Platform: "some-platform",
			Tags:     []string{"some", "tags"},
			Image:    "some-image",
			Params:   map[string]string{"PARAM": "value"},
			Run: atc.TaskRunConfig{
				Path: "ls",
				Args: []string{"-al"},
			},
			Inputs: []atc.TaskInputConfig{
				{Name: "some-input", Path: "some-path"},
			},
		}

		fakeSourceRepository *fakes.FakeSourceRepository
	)

	BeforeEach(func() {
		fakeSourceRepository = new(fakes.FakeSourceRepository)
	})

	Describe("StaticConfigSource", func() {
		var (
			configSource TaskConfigSource

			fetchedConfig atc.TaskConfig
			fetchErr      error
		)

		BeforeEach(func() {
			configSource = StaticConfigSource{Config: someConfig}
		})

		JustBeforeEach(func() {
			fetchedConfig, fetchErr = configSource.FetchConfig(fakeSourceRepository)
		})

		It("succeeds", func() {
			Ω(fetchErr).ShouldNot(HaveOccurred())
		})

		It("returns the static config", func() {
			Ω(fetchedConfig).Should(Equal(someConfig))
		})
	})

	Describe("FileConfigSource", func() {
		var (
			configSource FileConfigSource

			fetchedConfig atc.TaskConfig
			fetchErr      error
		)

		BeforeEach(func() {
			configSource = FileConfigSource{Path: "some/build.yml"}
		})

		JustBeforeEach(func() {
			fetchedConfig, fetchErr = configSource.FetchConfig(fakeSourceRepository)
		})

		Context("when the path does not indicate an artifact source", func() {
			BeforeEach(func() {
				configSource.Path = "foo-bar.yml"
			})

			It("returns an error", func() {
				Ω(fetchErr).Should(Equal(UnspecifiedArtifactSourceError{"foo-bar.yml"}))
			})
		})

		Context("when the file's artifact source can be found in the repository", func() {
			var fakeArtifactSource *fakes.FakeArtifactSource

			BeforeEach(func() {
				fakeArtifactSource = new(fakes.FakeArtifactSource)
				fakeSourceRepository.SourceForReturns(fakeArtifactSource, true)
			})

			Context("when the artifact source provides a proper file", func() {
				var streamedOut *gbytes.Buffer

				BeforeEach(func() {
					marshalled, err := candiedyaml.Marshal(someConfig)
					Ω(err).ShouldNot(HaveOccurred())

					streamedOut = gbytes.BufferWithBytes(marshalled)
					fakeArtifactSource.StreamFileReturns(streamedOut, nil)
				})

				It("finds the artifact source via the first path segment", func() {
					Ω(fakeSourceRepository.SourceForArgsForCall(0)).Should(Equal(SourceName("some")))
				})

				It("fetches the file via the correct path", func() {
					Ω(fakeArtifactSource.StreamFileArgsForCall(0)).Should(Equal("build.yml"))
				})

				It("succeeds", func() {
					Ω(fetchErr).ShouldNot(HaveOccurred())
				})

				It("returns the unmarshalled config", func() {
					Ω(fetchedConfig).Should(Equal(someConfig))
				})

				It("closes the stream", func() {
					Ω(streamedOut.Closed()).Should(BeTrue())
				})
			})

			Context("when the artifact source provides an invalid configuration", func() {
				var streamedOut *gbytes.Buffer

				BeforeEach(func() {
					invalidConfig := someConfig
					invalidConfig.Platform = ""
					invalidConfig.Run = atc.TaskRunConfig{}

					marshalled, err := candiedyaml.Marshal(invalidConfig)
					Ω(err).ShouldNot(HaveOccurred())

					streamedOut = gbytes.BufferWithBytes(marshalled)
					fakeArtifactSource.StreamFileReturns(streamedOut, nil)
				})

				It("returns an error", func() {
					Ω(fetchErr).Should(HaveOccurred())
				})
			})

			Context("when the artifact source provides a malformed file", func() {
				var streamedOut *gbytes.Buffer

				BeforeEach(func() {
					streamedOut = gbytes.BufferWithBytes([]byte("bogus"))
					fakeArtifactSource.StreamFileReturns(streamedOut, nil)
				})

				It("fails", func() {
					Ω(fetchErr).Should(HaveOccurred())
				})

				It("closes the stream", func() {
					Ω(streamedOut.Closed()).Should(BeTrue())
				})
			})

			Context("when streaming the file out fails", func() {
				disaster := errors.New("nope")

				BeforeEach(func() {
					fakeArtifactSource.StreamFileReturns(nil, disaster)
				})

				It("returns the error", func() {
					Ω(fetchErr).Should(HaveOccurred())
				})
			})
		})

		Context("when the file's artifact source cannot be found in the repository", func() {
			BeforeEach(func() {
				fakeSourceRepository.SourceForReturns(nil, false)
			})

			It("returns an UnknownArtifactSourceError", func() {
				Ω(fetchErr).Should(Equal(UnknownArtifactSourceError{"some"}))
			})
		})
	})

	Describe("MergedConfigSource", func() {
		var (
			fakeConfigSourceA *fakes.FakeTaskConfigSource
			fakeConfigSourceB *fakes.FakeTaskConfigSource

			configSource TaskConfigSource

			fetchedConfig atc.TaskConfig
			fetchErr      error
		)

		BeforeEach(func() {
			fakeConfigSourceA = new(fakes.FakeTaskConfigSource)
			fakeConfigSourceB = new(fakes.FakeTaskConfigSource)

			configSource = MergedConfigSource{
				A: fakeConfigSourceA,
				B: fakeConfigSourceB,
			}
		})

		JustBeforeEach(func() {
			fetchedConfig, fetchErr = configSource.FetchConfig(fakeSourceRepository)
		})

		Context("when fetching via A succeeds", func() {
			var configA = atc.TaskConfig{
				Image:  "some-image",
				Params: map[string]string{"PARAM": "A"},
			}

			BeforeEach(func() {
				fakeConfigSourceA.FetchConfigReturns(configA, nil)
			})

			Context("and fetching via B succeeds", func() {
				var configB = atc.TaskConfig{
					Params: map[string]string{"PARAM": "B"},
				}

				BeforeEach(func() {
					fakeConfigSourceB.FetchConfigReturns(configB, nil)
				})

				It("fetches via the input source", func() {
					Ω(fakeConfigSourceA.FetchConfigArgsForCall(0)).Should(Equal(fakeSourceRepository))
					Ω(fakeConfigSourceB.FetchConfigArgsForCall(0)).Should(Equal(fakeSourceRepository))
				})

				It("succeeds", func() {
					Ω(fetchErr).ShouldNot(HaveOccurred())
				})

				It("returns the merged config", func() {
					Ω(fetchedConfig).Should(Equal(atc.TaskConfig{
						Image:  "some-image",
						Params: map[string]string{"PARAM": "B"},
					}))
				})
			})

			Context("and fetching via B fails", func() {
				disaster := errors.New("nope")

				BeforeEach(func() {
					fakeConfigSourceB.FetchConfigReturns(atc.TaskConfig{}, disaster)
				})

				It("returns the error", func() {
					Ω(fetchErr).Should(Equal(disaster))
				})
			})
		})

		Context("when fetching via A fails", func() {
			disaster := errors.New("nope")

			BeforeEach(func() {
				fakeConfigSourceA.FetchConfigReturns(atc.TaskConfig{}, disaster)
			})

			It("returns the error", func() {
				Ω(fetchErr).Should(Equal(disaster))
			})

			It("does not fetch via B", func() {
				Ω(fakeConfigSourceB.FetchConfigCallCount()).Should(Equal(0))
			})
		})
	})
})
