// Copyright The Shipwright Contributors
//
// SPDX-License-Identifier: Apache-2.0

package sources_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	buildv1beta1 "github.com/shipwright-io/build/pkg/apis/build/v1beta1"
	"github.com/shipwright-io/build/pkg/config"
	"github.com/shipwright-io/build/pkg/reconciler/buildrun/resources/sources"

	pipelineapi "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	"k8s.io/utils/pointer"
)

var _ = Describe("Git", func() {

	cfg := config.NewDefaultConfig()

	Context("when adding a public Git source", func() {

		var taskSpec *pipelineapi.TaskSpec

		BeforeEach(func() {
			taskSpec = &pipelineapi.TaskSpec{}
		})

		JustBeforeEach(func() {
			sources.AppendGitStep(cfg, taskSpec, buildv1beta1.Git{
				URL: "https://github.com/shipwright-io/build",
			}, "default")
		})

		It("adds results for the commit sha, commit author and branch name", func() {
			Expect(len(taskSpec.Results)).To(Equal(3))
			Expect(taskSpec.Results[0].Name).To(Equal("shp-source-default-commit-sha"))
			Expect(taskSpec.Results[1].Name).To(Equal("shp-source-default-commit-author"))
			Expect(taskSpec.Results[2].Name).To(Equal("shp-source-default-branch-name"))
		})

		It("adds a step", func() {
			Expect(len(taskSpec.Steps)).To(Equal(1))
			Expect(taskSpec.Steps[0].Name).To(Equal("source-default"))
			Expect(taskSpec.Steps[0].Image).To(Equal(cfg.GitContainerTemplate.Image))
			Expect(taskSpec.Steps[0].Args).To(Equal([]string{
				"--url",
				"https://github.com/shipwright-io/build",
				"--target",
				"$(params.shp-source-root)",
				"--result-file-commit-sha",
				"$(results.shp-source-default-commit-sha.path)",
				"--result-file-commit-author",
				"$(results.shp-source-default-commit-author.path)",
				"--result-file-branch-name",
				"$(results.shp-source-default-branch-name.path)",
				"--result-file-error-message",
				"$(results.shp-error-message.path)",
				"--result-file-error-reason",
				"$(results.shp-error-reason.path)",
			}))
		})
	})

	Context("when adding a private Git source", func() {

		var taskSpec *pipelineapi.TaskSpec

		BeforeEach(func() {
			taskSpec = &pipelineapi.TaskSpec{}
		})

		JustBeforeEach(func() {
			sources.AppendGitStep(cfg, taskSpec, buildv1beta1.Git{
				URL:         "git@github.com:shipwright-io/build.git",
				CloneSecret: pointer.String("a.secret"),
			}, "default")
		})

		It("adds results for the commit sha, commit author and branch name", func() {
			Expect(len(taskSpec.Results)).To(Equal(3))
			Expect(taskSpec.Results[0].Name).To(Equal("shp-source-default-commit-sha"))
			Expect(taskSpec.Results[1].Name).To(Equal("shp-source-default-commit-author"))
			Expect(taskSpec.Results[2].Name).To(Equal("shp-source-default-branch-name"))
		})

		It("adds a volume for the secret", func() {
			Expect(len(taskSpec.Volumes)).To(Equal(1))
			Expect(taskSpec.Volumes[0].Name).To(Equal("shp-a-secret"))
			Expect(taskSpec.Volumes[0].VolumeSource.Secret).NotTo(BeNil())
			Expect(taskSpec.Volumes[0].VolumeSource.Secret.SecretName).To(Equal("a.secret"))
		})

		It("adds a step", func() {
			Expect(len(taskSpec.Steps)).To(Equal(1))
			Expect(taskSpec.Steps[0].Name).To(Equal("source-default"))
			Expect(taskSpec.Steps[0].Image).To(Equal(cfg.GitContainerTemplate.Image))
			Expect(taskSpec.Steps[0].Args).To(Equal([]string{
				"--url",
				"git@github.com:shipwright-io/build.git",
				"--target",
				"$(params.shp-source-root)",
				"--result-file-commit-sha",
				"$(results.shp-source-default-commit-sha.path)",
				"--result-file-commit-author",
				"$(results.shp-source-default-commit-author.path)",
				"--result-file-branch-name",
				"$(results.shp-source-default-branch-name.path)",
				"--result-file-error-message",
				"$(results.shp-error-message.path)",
				"--result-file-error-reason",
				"$(results.shp-error-reason.path)",
				"--secret-path",
				"/workspace/shp-source-secret",
			}))
			Expect(len(taskSpec.Steps[0].VolumeMounts)).To(Equal(1))
			Expect(taskSpec.Steps[0].VolumeMounts[0].Name).To(Equal("shp-a-secret"))
			Expect(taskSpec.Steps[0].VolumeMounts[0].MountPath).To(Equal("/workspace/shp-source-secret"))
			Expect(taskSpec.Steps[0].VolumeMounts[0].ReadOnly).To(BeTrue())
		})
	})
})
