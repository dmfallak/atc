package exec_test

import (
	"bytes"
	"errors"
	"io"
	"os"

	"github.com/concourse/atc"
	. "github.com/concourse/atc/exec"
	"github.com/concourse/atc/exec/fakes"
	"github.com/concourse/atc/exec/resource"
	rfakes "github.com/concourse/atc/exec/resource/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tedsuo/ifrit"
)

var _ = Describe("GardenFactory", func() {
	var (
		fakeTracker *rfakes.FakeTracker

		factory Factory
	)

	BeforeEach(func() {
		fakeTracker = new(rfakes.FakeTracker)

		factory = NewGardenFactory(fakeTracker)
	})

	Describe("Get", func() {
		var (
			resourceConfig atc.ResourceConfig
			params         atc.Params
			version        atc.Version

			inSource ArtifactSource

			source  ArtifactSource
			process ifrit.Process
		)

		BeforeEach(func() {
			resourceConfig = atc.ResourceConfig{
				Name:   "some-resource",
				Type:   "some-resource-type",
				Source: atc.Source{"some": "source"},
			}

			params = atc.Params{"some-param": "some-value"}

			version = atc.Version{"some-version": "some-value"}

			inSource = nil // not needed for Get
		})

		JustBeforeEach(func() {
			source = factory.Get(resourceConfig, params, version).Using(inSource)
			process = ifrit.Invoke(source)
		})

		Context("when the tracker can initialize the resource", func() {
			var (
				fakeResource        *rfakes.FakeResource
				fakeVersionedSource *rfakes.FakeVersionedSource
			)

			BeforeEach(func() {
				fakeResource = new(rfakes.FakeResource)
				fakeTracker.InitReturns(fakeResource, nil)

				fakeVersionedSource = new(rfakes.FakeVersionedSource)
				fakeResource.GetReturns(fakeVersionedSource)
			})

			It("initializes the resource with the correct type", func() {
				Ω(fakeTracker.InitCallCount()).Should(Equal(1))

				typ := fakeTracker.InitArgsForCall(0)
				Ω(typ).Should(Equal(resource.ResourceType("some-resource-type")))
			})

			It("gets the resource with the correct source, params, and version", func() {
				Ω(fakeResource.GetCallCount()).Should(Equal(1))

				gotSource, gotParams, gotVersion := fakeResource.GetArgsForCall(0)
				Ω(gotSource).Should(Equal(resourceConfig.Source))
				Ω(gotParams).Should(Equal(params))
				Ω(gotVersion).Should(Equal(version))
			})

			It("executes the get resource action", func() {
				Ω(fakeVersionedSource.RunCallCount()).Should(Equal(1))
			})

			Describe("signalling", func() {
				var receivedSignals <-chan os.Signal

				BeforeEach(func() {
					sigs := make(chan os.Signal)
					receivedSignals = sigs

					fakeVersionedSource.RunStub = func(signals <-chan os.Signal, ready chan<- struct{}) error {
						close(ready)
						sigs <- <-signals
						return nil
					}
				})

				It("forwards to the resource", func() {
					process.Signal(os.Interrupt)
					Eventually(receivedSignals).Should(Receive(Equal(os.Interrupt)))
					Eventually(process.Wait()).Should(Receive())
				})
			})

			Describe("releasing", func() {
				Context("when releasing the resource succeeds", func() {
					BeforeEach(func() {
						fakeResource.ReleaseReturns(nil)
					})

					It("releases the resource", func() {
						Ω(fakeResource.ReleaseCallCount()).Should(BeZero())

						err := source.Release()
						Ω(err).ShouldNot(HaveOccurred())

						Ω(fakeResource.ReleaseCallCount()).Should(Equal(1))
					})
				})

				Context("when releasing the resource fails", func() {
					disaster := errors.New("nope")

					BeforeEach(func() {
						fakeResource.ReleaseReturns(disaster)
					})

					It("returns the error", func() {
						err := source.Release()
						Ω(err).Should(Equal(disaster))
					})
				})
			})

			Describe("streaming out", func() {
				Context("when the resource can stream out", func() {
					var streamedOut io.Reader

					BeforeEach(func() {
						streamedOut = bytes.NewBufferString("lol")
						fakeVersionedSource.StreamOutReturns(streamedOut, nil)
					})

					It("streams out the given path", func() {
						out, err := source.StreamOut("some-path")
						Ω(err).ShouldNot(HaveOccurred())

						Ω(out).Should(Equal(streamedOut))

						Ω(fakeVersionedSource.StreamOutArgsForCall(0)).Should(Equal("some-path"))
					})
				})

				Context("when the resource cannot stream out", func() {
					disaster := errors.New("nope")

					BeforeEach(func() {
						fakeVersionedSource.StreamOutReturns(nil, disaster)
					})

					It("returns the error", func() {
						_, err := source.StreamOut("some-path")
						Ω(err).Should(Equal(disaster))
					})
				})
			})
		})

		Context("when the tracker fails to initialize the resource", func() {
			disaster := errors.New("nope")

			BeforeEach(func() {
				fakeTracker.InitReturns(nil, disaster)
			})

			It("exits with the failure", func() {
				Eventually(process.Wait()).Should(Receive(Equal(disaster)))
			})
		})
	})

	Describe("Put", func() {
		var (
			resourceConfig atc.ResourceConfig
			params         atc.Params

			inSource *fakes.FakeArtifactSource

			source  ArtifactSource
			process ifrit.Process
		)

		BeforeEach(func() {
			resourceConfig = atc.ResourceConfig{
				Name:   "some-resource",
				Type:   "some-resource-type",
				Source: atc.Source{"some": "source"},
			}

			params = atc.Params{"some-param": "some-value"}

			inSource = new(fakes.FakeArtifactSource)
		})

		JustBeforeEach(func() {
			source = factory.Put(resourceConfig, params).Using(inSource)
			process = ifrit.Invoke(source)
		})

		Context("when the tracker can initialize the resource", func() {
			var (
				fakeResource        *rfakes.FakeResource
				fakeVersionedSource *rfakes.FakeVersionedSource

				streamedOut io.Reader
			)

			BeforeEach(func() {
				fakeResource = new(rfakes.FakeResource)
				fakeTracker.InitReturns(fakeResource, nil)

				fakeVersionedSource = new(rfakes.FakeVersionedSource)
				fakeResource.PutReturns(fakeVersionedSource)

				streamedOut = bytes.NewBufferString("lol")
				inSource.StreamOutReturns(streamedOut, nil)
			})

			It("initializes the resource with the correct type", func() {
				Ω(fakeTracker.InitCallCount()).Should(Equal(1))

				typ := fakeTracker.InitArgsForCall(0)
				Ω(typ).Should(Equal(resource.ResourceType("some-resource-type")))
			})

			It("puts the resource with the correct source and params, and the bits from the given source", func() {
				Ω(fakeResource.PutCallCount()).Should(Equal(1))

				putSource, putParams, putStream := fakeResource.PutArgsForCall(0)
				Ω(putSource).Should(Equal(resourceConfig.Source))
				Ω(putParams).Should(Equal(params))
				Ω(putStream).Should(Equal(streamedOut))
			})

			It("executes the get resource action", func() {
				Ω(fakeVersionedSource.RunCallCount()).Should(Equal(1))
			})

			Context("when streaming out from the previous source fails", func() {
				disaster := errors.New("nope")

				BeforeEach(func() {
					inSource.StreamOutReturns(nil, disaster)
				})

				It("exits with the error", func() {
					Eventually(process.Wait()).Should(Receive(Equal(disaster)))
				})
			})

			Describe("signalling", func() {
				var receivedSignals <-chan os.Signal

				BeforeEach(func() {
					sigs := make(chan os.Signal)
					receivedSignals = sigs

					fakeVersionedSource.RunStub = func(signals <-chan os.Signal, ready chan<- struct{}) error {
						close(ready)
						sigs <- <-signals
						return nil
					}
				})

				It("forwards to the resource", func() {
					process.Signal(os.Interrupt)
					Eventually(receivedSignals).Should(Receive(Equal(os.Interrupt)))
					Eventually(process.Wait()).Should(Receive())
				})
			})

			Describe("releasing", func() {
				Context("when releasing the resource succeeds", func() {
					BeforeEach(func() {
						fakeResource.ReleaseReturns(nil)
					})

					It("releases the resource", func() {
						Ω(fakeResource.ReleaseCallCount()).Should(BeZero())

						err := source.Release()
						Ω(err).ShouldNot(HaveOccurred())

						Ω(fakeResource.ReleaseCallCount()).Should(Equal(1))
					})
				})

				Context("when releasing the resource fails", func() {
					disaster := errors.New("nope")

					BeforeEach(func() {
						fakeResource.ReleaseReturns(disaster)
					})

					It("returns the error", func() {
						err := source.Release()
						Ω(err).Should(Equal(disaster))
					})
				})
			})

			Describe("streaming out", func() {
				Context("when the resource can stream out", func() {
					var streamedOut io.Reader

					BeforeEach(func() {
						streamedOut = bytes.NewBufferString("lol")
						fakeVersionedSource.StreamOutReturns(streamedOut, nil)
					})

					It("streams out the given path", func() {
						out, err := source.StreamOut("some-path")
						Ω(err).ShouldNot(HaveOccurred())

						Ω(out).Should(Equal(streamedOut))

						Ω(fakeVersionedSource.StreamOutArgsForCall(0)).Should(Equal("some-path"))
					})
				})

				Context("when the resource cannot stream out", func() {
					disaster := errors.New("nope")

					BeforeEach(func() {
						fakeVersionedSource.StreamOutReturns(nil, disaster)
					})

					It("returns the error", func() {
						_, err := source.StreamOut("some-path")
						Ω(err).Should(Equal(disaster))
					})
				})
			})
		})

		Context("when the tracker fails to initialize the resource", func() {
			disaster := errors.New("nope")

			BeforeEach(func() {
				fakeTracker.InitReturns(nil, disaster)
			})

			It("exits with the failure", func() {
				Eventually(process.Wait()).Should(Receive(Equal(disaster)))
			})
		})
	})
})
