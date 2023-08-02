// Copyright The Shipwright Contributors
//
// SPDX-License-Identifier: Apache-2.0

package v1beta1

import (
	"context"

	"github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
	"github.com/shipwright-io/build/pkg/ctxlog"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// To Alpha
func (src *BuildRun) ConvertTo(ctx context.Context, obj *unstructured.Unstructured) {

}

// From Alpha
func (src *BuildRun) ConvertFrom(ctx context.Context, obj *unstructured.Unstructured) {
	unstructured := obj.UnstructuredContent()
	var build v1alpha1.Build

	err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructured, &build)
	if err != nil {
		ctxlog.Error(ctx, err, "failed unstructuring the convertedObject")
	}

	// you have a alpha build, which you will set in beta

}

// // ConvertTo converts this BuildRun to the Hub version (v1alpha1)
// func (src *BuildRun) ConvertTo(dstRaw conversion.Hub) error {
// 	dst := dstRaw.(*v1alpha1.BuildRun)
// 	dst.ObjectMeta = src.ObjectMeta

// 	// BuildRunSpec BuildSpec
// 	newBuildSpec := v1alpha1.BuildSpec{}
// 	if err := src.Spec.Build.Build.ConvertTo(&newBuildSpec); err != nil {
// 		return err
// 	}

// 	dst.Spec.BuildSpec = &newBuildSpec

// 	// BuildRunSpec BuildRef
// 	dst.Spec.BuildRef = &v1alpha1.BuildRef{
// 		Name: src.Spec.Build.Name,
// 	}

// 	// BuildRunSpec ServiceAccount
// 	dst.Spec.ServiceAccount = &v1alpha1.ServiceAccount{
// 		Name: src.Spec.ServiceAccount,
// 	}

// 	// BuildRunSpec Timeout
// 	dst.Spec.Timeout = src.Spec.Timeout

// 	// BuildRunSpec ParamValues
// 	dst.Spec.ParamValues = nil
// 	for _, p := range src.Spec.ParamValues {
// 		new := v1alpha1.ParamValue{}
// 		p.convertTo(&new)
// 		dst.Spec.ParamValues = append(dst.Spec.ParamValues, new)
// 	}

// 	// BuildRunSpec Image
// 	dst.Spec.Output = &v1alpha1.Image{
// 		Image: src.Spec.Output.Image,
// 		Credentials: &corev1.LocalObjectReference{
// 			Name: *src.Spec.Output.PushSecret,
// 		},
// 		Annotations: src.Spec.Output.Annotations,
// 		Labels:      src.Spec.Output.Labels,
// 	}

// 	// BuildRunSpec State
// 	dst.Spec.State = (*v1alpha1.BuildRunRequestedState)(src.Spec.State)

// 	// BuildRunSpec Env
// 	dst.Spec.Env = src.Spec.Env

// 	// BuildRunSpec Retention
// 	dst.Spec.Retention = (*v1alpha1.BuildRunRetention)(src.Spec.Retention)

// 	// BuildRunSpec Volumes
// 	for i, vol := range src.Spec.Volumes {
// 		dst.Spec.Volumes[i].Name = vol.Name
// 		dst.Spec.Volumes[i].VolumeSource = vol.VolumeSource

// 	}
// 	return nil
// }

// // ConvertFrom converts from the Hub version (v1alpha1) to this version.
// // TODO: Not needed?
// func (dst *BuildRun) ConvertFrom(srcRaw conversion.Hub) error {
// 	return nil
// }
