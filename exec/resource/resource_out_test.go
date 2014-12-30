package resource_test

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"os"

	garden "github.com/cloudfoundry-incubator/garden/api"
	gfakes "github.com/cloudfoundry-incubator/garden/api/fakes"
	"github.com/tedsuo/ifrit"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	"github.com/concourse/atc"
	. "github.com/concourse/atc/exec/resource"
)

var _ = Describe("Resource Out", func() {
	var (
		source  atc.Source
		params  atc.Params
		version atc.Version

		outScriptStdout     string
		outScriptStderr     string
		outScriptExitStatus int
		runOutError         error

		outScriptProcess *gfakes.FakeProcess

		versionedSource VersionedSource
		outProcess      ifrit.Process
	)

	BeforeEach(func() {
		source = atc.Source{"some": "source"}
		params = atc.Params{"some": "params"}
		version = atc.Version{"original": "version"}

		outScriptStdout = "{}"
		outScriptStderr = ""
		outScriptExitStatus = 0
		runOutError = nil

		outScriptProcess = new(gfakes.FakeProcess)
		outScriptProcess.WaitStub = func() (int, error) {
			return outScriptExitStatus, nil
		}
	})

	JustBeforeEach(func() {
		gardenClient.Connection.RunStub = func(handle string, spec garden.ProcessSpec, io garden.ProcessIO) (garden.Process, error) {
			if runOutError != nil {
				return nil, runOutError
			}

			_, err := io.Stdout.Write([]byte(outScriptStdout))
			Ω(err).ShouldNot(HaveOccurred())

			_, err = io.Stderr.Write([]byte(outScriptStderr))
			Ω(err).ShouldNot(HaveOccurred())

			return outScriptProcess, nil
		}

		versionedSource = resource.Put(source, params, version, bytes.NewBufferString("the-source"))
		outProcess = ifrit.Invoke(versionedSource)
	})

	It("runs /opt/resource/out <source path> with the request on stdin", func() {
		Eventually(outProcess.Wait()).Should(Receive(BeNil()))

		handle, spec, io := gardenClient.Connection.RunArgsForCall(0)
		Ω(handle).Should(Equal("some-handle"))
		Ω(spec.Path).Should(Equal("/opt/resource/out"))
		Ω(spec.Args).Should(Equal([]string{"/tmp/build/src"}))
		Ω(spec.Privileged).Should(BeTrue())

		request, err := ioutil.ReadAll(io.Stdin)
		Ω(err).ShouldNot(HaveOccurred())

		Ω(request).Should(MatchJSON(`{
			"params": {"some":"params"},
			"source": {"some":"source"},
			"version": {"original":"version"}
		}`))
	})

	Context("when streaming the source in succeeds", func() {
		var streamedIn *gbytes.Buffer

		BeforeEach(func() {
			streamedIn = gbytes.NewBuffer()

			gardenClient.Connection.StreamInStub = func(handle string, destination string, source io.Reader) error {
				Ω(handle).Should(Equal("some-handle"))

				if destination == "/tmp/build/src" {
					_, err := io.Copy(streamedIn, source)
					Ω(err).ShouldNot(HaveOccurred())
				}

				return nil
			}
		})

		It("writes the stream source to the destination", func() {
			Eventually(outProcess.Wait()).Should(Receive(BeNil()))

			Ω(string(streamedIn.Contents())).Should(Equal("the-source"))
		})
	})

	Context("when /opt/resource/out prints the version and metadata", func() {
		BeforeEach(func() {
			outScriptStdout = `{
				"version": {"some": "new-version"},
				"metadata": [
					{"name": "a", "value":"a-value"},
					{"name": "b","value": "b-value"}
				]
			}`
		})

		It("returns the build source printed out by /opt/resource/out", func() {
			Eventually(outProcess.Wait()).Should(Receive(BeNil()))

			Ω(versionedSource.Version()).Should(Equal(atc.Version{"some": "new-version"}))
			Ω(versionedSource.Metadata()).Should(Equal([]atc.MetadataField{
				{Name: "a", Value: "a-value"},
				{Name: "b", Value: "b-value"},
			}))
		})
	})

	// Context("when /out outputs to stderr", func() {
	// 	BeforeEach(func() {
	// 		outScriptStderr = "some stderr data"
	// 	})
	//
	// 	It("emits it to the log sink", func() {
	// 		Ω(outErr).ShouldNot(HaveOccurred())
	//
	// 		Ω(string(logs.Contents())).Should(Equal("some stderr data"))
	// 	})
	// })

	Context("when streaming in the source fails", func() {
		disaster := errors.New("oh no!")

		BeforeEach(func() {
			gardenClient.Connection.StreamInReturns(disaster)
		})

		It("returns the error", func() {
			Eventually(outProcess.Wait()).Should(Receive(Equal(disaster)))
		})
	})

	Context("when running /opt/resource/out fails", func() {
		disaster := errors.New("oh no!")

		BeforeEach(func() {
			runOutError = disaster
		})

		It("returns the error", func() {
			Eventually(outProcess.Wait()).Should(Receive(Equal(disaster)))
		})
	})

	Context("when /opt/resource/out exits nonzero", func() {
		BeforeEach(func() {
			outScriptStdout = "some-stdout-data"
			outScriptStderr = "some-stderr-data"
			outScriptExitStatus = 9
		})

		It("returns an err containing stdout/stderr of the process", func() {
			var outErr error
			Eventually(outProcess.Wait()).Should(Receive(&outErr))

			Ω(outErr).Should(HaveOccurred())
			Ω(outErr.Error()).Should(ContainSubstring("some-stdout-data"))
			Ω(outErr.Error()).Should(ContainSubstring("some-stderr-data"))
			Ω(outErr.Error()).Should(ContainSubstring("exit status 9"))
		})
	})

	Context("when a signal is received", func() {
		var waited chan<- struct{}

		BeforeEach(func() {
			waiting := make(chan struct{})
			waited = waiting

			outScriptProcess.WaitStub = func() (int, error) {
				// cause waiting to block so that it can be aborted
				<-waiting
				return 0, nil
			}
		})

		It("stops the container", func() {
			outProcess.Signal(os.Interrupt)

			Eventually(gardenClient.Connection.StopCallCount).Should(Equal(1))

			handle, kill := gardenClient.Connection.StopArgsForCall(0)
			Ω(handle).Should(Equal("some-handle"))
			Ω(kill).Should(BeFalse())

			close(waited)
		})
	})

	Describe("streaming bits out", func() {
		Context("when streaming out succeeds", func() {
			BeforeEach(func() {
				gardenClient.Connection.StreamOutStub = func(handle string, source string) (io.ReadCloser, error) {
					Ω(handle).Should(Equal("some-handle"))

					streamOut := new(bytes.Buffer)

					if source == "/tmp/build/src/some/subdir" {
						streamOut.WriteString("sup")
					}

					return ioutil.NopCloser(streamOut), nil
				}
			})

			It("returns the output stream of /tmp/build/src/some-name/", func() {
				Eventually(outProcess.Wait()).Should(Receive(BeNil()))

				inStream, err := versionedSource.StreamOut("some/subdir")
				Ω(err).ShouldNot(HaveOccurred())

				contents, err := ioutil.ReadAll(inStream)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(string(contents)).Should(Equal("sup"))
			})
		})

		Context("when streaming out fails", func() {
			disaster := errors.New("oh no!")

			BeforeEach(func() {
				gardenClient.Connection.StreamOutReturns(nil, disaster)
			})

			It("returns the error", func() {
				Eventually(outProcess.Wait()).Should(Receive(BeNil()))

				_, err := versionedSource.StreamOut("some/subdir")
				Ω(err).Should(Equal(disaster))
			})
		})
	})
})
