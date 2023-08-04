// Copyright The Shipwright Contributors
//
// SPDX-License-Identifier: Apache-2.0
package conversion

import (
	"context"
	"fmt"

	"github.com/shipwright-io/build/pkg/apis/build/v1beta1"
	"github.com/shipwright-io/build/pkg/ctxlog"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	BETA_GROUP_VERSION  = "shipwright.io/v1beta1"
	ALPHA_GROUP_VERSION = "shipwright.io/v1alpha1"
	BUILD_KIND          = "Build"
	BUILDRUN_KIND       = "BuildRun"
	KIND                = "kind"
)

// convertSHPCR takes an unstructured object with certain CR apiversion, parses it to a known Object type,
// modify the type to a desired version of that type, and converts it back to unstructured
func convertSHPCR(Object *unstructured.Unstructured, toVersion string, ctx context.Context) (*unstructured.Unstructured, metav1.Status) {
	ctxlog.Info(ctx, "converting custom resource")

	convertedObject := Object.DeepCopy()
	fromVersion := Object.GetAPIVersion()

	if fromVersion == toVersion {
		ctxlog.Info(ctx, "nothing to convert")
		return convertedObject, statusSucceed()
	}

	switch Object.GetAPIVersion() {
	case BETA_GROUP_VERSION:
		switch toVersion {

		case ALPHA_GROUP_VERSION:
			if convertedObject.Object[KIND] == BUILD_KIND {

				unstructured := convertedObject.UnstructuredContent()
				var build v1beta1.Build
				err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructured, &build)
				if err != nil {
					ctxlog.Error(ctx, err, "failed unstructuring the build convertedObject")
				}
				build.ConvertTo(ctx, convertedObject)

			} else if convertedObject.Object[KIND] == BUILDRUN_KIND {
				unstructured := convertedObject.UnstructuredContent()
				var buildRun v1beta1.BuildRun
				err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructured, &buildRun)
				if err != nil {
					ctxlog.Error(ctx, err, "failed unstructuring the buildRun convertedObject")
				}
				buildRun.ConvertTo(ctx, convertedObject)

			} else {
				return nil, statusErrorWithMessage("unsupporteddda skdksdjkjsd Kind")
			}
		default:
			return nil, statusErrorWithMessage("unexpected conversion version to %q", toVersion)
		}
	case ALPHA_GROUP_VERSION:
		switch toVersion {
		case BETA_GROUP_VERSION:
			if convertedObject.Object[KIND] == BUILD_KIND {

				var buildBeta v1beta1.Build

				buildBeta.ConvertFrom(ctx, convertedObject)

				mapito, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&buildBeta)
				if err != nil {
					ctxlog.Error(ctx, err, "failed structuring the newObject")
				}
				convertedObject.Object = mapito

			} else if convertedObject.Object[KIND] == BUILDRUN_KIND {
				var buildRunBeta v1beta1.BuildRun

				buildRunBeta.ConvertFrom(ctx, convertedObject)

				mapito, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&buildRunBeta)
				if err != nil {
					ctxlog.Error(ctx, err, "failed structuring the newObject")
				}
				convertedObject.Object = mapito
			} else {
				return nil, statusErrorWithMessage("unsupported Kind")
			}
		default:
			return nil, statusErrorWithMessage("unexpected conversion version to %q", toVersion)
		}
	default:
		return nil, statusErrorWithMessage("unexpected conversion version from %q", fromVersion)
	}
	return convertedObject, statusSucceed()
}

func statusErrorWithMessage(msg string, params ...interface{}) metav1.Status {
	return metav1.Status{
		Message: fmt.Sprintf(msg, params...),
		Status:  metav1.StatusFailure,
	}
}
