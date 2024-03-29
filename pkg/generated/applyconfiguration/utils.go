/*
Copyright The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by applyconfiguration-gen. DO NOT EDIT.

package applyconfiguration

import (
	v1alpha1 "MinecraftServerController/pkg/apis/minecraftserver/v1alpha1"
	minecraftserverv1alpha1 "MinecraftServerController/pkg/generated/applyconfiguration/minecraftserver/v1alpha1"

	schema "k8s.io/apimachinery/pkg/runtime/schema"
)

// ForKind returns an apply configuration type for the given GroupVersionKind, or nil if no
// apply configuration type exists for the given GroupVersionKind.
func ForKind(kind schema.GroupVersionKind) interface{} {
	switch kind {
	// Group=minecraftserver.k8s.io, Version=v1alpha1
	case v1alpha1.SchemeGroupVersion.WithKind("MinecraftPersistentVolume"):
		return &minecraftserverv1alpha1.MinecraftPersistentVolumeApplyConfiguration{}
	case v1alpha1.SchemeGroupVersion.WithKind("MinecraftServer"):
		return &minecraftserverv1alpha1.MinecraftServerApplyConfiguration{}
	case v1alpha1.SchemeGroupVersion.WithKind("MinecraftServerSettings"):
		return &minecraftserverv1alpha1.MinecraftServerSettingsApplyConfiguration{}
	case v1alpha1.SchemeGroupVersion.WithKind("MinecraftServerSpec"):
		return &minecraftserverv1alpha1.MinecraftServerSpecApplyConfiguration{}
	case v1alpha1.SchemeGroupVersion.WithKind("MonecraftServerStatus"):
		return &minecraftserverv1alpha1.MonecraftServerStatusApplyConfiguration{}

	}
	return nil
}
