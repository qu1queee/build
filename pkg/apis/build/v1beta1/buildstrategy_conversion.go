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
var _ webhook.Conversion = (*BuildStrategy)(nil)

// To Alpha
func (src *BuildStrategy) ConvertTo(ctx context.Context, obj *unstructured.Unstructured) error {
	var bs v1alpha1.BuildStrategy
	bs.TypeMeta = src.TypeMeta
	bs.TypeMeta.APIVersion = ALPHA_GROUP_VERSION
	bs.ObjectMeta = src.ObjectMeta

	bs.Spec.BuildSteps = []v1alpha1.BuildStep{}
	for _, step := range src.Spec.Steps {

		buildStep := v1alpha1.BuildStep{
			Container: corev1.Container{
				Name:            step.Name,
				Image:           step.Image,
				Command:         step.Command,
				Args:            step.Args,
				WorkingDir:      step.WorkingDir,
				Env:             step.Env,
				Resources:       step.Resources,
				VolumeMounts:    step.VolumeMounts,
				ImagePullPolicy: step.ImagePullPolicy,
			},
		}

		if step.SecurityContext != nil {
			buildStep.SecurityContext = step.SecurityContext
		}

		bs.Spec.BuildSteps = append(bs.Spec.BuildSteps, buildStep)
	}

	bs.Spec.Parameters = []v1alpha1.Parameter{}
	for _, param := range src.Spec.Parameters {
		bs.Spec.Parameters = append(bs.Spec.Parameters, v1alpha1.Parameter{
			Name:        param.Name,
			Description: param.Description,
			Type:        v1alpha1.ParameterType(param.Type),
			Default:     param.Default,
			Defaults:    param.Defaults,
		})
	}

	if src.Spec.SecurityContext != nil {
		bs.Spec.SecurityContext = (*v1alpha1.BuildStrategySecurityContext)(src.Spec.SecurityContext)
	}

	bs.Spec.Volumes = []v1alpha1.BuildStrategyVolume{}
	for _, vol := range src.Spec.Volumes {
		bs.Spec.Volumes = append(bs.Spec.Volumes, v1alpha1.BuildStrategyVolume{
			Overridable:  vol.Overridable,
			Name:         vol.Name,
			Description:  &vol.Name,
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
func (src *BuildStrategy) ConvertFrom(ctx context.Context, obj *unstructured.Unstructured) error {
	var br v1alpha1.BuildStrategy

	unstructured := obj.UnstructuredContent()
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructured, &br)
	if err != nil {
		ctxlog.Error(ctx, err, "failed unstructuring the buildrun convertedObject")
	}

	src.ObjectMeta = br.ObjectMeta
	src.TypeMeta = br.TypeMeta
	src.TypeMeta.APIVersion = BETA_GROUP_VERSION

	src.Spec.Steps = []Step{}
	for _, brStep := range br.Spec.BuildSteps {

		step := Step{
			Name:            brStep.Name,
			Image:           brStep.Image,
			Command:         brStep.Command,
			Args:            brStep.Args,
			WorkingDir:      brStep.WorkingDir,
			Env:             brStep.Env,
			Resources:       brStep.Resources,
			VolumeMounts:    brStep.VolumeMounts,
			ImagePullPolicy: brStep.ImagePullPolicy,
		}

		if brStep.SecurityContext != nil {
			step.SecurityContext = brStep.SecurityContext
		}

		src.Spec.Steps = append(src.Spec.Steps, step)
	}

	src.Spec.Parameters = []Parameter{}
	for _, param := range br.Spec.Parameters {
		src.Spec.Parameters = append(src.Spec.Parameters, Parameter{
			Name:        param.Name,
			Description: param.Description,
			Type:        ParameterType(param.Type),
			Default:     param.Default,
			Defaults:    param.Defaults,
		})
	}

	if br.Spec.SecurityContext != nil {
		src.Spec.SecurityContext = (*BuildStrategySecurityContext)(br.Spec.SecurityContext)
	}

	src.Spec.Volumes = []BuildStrategyVolume{}
	for _, vol := range br.Spec.Volumes {
		src.Spec.Volumes = append(src.Spec.Volumes, BuildStrategyVolume{
			Overridable:  vol.Overridable,
			Name:         vol.Name,
			Description:  &vol.Name,
			VolumeSource: vol.VolumeSource,
		})
	}

	return nil
}
