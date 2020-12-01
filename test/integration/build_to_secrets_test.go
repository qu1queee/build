// Copyright The Shipwright Contributors
//
// SPDX-License-Identifier: Apache-2.0
package integration_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
	"github.com/shipwright-io/build/test"
	corev1 "k8s.io/api/core/v1"
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

			sampleSecret := tb.Catalog.SecretWithAnnotation(buildObject.Spec.Output.SecretRef.Name, buildObject.Namespace)

			Expect(tb.CreateSecret(sampleSecret)).To(BeNil())

			Expect(tb.CreateBuild(buildObject)).To(BeNil())

			// wait until the Build finish the validation
			buildObject, err := tb.GetBuildTillValidation(buildName)
			Expect(err).To(BeNil())
			Expect(buildObject.Status.Registered).To(Equal(corev1.ConditionTrue))
			Expect(buildObject.Status.Reason).To(Equal("Succeeded"))

			// delete a secret
			Expect(tb.DeleteSecret(buildObject.Spec.Output.SecretRef.Name)).To(BeNil())

			// assert that the validation happened one more time
			buildObject, err = tb.GetBuildTillRegistration(buildName, corev1.ConditionFalse)
			Expect(err).To(BeNil())
			Expect(buildObject.Status.Registered).To(Equal(corev1.ConditionFalse))
			Expect(buildObject.Status.Reason).To(Equal(fmt.Sprintf("secret %s does not exist", buildObject.Spec.Output.SecretRef.Name)))

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

			sampleSecret := tb.Catalog.SecretWithAnnotation(buildObject.Spec.Output.SecretRef.Name, buildObject.Namespace)

			// generate resources
			Expect(tb.CreateSecret(sampleSecret)).To(BeNil())

			// assert that the validation happened one more time
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

			sampleSecret := tb.Catalog.SecretWithoutAnnotation(buildObject.Spec.Output.SecretRef.Name, buildObject.Namespace)

			// generate resources
			Expect(tb.CreateSecret(sampleSecret)).To(BeNil())
			Expect(tb.CreateBuild(buildObject)).To(BeNil())

			// wait until the Build finish the validation
			buildObject, err := tb.GetBuildTillValidation(buildName)
			Expect(err).To(BeNil())
			Expect(buildObject.Status.Registered).To(Equal(corev1.ConditionTrue))
			Expect(buildObject.Status.Reason).To(Equal("Succeeded"))

			// delete a secret
			Expect(tb.DeleteSecret(buildObject.Spec.Output.SecretRef.Name)).To(BeNil())

			// assert that the validation happened one more time
			buildObject, err = tb.GetBuild(buildName)
			Expect(err).To(BeNil())
			Expect(buildObject.Status.Registered).To(Equal(corev1.ConditionTrue))
		})

		It("should not validate when a missing secret is recreated without annotation", func() {
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

			sampleSecret := tb.Catalog.SecretWithoutAnnotation(buildObject.Spec.Output.SecretRef.Name, buildObject.Namespace)

			// generate resources
			Expect(tb.CreateSecret(sampleSecret)).To(BeNil())

			// // assert that the validation happened one more time
			buildObject, err = tb.GetBuildTillRegistration(buildName, corev1.ConditionFalse)
			Expect(err).To(BeNil())
			Expect(buildObject.Status.Registered).To(Equal(corev1.ConditionFalse))
			Expect(buildObject.Status.Reason).To(Equal(fmt.Sprintf("secret %s does not exist", "fake-secret")))

		})

		It("should validate when a missing secret is recreated with annotation", func() {
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

			sampleSecret := tb.Catalog.SecretWithoutAnnotation(buildObject.Spec.Output.SecretRef.Name, buildObject.Namespace)

			// generate resources
			Expect(tb.CreateSecret(sampleSecret)).To(BeNil())

			// we modify the annotation so automatic delete does not take place
			data := []byte(fmt.Sprintf(`{"metadata":{"annotations":{"%s":"true"}}}`, v1alpha1.AnnotationBuildRefSecret))

			_, err = tb.PatchSecret(buildObject.Spec.Output.SecretRef.Name, data)
			Expect(err).To(BeNil())

			// // assert that the validation happened one more time
			buildObject, err = tb.GetBuildTillRegistration(buildName, corev1.ConditionTrue)
			Expect(err).To(BeNil())
			Expect(buildObject.Status.Registered).To(Equal(corev1.ConditionTrue))
			Expect(buildObject.Status.Reason).To(Equal("Succeeded"))

		})
	})
})

