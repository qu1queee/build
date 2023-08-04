// Copyright The Shipwright Contributors
//
// SPDX-License-Identifier: Apache-2.0

package v1beta1

import (
	"context"

	"github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
	"github.com/shipwright-io/build/pkg/ctxlog"
	"github.com/shipwright-io/build/pkg/webhook"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

const (
	BETA_GROUP_VERSION  = "shipwright.io/v1beta1"
	ALPHA_GROUP_VERSION = "shipwright.io/v1alpha1"
)

// ensure v1beta1 implements the Conversion interface
var _ webhook.Conversion = (*Build)(nil)

func (src *Build) ConvertTo(ctx context.Context, obj *unstructured.Unstructured) error {
	var bs v1alpha1.Build

	bs.TypeMeta = src.TypeMeta
	bs.TypeMeta.APIVersion = ALPHA_GROUP_VERSION

	bs.ObjectMeta = src.ObjectMeta

	src.Spec.ConvertTo(&bs.Spec)
	mapito, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&bs)
	if err != nil {
		ctxlog.Error(ctx, err, "failed structuring the newObject")
	}
	obj.Object = mapito

	return nil

}

func (src *Build) ConvertFrom(ctx context.Context, obj *unstructured.Unstructured) error {

	var bs v1alpha1.Build

	unstructured := obj.UnstructuredContent()
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructured, &bs)
	if err != nil {
		ctxlog.Error(ctx, err, "failed unstructuring the convertedObject")
	}
	src.ObjectMeta = bs.ObjectMeta
	src.TypeMeta = bs.TypeMeta
	src.TypeMeta.APIVersion = BETA_GROUP_VERSION

	src.Spec.ConvertFrom(&bs.Spec)

	src.Status = BuildStatus{
		Registered: bs.Status.Registered,
		Reason:     (*BuildReason)(bs.Status.Reason),
		Message:    bs.Status.Message,
	}

	return nil
}

func (dest *BuildSpec) ConvertFrom(orig *v1alpha1.BuildSpec) error {
	// Handle BuildSpec Source
	specSource := Source{}
	if orig.Source.BundleContainer != nil {
		specSource.Type = OCIArtifactType
		specSource.OCIArtifact = &OCIArtifact{
			Image:      orig.Source.BundleContainer.Image,
			Prune:      (*PruneOption)(orig.Source.BundleContainer.Prune),
			PullSecret: &orig.Source.Credentials.Name,
		}
	} else {
		specSource.Type = GitType
		specSource.GitSource = &Git{
			URL:      orig.Source.URL,
			Revision: orig.Source.Revision,
		}
		if orig.Source.Credentials != nil {
			specSource.GitSource.CloneSecret = &orig.Source.Credentials.Name
		}
	}
	specSource.ContextDir = orig.Source.ContextDir
	dest.Source = specSource

	// Handle BuildSpec Triggers
	if orig.Trigger != nil {
		dest.Trigger = &Trigger{}
		for _, t := range orig.Trigger.When {
			tw := convertToBetaTriggers(&t)
			dest.Trigger.When = append(dest.Trigger.When, tw)
		}
		if orig.Trigger.SecretRef != nil {
			dest.Trigger.TriggerSecret = &orig.Trigger.SecretRef.Name
		}
	}

	// Handle BuildSpec Strategy
	dest.Strategy = Strategy{
		Name:       orig.StrategyName(),
		Kind:       (*BuildStrategyKind)(orig.Strategy.Kind),
		APIVersion: orig.Strategy.APIVersion,
	}

	// Handle BuildSpec ParamValues
	dest.ParamValues = []ParamValue{}
	for _, p := range orig.ParamValues {
		new := convertBetaParamValue(p)
		dest.ParamValues = append(dest.ParamValues, new)
	}

	// Handle BuildSpec Output
	dest.Output.Image = orig.Output.Image
	if orig.Output.Credentials != nil {
		dest.Output.PushSecret = &orig.Output.Credentials.Name
	}

	dest.Output.Annotations = orig.Output.Annotations
	dest.Output.Labels = orig.Output.Labels

	// Handle BuildSpec Timeout
	dest.Timeout = orig.Timeout

	// Handle BuildSpec Env
	dest.Env = orig.Env

	// Handle BuildSpec Retention
	dest.Retention = &BuildRetention{}
	if orig.Retention != nil {
		if orig.Retention.FailedLimit != nil {
			dest.Retention.FailedLimit = orig.Retention.FailedLimit
		}
		if orig.Retention.SucceededLimit != nil {
			dest.Retention.SucceededLimit = orig.Retention.SucceededLimit
		}
		if orig.Retention.TTLAfterFailed != nil {
			dest.Retention.TTLAfterFailed = orig.Retention.TTLAfterFailed
		}
		if orig.Retention.TTLAfterSucceeded != nil {
			dest.Retention.TTLAfterSucceeded = orig.Retention.TTLAfterSucceeded
		}
	}

	// Handle BuildSpec Volumes
	dest.Volumes = []BuildVolume{}
	for _, vol := range orig.Volumes {
		aux := BuildVolume{
			Name:         vol.Name,
			VolumeSource: vol.VolumeSource,
		}
		dest.Volumes = append(dest.Volumes, aux)
	}

	return nil
}

