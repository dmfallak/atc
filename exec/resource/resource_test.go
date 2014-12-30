package resource_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Resource", func() {
	Describe("Release", func() {
		It("destroys the container", func() {
			err := resource.Release()
			Ω(err).ShouldNot(HaveOccurred())

			Ω(gardenClient.Connection.DestroyCallCount()).Should(Equal(1))
			Ω(gardenClient.Connection.DestroyArgsForCall(0)).Should(Equal("some-handle"))
		})

		Context("when destroying the container fails", func() {
			disaster := errors.New("nope")

			BeforeEach(func() {
				gardenClient.Connection.DestroyReturns(disaster)
			})

			It("returns the error", func() {
				err := resource.Release()
				Ω(err).Should(Equal(disaster))
			})
		})
	})
})
