// Copyright The Shipwright Contributors
//
// SPDX-License-Identifier: Apache-2.0

package v1beta1

import (
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

// ConvertTo converts this Build to the Hub version (v1alpha1)
func (src *Build) ConvertTo(dstRaw conversion.Hub) error {
	// dst := dstRaw.(*v1alpha1.Build)

	return nil
}

// ConvertFrom converts from the Hub version (v1alpha1) to this version.
// TODO: Not needed?
func (dst *Build) ConvertFrom(srcRaw conversion.Hub) error {
	return nil
}
