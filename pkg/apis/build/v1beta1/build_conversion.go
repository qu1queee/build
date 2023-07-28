// Copyright The Shipwright Contributors
//
// SPDX-License-Identifier: Apache-2.0

package v1beta1

import (
	"github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

// ConvertTo converts this Build to the Hub version (v1alpha1)
func (src *Build) ConvertTo(bs *v1alpha1.Build) error {
	bs.ObjectMeta = src.ObjectMeta
	return src.Spec.ConvertTo(&bs.Spec)
}

func (srcSpec *BuildSpec) ConvertTo(bs *v1alpha1.BuildSpec) error {
	// BuildSpec Source
	bs.Source = getBuildSource(*srcSpec)

	// BuildSpec Trigger
	if srcSpec.Trigger != nil {
		bs.Trigger = &v1alpha1.Trigger{}
		for _, t := range srcSpec.Trigger.When {
			tw := v1alpha1.TriggerWhen{}
			t.convertTo(&tw)
			bs.Trigger.When = append(bs.Trigger.When, tw)
		}
		if srcSpec.Trigger.TriggerSecret != nil {
			bs.Trigger.SecretRef = &corev1.LocalObjectReference{Name: *srcSpec.Trigger.TriggerSecret}
		}
	}

	// BuildSpec Strategy
	bs.Strategy = v1alpha1.Strategy{
		Name:       srcSpec.StrategyName(),
		Kind:       (*v1alpha1.BuildStrategyKind)(srcSpec.Strategy.Kind),
		APIVersion: srcSpec.Strategy.APIVersion,
	}

	// BuildSpec Builder, no migration possible
	bs.Builder = nil

	// BuildSpec Dockerfile, no migration possible
	bs.Dockerfile = nil

	// BuildSpec ParamValues
	bs.ParamValues = nil
	for _, p := range srcSpec.ParamValues {
		new := v1alpha1.ParamValue{}
		p.convertTo(&new)
		bs.ParamValues = append(bs.ParamValues, new)
	}

	// BuildSpec Output
	insecure := false
	bs.Output.Image = srcSpec.Output.Image
	bs.Output.Insecure = &insecure
	if srcSpec.Output.PushSecret != nil {
		bs.Output.Credentials = &corev1.LocalObjectReference{}
		bs.Output.Credentials.Name = *srcSpec.Output.PushSecret
	}
	bs.Output.Annotations = srcSpec.Output.Annotations
	bs.Output.Labels = srcSpec.Output.Labels

	// BuildSpec Timeout
	bs.Timeout = srcSpec.Timeout

	// BuildSpec Env
	bs.Env = srcSpec.Env

	// BuildSpec Retention
	bs.Retention = &v1alpha1.BuildRetention{}
	if srcSpec.Retention != nil && srcSpec.Retention.FailedLimit != nil {
		bs.Retention.FailedLimit = srcSpec.Retention.FailedLimit
	}
	if srcSpec.Retention != nil && srcSpec.Retention.SucceededLimit != nil {

		bs.Retention.SucceededLimit = srcSpec.Retention.SucceededLimit
	}
	if srcSpec.Retention != nil && srcSpec.Retention.TTLAfterFailed != nil {
		bs.Retention.TTLAfterFailed = srcSpec.Retention.TTLAfterFailed
	}
	if srcSpec.Retention != nil && srcSpec.Retention.TTLAfterSucceeded != nil {
		bs.Retention.TTLAfterSucceeded = srcSpec.Retention.TTLAfterSucceeded
	}

	// BuildSpec Volumes
	bs.Volumes = []v1alpha1.BuildVolume{}
	for _, vol := range srcSpec.Volumes {
		aux := v1alpha1.BuildVolume{
			Name:         vol.Name,
			Description:  nil,
			VolumeSource: vol.VolumeSource,
		}
		bs.Volumes = append(bs.Volumes, aux)
	}
	return nil
}

func (p ParamValue) convertTo(dest *v1alpha1.ParamValue) {

	if p.SingleValue != nil && p.SingleValue.Value != nil {
		dest.SingleValue = &v1alpha1.SingleValue{}
		dest.Value = p.Value
	}

	if p.ConfigMapValue != nil {
		dest.ConfigMapValue = &v1alpha1.ObjectKeyRef{}
		dest.ConfigMapValue = (*v1alpha1.ObjectKeyRef)(p.ConfigMapValue)
	}
	if p.SecretValue != nil {
		dest.SecretValue = (*v1alpha1.ObjectKeyRef)(p.SecretValue)
	}

	dest.Name = p.Name

	for _, singleValue := range p.Values {
		dest.Values = append(dest.Values, v1alpha1.SingleValue{
			Value:          singleValue.Value,
			ConfigMapValue: (*v1alpha1.ObjectKeyRef)(singleValue.ConfigMapValue),
			SecretValue:    (*v1alpha1.ObjectKeyRef)(singleValue.SecretValue),
		})
	}
}

func (p TriggerWhen) convertTo(dest *v1alpha1.TriggerWhen) {
	dest.Name = p.Name
	dest.Type = v1alpha1.TriggerType(p.Type)

	dest.GitHub = &v1alpha1.WhenGitHub{}
	for _, e := range p.GitHub.Events {
		dest.GitHub.Events = append(dest.GitHub.Events, v1alpha1.GitHubEventName(e))
	}
	dest.GitHub.Branches = p.GetBranches(GitHubWebHookTrigger)

	dest.Image = (*v1alpha1.WhenImage)(p.Image)
	dest.ObjectRef = (*v1alpha1.WhenObjectRef)(p.ObjectRef)

}

func getBuildSource(src BuildSpec) v1alpha1.Source {
	source := v1alpha1.Source{}
	var credentials corev1.LocalObjectReference
	var revision *string

	switch src.Source.Type {
	case OCIArtifactType:
		credentials = corev1.LocalObjectReference{
			Name: *src.Source.OCIArtifact.PullSecret,
		}
		source.BundleContainer = &v1alpha1.BundleContainer{
			Image: src.Source.OCIArtifact.Image,
			Prune: (*v1alpha1.PruneOption)(src.Source.OCIArtifact.Prune),
		}
	default:
		if src.Source.GitSource != nil && src.Source.GitSource.CloneSecret != nil {
			credentials = corev1.LocalObjectReference{
				Name: *src.Source.GitSource.CloneSecret,
			}
		}
		if src.Source.GitSource != nil {
			source.URL = src.Source.GitSource.URL
			revision = src.Source.GitSource.Revision
		}

	}

	source.Credentials = &credentials
	source.Revision = revision
	source.ContextDir = src.Source.ContextDir

	return source
}
