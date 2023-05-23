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

	// BuildSpec Source
	dst.Spec.Source = getBuildSource(*src)

	// BuildSpec Sources: todo
	// Note: conversion does not matches, as we come from v1beta1, where
	// we only have a single source

	// BuildSpec Trigger
	for _, t := range src.Spec.Trigger.When {
		tw := v1alpha1.TriggerWhen{}
		t.convertTo(&tw)
		dst.Spec.Trigger.When = append(dst.Spec.Trigger.When, tw)
	}

	// BuildSpec Strategy
	dst.Spec.Strategy = v1alpha1.Strategy{
		Name:       src.Spec.StrategyName(),
		Kind:       (*v1alpha1.BuildStrategyKind)(src.Spec.Strategy.Kind),
		APIVersion: src.Spec.Strategy.APIVersion,
	}

	// BuildSpec Builder, no migration possible
	dst.Spec.Builder = nil

	// BuildSpec Dockerfile, no migration possible
	dst.Spec.Dockerfile = nil

	// BuildSpec ParamValues
	dst.Spec.ParamValues = nil
	for _, p := range src.Spec.ParamValues {
		new := v1alpha1.ParamValue{}
		p.convertTo(&new)
		dst.Spec.ParamValues = append(dst.Spec.ParamValues, new)
	}

	// BuildSpec Output
	insecure := false
	dst.Spec.Output.Image = src.Spec.Output.Image
	dst.Spec.Output.Insecure = &insecure
	dst.Spec.Output.Credentials.Name = *src.Spec.Output.PushSecret
	dst.Spec.Output.Annotations = src.Spec.Output.Annotations
	dst.Spec.Output.Labels = src.Spec.Output.Labels

	// BuildSpec Timeout
	dst.Spec.Timeout = src.Spec.Timeout

	// BuildSpec Env
	dst.Spec.Env = src.Spec.Env

	// BuildSpec Retention
	dst.Spec.Retention.FailedLimit = src.Spec.Retention.FailedLimit
	dst.Spec.Retention.SucceededLimit = src.Spec.Retention.SucceededLimit
	dst.Spec.Retention.TTLAfterFailed = src.Spec.Retention.TTLAfterFailed
	dst.Spec.Retention.TTLAfterSucceeded = src.Spec.Retention.TTLAfterSucceeded

	// BuildSpec Volumes
	for i, vol := range src.Spec.Volumes {
		dst.Spec.Volumes[i].Name = vol.Name
		dst.Spec.Volumes[i].Description = nil
		dst.Spec.Volumes[i].VolumeSource = vol.VolumeSource

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

func getBuildSource(src Build) v1alpha1.Source {
	source := v1alpha1.Source{}
	var credentials corev1.LocalObjectReference
	var revision *string

	switch src.Spec.Source.Type {
	case OCIArtifactType:
		credentials = corev1.LocalObjectReference{
			Name: *src.Spec.Source.OCIArtifact.PullSecret,
		}
		source.BundleContainer = &v1alpha1.BundleContainer{
			Image: src.Spec.Source.OCIArtifact.Image,
			Prune: (*v1alpha1.PruneOption)(src.Spec.Source.OCIArtifact.Prune),
		}
	default:
		credentials = corev1.LocalObjectReference{
			Name: *src.Spec.Source.GitSource.CloneSecret,
		}
		source.URL = src.Spec.Source.GitSource.URL
		revision = src.Spec.Source.GitSource.Revision
	}

	source.Credentials = &credentials
	source.Revision = revision
	source.ContextDir = src.Spec.Source.ContextDir

	return source
}
