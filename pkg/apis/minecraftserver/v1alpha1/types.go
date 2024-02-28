/*
Copyright 2017 The Kubernetes Authors.

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

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MinecraftServer is a specification for a MinecraftServer resource
type MinecraftServer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MinecraftServerSpec   `json:"spec"`
	Status MonecraftServerStatus `json:"status"`
}

// MinecraftServerSpec is a desired state description of MinecraftServer
type MinecraftServerSpec struct {
	// ServerVersion sets the version of the Minecraft server to run. This value is required and must be
	// a valid Minecraft server version.
	// Examples of valid values are "1.16.5" and "1.17.1" and "1.18.1.62".
	ServerVersion string `json:"serverVersion"`
	// ServiceAccountName is the name of the ServiceAccount to use to run the Minecraft server.
	ServiceAccountName string `json:"serviceAccountName,omitempty"`
	// CreateService is a flag that indicates whether or not to create a Service for the Minecraft server.
	// A service is required to expose the Minecraft server to the internet.
	CreateService bool `json:"createService,omitempty"`
	// RestartPolicy is the restart policy to apply to the Minecraft server pod.
	RestartPolicy string `json:"restartPolicy,omitempty"`
	// DataVolume is defines a Volume Claim Template that will be used to generate
	// a Persistent Volume Claim for persisting Minecraft server data. If not provided,
	// a persistent volume will not be created.
	DataVolume *MinecraftPersistentVolume `json:"dataVolume,omitempty"`
	// Replicas sets the number of replicas to run for the Minecraft server.
	Replicas int32 `json:"replicas,omitempty"`
	// SecurityContext is the security context to apply to the Minecraft server pod.
	SecurityContext *corev1.PodSecurityContext `json:"securityContext,omitempty"`
	// ServerSettings is a set of settings to apply to the Minecraft server.
	ServerSettings MinecraftServerSettings `json:"serverSettings"`
}

type MinecraftPersistentVolume struct {
	PersistentVolumeClaim corev1.PersistentVolumeClaimSpec `json:"persistentVolumeClaim,omitempty"`
	Volume                corev1.Volume                    `json:"volume,omitempty"`
}

type MinecraftServerSettings struct {
	LevelSeed  string `json:"levelSeed"`
	GameMode   string `json:"gameMode"`
	Difficulty string `json:"difficulty"`
	MaxPlayers int32  `json:"maxPlayers"`
	LevelName  string `json:"levelName"`
}

type MonecraftServerStatus struct {
	AvailableReplicas int32 `json:"availableReplicas"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MinecraftServerList is a list of MinecraftServer resources
type MinecraftServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []MinecraftServer `json:"items"`
}
