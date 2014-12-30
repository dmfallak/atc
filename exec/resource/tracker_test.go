package resource_test

import (
	"errors"

	garden_api "github.com/cloudfoundry-incubator/garden/api"
	"github.com/cloudfoundry-incubator/garden/client/fake_api_client"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/concourse/atc/exec/resource"
)

var _ = Describe("Tracker", func() {
	var (
		resourceTypes ResourceMapping
		gardenClient  *fake_api_client.FakeClient

		tracker Tracker
	)

	BeforeEach(func() {
		resourceTypes = ResourceMapping{
			"type1": "image1",
			"type2": "image2",
		}

		gardenClient = fake_api_client.New()

		gardenClient.Connection.CreateReturns("some-handle", nil)

		tracker = NewTracker(resourceTypes, gardenClient)
	})

	Describe("Init", func() {
		var (
			initType ResourceType

			initResource Resource
			initErr      error
		)

		BeforeEach(func() {
			initType = "type1"
		})

		JustBeforeEach(func() {
			initResource, initErr = tracker.Init(initType)
		})

		It("does not error and returns a resource", func() {
			Ω(initErr).ShouldNot(HaveOccurred())
			Ω(initResource).ShouldNot(BeNil())
		})

		It("creates a privileged container with the resource type's image", func() {
			Ω(gardenClient.Connection.CreateArgsForCall(0)).Should(Equal(garden_api.ContainerSpec{
				RootFSPath: "image1",
				Privileged: true,
			}))
		})

		Context("when creating the container fails", func() {
			disaster := errors.New("oh no!")

			BeforeEach(func() {
				gardenClient.Connection.CreateReturns("", disaster)
			})

			It("returns the error and no resource", func() {
				Ω(initErr).Should(Equal(disaster))
				Ω(initResource).Should(BeNil())
			})
		})

		Context("with an unknown resource type", func() {
			BeforeEach(func() {
				initType = "bogus-type"
			})

			It("returns ErrUnknownResourceType", func() {
				Ω(initErr).Should(Equal(ErrUnknownResourceType))
			})
		})
	})
})
