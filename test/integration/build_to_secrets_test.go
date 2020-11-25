package integration_test

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
	"github.com/shipwright-io/build/test"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	Context("when a build reference a secret with annotations for the spec output", func() {
		It("should validate the Build after secret deletion", func() {

			// populate Build related vars
			buildName := BUILD + tb.Namespace
			buildObject, err = tb.Catalog.LoadBuildWithNameAndStrategy(
				buildName,
				STRATEGY+tb.Namespace,
				[]byte(test.BuildWithOutputRefSecret),
			)
			Expect(err).To(BeNil())

			sampleSecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "foo-secret",
				},
			}

			Expect(tb.CreateSecret(sampleSecret)).To(BeNil())

			data := []byte(fmt.Sprintf(`{"metadata":{"annotations":{"%s":"true"}}}`, v1alpha1.AnnotationBuildRefSecret))
			_, err = tb.PatchSecret("foo-secret", data)
			Expect(err).To(BeNil())

			Expect(tb.CreateBuild(buildObject)).To(BeNil())

			// wait until the Build finish the validation
			_, err := tb.GetBuildTillValidation(buildName)
			Expect(err).To(BeNil())

			// delete a secret
			Expect(tb.DeleteSecret("foo-secret")).To(BeNil())

			// assert that the validation happened one more time
			buildObject, err := tb.GetBuildTillRegistration(buildName, corev1.ConditionFalse)
			Expect(err).To(BeNil())
			Expect(buildObject.Status.Registered).To(Equal(corev1.ConditionFalse))
			Expect(buildObject.Status.Reason).To(Equal(fmt.Sprintf("secret %s does not exist", "foo-secret")))

		})

		It("should validate when a missing secret is recreated", func() {
			// populate Build related vars
			buildName := BUILD + tb.Namespace
			buildObject, err = tb.Catalog.LoadBuildWithNameAndStrategy(
				buildName,
				STRATEGY+tb.Namespace,
				[]byte(test.BuildCBSMinimalWithFakeSecret),
			)
			Expect(err).To(BeNil())

			Expect(tb.CreateBuild(buildObject)).To(BeNil())

			// wait until the Build finish the validation
			buildObject, err := tb.GetBuildTillValidation(buildName)
			Expect(err).To(BeNil())
			Expect(buildObject.Status.Registered).To(Equal(corev1.ConditionFalse))
			Expect(buildObject.Status.Reason).To(Equal(fmt.Sprintf("secret %s does not exist", "fake-secret")))

			sampleSecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "fake-secret",
				},
			}

			// generate resources
			Expect(tb.CreateSecret(sampleSecret)).To(BeNil())

			// we modify the annotation so automatic delete does not take place
			data := []byte(fmt.Sprintf(`{"metadata":{"annotations":{"%s":"true"}}}`, v1alpha1.AnnotationBuildRefSecret))
			_, err = tb.PatchSecret("fake-secret", data)
			Expect(err).To(BeNil())

			// // assert that the validation happened one more time
			buildObject, err = tb.GetBuildTillRegistration(buildName, corev1.ConditionTrue)
			Expect(err).To(BeNil())
			Expect(buildObject.Status.Registered).To(Equal(corev1.ConditionTrue))
			Expect(buildObject.Status.Reason).To(Equal("Succeeded"))

		})
	})

	Context("when a build reference a secret without annotations for the spec output", func() {
		It("should not validate the Build after a secret deletion", func() {

			// populate Build related vars
			buildName := BUILD + tb.Namespace
			buildObject, err = tb.Catalog.LoadBuildWithNameAndStrategy(
				buildName,
				STRATEGY+tb.Namespace,
				[]byte(test.BuildWithOutputRefSecret),
			)
			Expect(err).To(BeNil())

			// TODO: move this to the Catalog
			// TODO: secret name should be a variable in this test
			sampleSecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "foo-secret",
				},
			}

			// generate resources
			Expect(tb.CreateSecret(sampleSecret)).To(BeNil())
			Expect(tb.CreateBuild(buildObject)).To(BeNil())

			// wait until the Build finish the validation
			_, err := tb.GetBuildTillValidation(buildName)
			Expect(err).To(BeNil())

			// delete a secret
			Expect(tb.DeleteSecret("foo-secret")).To(BeNil())

			// assert that the validation happened one more time
			buildObject, err := tb.GetBuild(buildName)
			Expect(err).To(BeNil())
			Expect(buildObject.Status.Registered).To(Equal(corev1.ConditionTrue))
		})
	})

	Context("when the build reconciles on secret events", func() {
		It("should avoid reconciling a secret update that is not related to our annotation", func() {

			// TODO: We need to wait until the controller is up, we need to find a way to do that, sleep
			// is not ideal
			time.Sleep(time.Second * 10)

			sampleSecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "foo-secret",
				},
			}

			// generate resources
			Expect(tb.CreateSecret(sampleSecret)).To(BeNil())

			// we modify the annotation so automatic delete does not take place
			data := []byte(fmt.Sprintf(`{"metadata":{"annotations":{"%s":"true"}}}`, v1alpha1.AnnotationBuildRefSecret))
			_, err = tb.PatchSecret("foo-secret", data)
			Expect(err).To(BeNil())

			// we modify the annotation so automatic delete does not take place
			data = []byte(fmt.Sprintf(`{"metadata":{"annotations":{"%s":"true"}}}`, v1alpha1.AnnotationBuildRefSecret))
			_, err = tb.PatchSecret("foo-secret", data)
			Expect(err).To(BeNil())

		})
		It("should only reconcile on create events when the secret have the annotations", func() {
		})
		It("should only reconcile on delete events when the secret have the annotations", func() {
		})
	})
})
