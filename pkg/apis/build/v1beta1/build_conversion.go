// Copyright The Shipwright Contributors
//
// SPDX-License-Identifier: Apache-2.0

package v1beta1

import (
	"github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

// ConvertTo converts this Build to the Hub version (v1alpha1)
func (src *Build) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1alpha1.Build)

	dst.ObjectMeta = src.ObjectMeta

	return src.Spec.ConvertTo(&dst.Spec)
}

func (srcSpec *BuildSpec) ConvertTo(bs *v1alpha1.BuildSpec) error {
	// BuildSpec Source
	bs.Source = getBuildSource(*srcSpec)

	// BuildSpec Sources: todo
	// Note: conversion does not matches, as we come from v1beta1, where
	// we only have a single source

	// BuildSpec Trigger
	for _, t := range srcSpec.Trigger.When {
		tw := v1alpha1.TriggerWhen{}
		t.convertTo(&tw)
		bs.Trigger.When = append(bs.Trigger.When, tw)
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
	bs.Output.Credentials.Name = *srcSpec.Output.PushSecret
	bs.Output.Annotations = srcSpec.Output.Annotations
	bs.Output.Labels = srcSpec.Output.Labels

	// BuildSpec Timeout
	bs.Timeout = srcSpec.Timeout

	// BuildSpec Env
	bs.Env = srcSpec.Env

	// BuildSpec Retention
	bs.Retention.FailedLimit = srcSpec.Retention.FailedLimit
	bs.Retention.SucceededLimit = srcSpec.Retention.SucceededLimit
	bs.Retention.TTLAfterFailed = srcSpec.Retention.TTLAfterFailed
	bs.Retention.TTLAfterSucceeded = srcSpec.Retention.TTLAfterSucceeded

	// BuildSpec Volumes
	for i, vol := range srcSpec.Volumes {
		bs.Volumes[i].Name = vol.Name
		bs.Volumes[i].Description = nil
		bs.Volumes[i].VolumeSource = vol.VolumeSource

	}
	return nil
}

// ConvertFrom converts from the Hub version (v1alpha1) to this version.
// TODO: Not needed?
func (dst *Build) ConvertFrom(srcRaw conversion.Hub) error {
	return nil
}

// todo: could be placed in its own file
func (p ParamValue) convertTo(dest *v1alpha1.ParamValue) {
	dest.Value = p.Value
	dest.ConfigMapValue = (*v1alpha1.ObjectKeyRef)(p.ConfigMapValue)
	dest.SecretValue = (*v1alpha1.ObjectKeyRef)(p.SecretValue)
	dest.Name = p.Name

	for _, singleValue := range p.Values {
		dest.Values = append(dest.Values, v1alpha1.SingleValue{
			Value:          singleValue.Value,
			ConfigMapValue: (*v1alpha1.ObjectKeyRef)(singleValue.ConfigMapValue),
			SecretValue:    (*v1alpha1.ObjectKeyRef)(singleValue.SecretValue),
		})
	}
}

// todo: could be placed in its own file
func (p TriggerWhen) convertTo(dest *v1alpha1.TriggerWhen) {
	dest.Name = p.Name
	dest.Type = v1alpha1.TriggerType(p.Type)

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
		credentials = corev1.LocalObjectReference{
			Name: *src.Source.GitSource.CloneSecret,
		}
		source.URL = src.Source.GitSource.URL
		revision = src.Source.GitSource.Revision
	}

	source.Credentials = &credentials
	source.Revision = revision
	source.ContextDir = src.Source.ContextDir

	return source
}
