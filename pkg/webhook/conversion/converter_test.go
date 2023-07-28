// Copyright The Shipwright Contributors
//
// SPDX-License-Identifier: Apache-2.0
package conversion_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
	"github.com/shipwright-io/build/pkg/webhook/conversion"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
)

func getConversionReview(o string) (apiextensionsv1.ConversionReview, error) {
	convertReview := apiextensionsv1.ConversionReview{}
	response := httptest.NewRecorder()
	request, err := http.NewRequest("POST", "/convert", strings.NewReader(o))
	if err != nil {
		return convertReview, err
	}
	request.Header.Add("Content-Type", "application/yaml")

	conversion.CRDConvert(response, request, context.TODO())

	scheme := runtime.NewScheme()

	yamlSerializer := json.NewYAMLSerializer(json.DefaultMetaFactory, scheme, scheme)
	if _, _, err := yamlSerializer.Decode(response.Body.Bytes(), nil, &convertReview); err != nil {
		return convertReview, err
	}

	return convertReview, nil
}

var _ = Describe("ConvertCRD", func() {
	Context("for a Build CR from v1beta1 to v1alpha1", func() {

		var apiVersion = "apiextensions.k8s.io/v1"
		var desiredAPIVersion = "shipwright.io/v1alpha1"

		It("converts for source OCIArtifacts type", func() {
			ctxDir := "docker-build"
			image := "dockerhub/foobar/hello"
			pruneOption := "AfterPull"
			secretName := "foobar"
			buildTemplate := `kind: ConversionReview
apiVersion: %s
request:
  uid: 0000-0000-0000-0000
  desiredAPIVersion: %s
  objects:
    - apiVersion: shipwright.io/v1beta1
      kind: Build
      metadata:
        name: buildkit-build
      spec:
        source:
          type: OCI
          contextDir: %s
          ociArtifact:
            image: %s
            prune: %s
            pullSecret: %s
`
			o := fmt.Sprintf(buildTemplate, apiVersion,
				desiredAPIVersion, ctxDir,
				image, pruneOption,
				secretName)

			conversionReview, err := getConversionReview(o)
			Expect(err).To(BeNil())
			Expect(conversionReview.Response.Result.Status).To(Equal(v1.StatusSuccess))

			convertedObj, err := ToUnstructured(conversionReview)
			Expect(err).To(BeNil())

			build, err := toV1Alpha1BuildObject(convertedObj)
			Expect(err).To(BeNil())

			Expect(build.Spec.Source.Credentials).To(Equal(&corev1.LocalObjectReference{
				Name: secretName,
			}))

			Expect(build.Spec.Source.BundleContainer.Image).To(Equal(image))
			Expect(*build.Spec.Source.BundleContainer.Prune).To(Equal(v1alpha1.PruneAfterPull))
			Expect(build.Spec.Source.ContextDir).To(Equal(&ctxDir))
			Expect(build.Spec.Source.Revision).To(BeNil())
		})
		It("converts for source GitSource type", func() {
			ctxDir := "docker-build"
			url := "https://github.com/shipwright-io/sample-go"
			revision := "main"
			secretName := "foobar"
			buildTemplate := `kind: ConversionReview
apiVersion: %s
request:
  uid: 0000-0000-0000-0000
  desiredAPIVersion: %s
  objects:
    - apiVersion: shipwright.io/v1beta1
      kind: Build
      metadata:
        name: buildkit-build
      spec:
        source:
          type: Git
          contextDir: %s
          git:
            url: %s
            revision: %s
            cloneSecret: %s
`
			o := fmt.Sprintf(buildTemplate, apiVersion,
				desiredAPIVersion, ctxDir,
				url, revision,
				secretName)

			conversionReview, err := getConversionReview(o)
			Expect(err).To(BeNil())
			Expect(conversionReview.Response.Result.Status).To(Equal(v1.StatusSuccess))

			convertedObj, err := ToUnstructured(conversionReview)
			Expect(err).To(BeNil())

			build, err := toV1Alpha1BuildObject(convertedObj)
			Expect(err).To(BeNil())

			Expect(build.Spec.Source.Credentials).To(Equal(&corev1.LocalObjectReference{
				Name: secretName,
			}))

			Expect(build.Spec.Source.ContextDir).To(Equal(&ctxDir))
			Expect(build.Spec.Source.URL).To(Equal(&url))
			Expect(build.Spec.Source.Revision).To(Equal(&revision))
		})
		It("converts for spec triggers", func() {
			ttype := "GitHub"
			event := "Push"
			branch_01 := "main"
			branch_02 := "develop"
			buildTemplate := `kind: ConversionReview
apiVersion: %s
request:
  uid: 0000-0000-0000-0000
  desiredAPIVersion: %s
  objects:
    - apiVersion: shipwright.io/v1beta1
      kind: Build
      metadata:
        name: buildkit-build
      spec:
        trigger:
          when:
          - name:
            type: %s
            github:
              events:
              - %s
              branches:
              - %s
              - %s
          triggerSecret: foobar
`
			o := fmt.Sprintf(buildTemplate, apiVersion,
				desiredAPIVersion, ttype,
				event, branch_01,
				branch_02)

			conversionReview, err := getConversionReview(o)
			Expect(err).To(BeNil())
			Expect(conversionReview.Response.Result.Status).To(Equal(v1.StatusSuccess))

			convertedObj, err := ToUnstructured(conversionReview)
			Expect(err).To(BeNil())

			build, err := toV1Alpha1BuildObject(convertedObj)
			Expect(err).To(BeNil())

			Expect(build.Spec.Trigger.When[0].Type).To(Equal(v1alpha1.GitHubWebHookTrigger))

			Expect(build.Spec.Trigger.When[0].GitHub.Branches).To(ContainElements(branch_01, branch_02))
			Expect(build.Spec.Trigger.When[0].GitHub.Events).To(ContainElement(v1alpha1.GitHubPushEvent))
			Expect(build.Spec.Trigger.SecretRef).To(Equal(&corev1.LocalObjectReference{Name: "foobar"}))
		})
		It("converts for spec strategy", func() {
			strategyName := "buildkit"
			strategyKind := "ClusterBuildStrategy"
			buildTemplate := `kind: ConversionReview
apiVersion: %s
request:
  uid: 0000-0000-0000-0000
  desiredAPIVersion: %s
  objects:
    - apiVersion: shipwright.io/v1beta1
      kind: Build
      metadata:
        name: buildkit-build
      spec:
        strategy:
          name: %s
          kind: %s
`
			o := fmt.Sprintf(buildTemplate, apiVersion,
				desiredAPIVersion, strategyName,
				strategyKind)

			conversionReview, err := getConversionReview(o)
			Expect(err).To(BeNil())
			Expect(conversionReview.Response.Result.Status).To(Equal(v1.StatusSuccess))

			convertedObj, err := ToUnstructured(conversionReview)
			Expect(err).To(BeNil())

			build, err := toV1Alpha1BuildObject(convertedObj)
			Expect(err).To(BeNil())

			Expect(build.Spec.Strategy.Name).To(Equal(strategyName))
			Expect(*build.Spec.Strategy.Kind).To(Equal(v1alpha1.ClusterBuildStrategyKind))
		})
		It("converts for spec params", func() {
			buildTemplate := `kind: ConversionReview
apiVersion: %s
request:
  uid: 0000-0000-0000-0000
  desiredAPIVersion: %s
  objects:
    - apiVersion: shipwright.io/v1beta1
      kind: Build
      metadata:
        name: buildkit-build
      spec:
        paramValues:
        - name: foo1
          value: disabled
        - name: foo1
          values:
          - value: NODE_VERSION=16
          - configMapValue:
              name: project-configuration
              key: node-version
              format: NODE_VERSION=${CONFIGMAP_VALUE}
          - secretValue:
              name: npm-registry-access
              key: npm-auth-token
              format: NPM_AUTH_TOKEN=${SECRET_VALUE}
`
			o := fmt.Sprintf(buildTemplate, apiVersion,
				desiredAPIVersion)

			conversionReview, err := getConversionReview(o)
			Expect(err).To(BeNil())
			Expect(conversionReview.Response.Result.Status).To(Equal(v1.StatusSuccess))

			convertedObj, err := ToUnstructured(conversionReview)
			Expect(err).To(BeNil())

			build, err := toV1Alpha1BuildObject(convertedObj)
			Expect(err).To(BeNil())

			// we could extend here the assertions, keeping it simple for now
			Expect(len(build.Spec.ParamValues)).To(Equal(2))
			Expect(len(build.Spec.ParamValues[1].Values)).To(Equal(3))
		})
		It("converts for spec output", func() {
			img := "image-registry.openshift-image-registry"
			secretName := "foobar"
			buildTemplate := `kind: ConversionReview
apiVersion: %s
request:
  uid: 0000-0000-0000-0000
  desiredAPIVersion: %s
  objects:
    - apiVersion: shipwright.io/v1beta1
      kind: Build
      metadata:
        name: buildkit-build
      spec:
        timeout: 10m
        output:
          image: %s
          pushSecret: %s
`
			o := fmt.Sprintf(buildTemplate, apiVersion,
				desiredAPIVersion, img, secretName)

			conversionReview, err := getConversionReview(o)
			Expect(err).To(BeNil())
			Expect(conversionReview.Response.Result.Status).To(Equal(v1.StatusSuccess))

			convertedObj, err := ToUnstructured(conversionReview)
			Expect(err).To(BeNil())

			build, err := toV1Alpha1BuildObject(convertedObj)
			Expect(err).To(BeNil())

			Expect(build.Spec.Output.Image).To(Equal(img))
			Expect(build.Spec.Output.Credentials.Name).To(Equal(secretName))
		})
		It("converts for spec retention and volumes", func() {
			limit := uint(10)
			buildTemplate := `kind: ConversionReview
apiVersion: %s
request:
  uid: 0000-0000-0000-0000
  desiredAPIVersion: %s
  objects:
    - apiVersion: shipwright.io/v1beta1
      kind: Build
      metadata:
        name: buildkit-build
      spec:
        retention:
          failedLimit: %v
          succeededLimit: %v
          ttlAfterFailed: 30m
          ttlAfterSucceeded: 30m
        volumes:
        - name: gocache
          description: "do it"
          overridable: true
          emptyDir: {}
        - name: foobar
          emptyDir: {}
`
			o := fmt.Sprintf(buildTemplate, apiVersion,
				desiredAPIVersion, limit, limit)

			conversionReview, err := getConversionReview(o)
			Expect(err).To(BeNil())
			Expect(conversionReview.Response.Result.Status).To(Equal(v1.StatusSuccess))

			convertedObj, err := ToUnstructured(conversionReview)
			Expect(err).To(BeNil())

			build, err := toV1Alpha1BuildObject(convertedObj)
			Expect(err).To(BeNil())

			Expect(build.Spec.Retention).To(Equal(&v1alpha1.BuildRetention{
				FailedLimit:    &limit,
				SucceededLimit: &limit,
				TTLAfterFailed: &v1.Duration{
					Duration: 30 * time.Minute},
				TTLAfterSucceeded: &v1.Duration{
					Duration: 30 * time.Minute,
				},
			}))
			Expect(len(build.Spec.Volumes)).To(Equal(2))
			Expect(build.Spec.Volumes[1].Name).To(Equal("foobar"))
			Expect(build.Spec.Volumes[1].EmptyDir).To(Equal(&corev1.EmptyDirVolumeSource{}))
		})
	})

	Context("for a Build CR from v1alpha1 to v1beta1", func() {

	})
})

func ToUnstructured(conversionReview apiextensionsv1.ConversionReview) (unstructured.Unstructured, error) {
	convertedObj := unstructured.Unstructured{}

	scheme := runtime.NewScheme()
	yamlSerializer := json.NewYAMLSerializer(json.DefaultMetaFactory, scheme, scheme)
	if _, _, err := yamlSerializer.Decode(conversionReview.Response.ConvertedObjects[0].Raw, nil, &convertedObj); err != nil {
		return convertedObj, err
	}
	return convertedObj, nil
}

func toV1Alpha1BuildObject(convertedObject unstructured.Unstructured) (v1alpha1.Build, error) {
	var build v1alpha1.Build
	u := convertedObject.UnstructuredContent()
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u, &build); err != nil {
		return build, err
	}
	return build, nil
}
