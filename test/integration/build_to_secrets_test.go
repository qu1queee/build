package integration_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
	"github.com/shipwright-io/build/test"
)

var _ = Describe("Integration tests Build and referenced Secrets", func() {

	var (
		cbsObject   *v1alpha1.ClusterBuildStrategy
		buildObject *v1alpha1.Build
	)
	// Load the ClusterBuildStrategies before each test case
	BeforeEach(func() {
		cbsObject, err = tb.Catalog.LoadCBSWithName(STRATEGY+tb.Namespace, []byte(test.ClusterBuildStrategySingleStep))
		Expect(err).To(BeNil())

		err = tb.CreateClusterBuildStrategy(cbsObject)
		Expect(err).To(BeNil())
	})

	// Delete the ClusterBuildStrategies after each test case
	AfterEach(func() {
		err := tb.DeleteClusterBuildStrategy(cbsObject.Name)
		Expect(err).To(BeNil())
	})

	Context("when a build reference a secret for the registry", func() {
		It("should validate the Build after secret deletion", func() {
			buildObject, err = tb.Catalog.LoadBuildWithNameAndStrategy(
				BUILD+tb.Namespace,
				STRATEGY+tb.Namespace,
				[]byte(test.BuildTODO),
			)
			Expect(err).To(BeNil())
			// TODO:
			// Create Secret
			Expect(tb.CreateBuild(buildObject)).To(BeNil())

			// TODO:
			// Delete secret-foo
			// Get Build
			// Assert Build REGISTERED field to equal FALSE
			// Assert Build REASON matches that the secret is not found
		})
	})
})