func (srcSpec *BuildSpec) ConvertTo(bs *v1alpha1.BuildSpec) error {
	// Handle BuildSpec Source
	bs.Source = getAlphaBuildSource(*srcSpec)

	// Handle BuildSpec Trigger
	if srcSpec.Trigger != nil {
		bs.Trigger = &v1alpha1.Trigger{}
		for _, t := range srcSpec.Trigger.When {
			tw := v1alpha1.TriggerWhen{}
			t.convertToAlpha(&tw)
			bs.Trigger.When = append(bs.Trigger.When, tw)
		}
		if srcSpec.Trigger.TriggerSecret != nil {
			bs.Trigger.SecretRef = &corev1.LocalObjectReference{Name: *srcSpec.Trigger.TriggerSecret}
		}
	}

	// Handle BuildSpec Strategy
	bs.Strategy = v1alpha1.Strategy{
		Name:       srcSpec.StrategyName(),
		Kind:       (*v1alpha1.BuildStrategyKind)(srcSpec.Strategy.Kind),
		APIVersion: srcSpec.Strategy.APIVersion,
	}

	// Handle BuildSpec Builder, TODO
	bs.Builder = nil

	// Handle BuildSpec Dockerfile, TODO
	bs.Dockerfile = nil

	// Handle BuildSpec ParamValues
	bs.ParamValues = nil
	for _, p := range srcSpec.ParamValues {
		new := v1alpha1.ParamValue{}
		p.convertToAlpha(&new)
		bs.ParamValues = append(bs.ParamValues, new)
	}

	// Handle BuildSpec Output
	insecure := false
	bs.Output.Image = srcSpec.Output.Image
	bs.Output.Insecure = &insecure
	if srcSpec.Output.PushSecret != nil {
		bs.Output.Credentials = &corev1.LocalObjectReference{}
		bs.Output.Credentials.Name = *srcSpec.Output.PushSecret
	}
	bs.Output.Annotations = srcSpec.Output.Annotations
	bs.Output.Labels = srcSpec.Output.Labels

	// Handle BuildSpec Timeout
	bs.Timeout = srcSpec.Timeout

	// Handle BuildSpec Env
	bs.Env = srcSpec.Env

	// Handle BuildSpec Retention
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

	// Handle BuildSpec Volumes
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

func (p ParamValue) convertToAlpha(dest *v1alpha1.ParamValue) {

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

func (p TriggerWhen) convertToAlpha(dest *v1alpha1.TriggerWhen) {
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

func convertBetaParamValue(orig v1alpha1.ParamValue) ParamValue {
	p := ParamValue{}
	if orig.SingleValue != nil && orig.SingleValue.Value != nil {
		p.SingleValue = &SingleValue{}
		p.Value = orig.Value
	}

	if orig.ConfigMapValue != nil {
		p.ConfigMapValue = &ObjectKeyRef{}
		p.ConfigMapValue = (*ObjectKeyRef)(orig.ConfigMapValue)
	}
	if orig.SecretValue != nil {
		p.SecretValue = (*ObjectKeyRef)(orig.SecretValue)
	}

	p.Name = orig.Name

	for _, singleValue := range orig.Values {
		p.Values = append(p.Values, SingleValue{
			Value:          singleValue.Value,
			ConfigMapValue: (*ObjectKeyRef)(singleValue.ConfigMapValue),
			SecretValue:    (*ObjectKeyRef)(singleValue.SecretValue),
		})
	}
	return p
}

func convertToBetaTriggers(orig *v1alpha1.TriggerWhen) TriggerWhen {
	dest := TriggerWhen{
		Name: orig.Name,
		Type: TriggerType(orig.Type),
	}

	dest.GitHub = &WhenGitHub{}
	for _, e := range orig.GitHub.Events {
		dest.GitHub.Events = append(dest.GitHub.Events, GitHubEventName(e))
	}

	dest.GitHub.Branches = orig.GetBranches(v1alpha1.GitHubWebHookTrigger)
	dest.Image = (*WhenImage)(orig.Image)
	dest.ObjectRef = (*WhenObjectRef)(orig.ObjectRef)

	return dest
}

func getAlphaBuildSource(src BuildSpec) v1alpha1.Source {
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

	if credentials.Name != "" {
		source.Credentials = &credentials
	}

	source.Revision = revision
	source.ContextDir = src.Source.ContextDir

	return source
}
