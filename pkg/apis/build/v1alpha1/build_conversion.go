// Copyright The Shipwright Contributors
//
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

// ConvertTo converts this Build to the Hub version (v1alpha1)
func (src *Build) ConvertFrom(bs *Build) error {
	bs.ObjectMeta = src.ObjectMeta
	return nil
}
