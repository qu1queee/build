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

// ensure v1beta1 implements the Conversion interface
var _ webhook.Conversion = (*BuildRun)(nil)

// To Alpha
func (src *BuildRun) ConvertTo(ctx context.Context, obj *unstructured.Unstructured) error {

	var bs v1alpha1.BuildRun

	bs.TypeMeta = src.TypeMeta
	bs.TypeMeta.APIVersion = ALPHA_GROUP_VERSION
	bs.ObjectMeta = src.ObjectMeta

	// BuildRunSpec BuildSpec
	newBuildSpec := v1alpha1.BuildSpec{}
	if src.Spec.Build.Build != nil {
		if err := src.Spec.Build.Build.ConvertTo(&newBuildSpec); err != nil {
			return err
		}
	}

	if src.Spec.Build.Build != nil {
		bs.Spec.BuildSpec = &newBuildSpec
	} else {
		bs.Spec.BuildRef = &v1alpha1.BuildRef{
			Name: src.Spec.Build.Name,
		}
	}

	// BuildRunSpec ServiceAccount
	bs.Spec.ServiceAccount = &v1alpha1.ServiceAccount{
		Name: src.Spec.ServiceAccount,
	}

	// BuildRunSpec Timeout
	bs.Spec.Timeout = src.Spec.Timeout

	// BuildRunSpec ParamValues
	bs.Spec.ParamValues = nil
	for _, p := range src.Spec.ParamValues {
		new := v1alpha1.ParamValue{}
		p.convertToAlpha(&new)
		bs.Spec.ParamValues = append(bs.Spec.ParamValues, new)
	}

	// BuildRunSpec Image

	if src.Spec.Output != nil {
		bs.Spec.Output = &v1alpha1.Image{
			Image:       src.Spec.Output.Image,
			Annotations: src.Spec.Output.Annotations,
			Labels:      src.Spec.Output.Labels,
		}
		if src.Spec.Output.PushSecret != nil {
			bs.Spec.Output.Credentials = &corev1.LocalObjectReference{
				Name: *src.Spec.Output.PushSecret,
			}
		}
	}

	// BuildRunSpec State
	bs.Spec.State = (*v1alpha1.BuildRunRequestedState)(src.Spec.State)

	// BuildRunSpec Env
	bs.Spec.Env = src.Spec.Env

	// BuildRunSpec Retention
	bs.Spec.Retention = (*v1alpha1.BuildRunRetention)(src.Spec.Retention)

	// BuildRunSpec Volumes
	bs.Spec.Volumes = []v1alpha1.BuildVolume{}
	for _, vol := range src.Spec.Volumes {
		bs.Spec.Volumes = append(bs.Spec.Volumes, v1alpha1.BuildVolume{
			Name:         vol.Name,
			VolumeSource: vol.VolumeSource,
		})
	}

	mapito, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&bs)
	if err != nil {
		ctxlog.Error(ctx, err, "failed structuring the newObject")
	}
	obj.Object = mapito

	return nil

}

// From Alpha
func (src *BuildRun) ConvertFrom(ctx context.Context, obj *unstructured.Unstructured) error {

	var br v1alpha1.BuildRun

	unstructured := obj.UnstructuredContent()
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructured, &br)
	if err != nil {
		ctxlog.Error(ctx, err, "failed unstructuring the buildrun convertedObject")
	}

	src.ObjectMeta = br.ObjectMeta
	src.TypeMeta = br.TypeMeta
	src.TypeMeta.APIVersion = BETA_GROUP_VERSION

	src.Spec.ConvertFrom(&br.Spec)

	sources := []SourceResult{}
	for _, s := range br.Status.Sources {
		sr := SourceResult{
			Name:        s.Name,
			Git:         (*GitSourceResult)(s.Git),
			OciArtifact: (*OciArtifactSourceResult)(s.Bundle),
		}
		sources = append(sources, sr)
	}

	conditions := []Condition{}

	for _, c := range br.Status.Conditions {
		ct := Condition{
			Type:               Type(c.Type),
			Status:             c.Status,
			LastTransitionTime: c.LastTransitionTime,
			Reason:             c.Reason,
			Message:            c.Message,
		}
		conditions = append(conditions, ct)
	}

	buildBeta := Build{}
	if br.Status.BuildSpec != nil {
		buildBeta.Spec.ConvertFrom(br.Status.BuildSpec)
	}

	src.Status = BuildRunStatus{
		Sources:        sources,
		Output:         (*Output)(br.Status.Output),
		Conditions:     conditions,
		TaskRunName:    br.Status.LatestTaskRunRef,
		StartTime:      br.Status.StartTime,
		CompletionTime: br.Status.CompletionTime,
		BuildSpec:      &buildBeta.Spec,
		FailureDetails: src.Status.FailureDetails,
	}

	return nil
}

func (dest *BuildRunSpec) ConvertFrom(orig *v1alpha1.BuildRunSpec) error {

	// BuildRunSpec BuildSpec
	dest.Build = &ReferencedBuild{}
	if orig.BuildSpec != nil {
		if dest.Build.Build != nil {
			dest.Build.Build.ConvertFrom(orig.BuildSpec)
		}
	}
	if orig.BuildRef != nil {
		dest.Build.Name = orig.BuildRef.Name
	}

	if orig.ServiceAccount != nil {
		dest.ServiceAccount = orig.ServiceAccount.Name
	}

	dest.Timeout = orig.Timeout

	// BuildRunSpec ParamValues
	dest.ParamValues = []ParamValue{}
	for _, p := range orig.ParamValues {
		new := convertBetaParamValue(p)
		dest.ParamValues = append(dest.ParamValues, new)
	}

	// Handle BuildSpec Output
	dest.Output = &Image{}
	if orig.Output != nil {
		dest.Output.Image = orig.Output.Image
	}

	if orig.Output != nil && orig.Output.Credentials != nil {
		dest.Output.PushSecret = &orig.Output.Credentials.Name
	}

	// BuildRunSpec State
	dest.State = (*BuildRunRequestedState)(orig.State)

	// BuildRunSpec Env
	dest.Env = orig.Env

	// BuildRunSpec Retention
	dest.Retention = (*BuildRunRetention)(orig.Retention)

	// BuildRunSpec Volumes
	for i, vol := range orig.Volumes {
		dest.Volumes[i].Name = vol.Name
		dest.Volumes[i].VolumeSource = vol.VolumeSource

	}
	return nil
}
