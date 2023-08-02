// Copyright The Shipwright Contributors
//
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"context"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func (src *Build) ConvertTo(ctx context.Context, obj *unstructured.Unstructured) error {
	return nil
}

func (src *Build) ConvertFrom(ctx context.Context, obj *unstructured.Unstructured) error {
	return nil
}
