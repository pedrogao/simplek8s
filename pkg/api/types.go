/*
Copyright 2014 Google Inc. All rights reserved.

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

// Package api includes all types used to communicate between the various
// parts of the Kubernetes system.
package api

// ContainerManifest corresponds to the Container Manifest format, documented at:
// https://developers.google.com/compute/docs/containers#container_manifest
// This is used as the representation of Kubernete's workloads.
type ContainerManifest struct {
	Version    string      `yaml:"version" json:"version"`
	Volumes    []Volume    `yaml:"volumes" json:"volumes"`
	Containers []Container `yaml:"containers" json:"containers"`
	Id         string      `yaml:"id,omitempty" json:"id,omitempty"`
}

type Volume struct {
	Name string `yaml:"name" json:"name"`
}

type Port struct {
	Name          string `yaml:"name,omitempty" json:"name,omitempty"`
	HostPort      int    `yaml:"hostPort,omitempty" json:"hostPort,omitempty"`
	ContainerPort int    `yaml:"containerPort,omitempty" json:"containerPort,omitempty"`
	Protocol      string `yaml:"protocol,omitempty" json:"protocol,omitempty"`
}

type VolumeMount struct {
	Name      string `yaml:"name,omitempty" json:"name,omitempty"`
	ReadOnly  bool   `yaml:"readOnly,omitempty" json:"readOnly,omitempty"`
	MountPath string `yaml:"mountPath,omitempty" json:"mountPath,omitempty"`
}

type EnvVar struct {
	Name  string `yaml:"name,omitempty" json:"name,omitempty"`
	Value string `yaml:"value,omitempty" json:"value,omitempty"`
}

// Container represents a single container that is expected to be run on the host.
type Container struct {
	Name         string        `yaml:"name,omitempty" json:"name,omitempty"`
	Image        string        `yaml:"image,omitempty" json:"image,omitempty"`
	Command      string        `yaml:"command,omitempty" json:"command,omitempty"`
	WorkingDir   string        `yaml:"workingDir,omitempty" json:"workingDir,omitempty"`
	Ports        []Port        `yaml:"ports,omitempty" json:"ports,omitempty"`
	Env          []EnvVar      `yaml:"env,omitempty" json:"env,omitempty"`
	Memory       int           `yaml:"memory,omitempty" json:"memory,omitempty"`
	CPU          int           `yaml:"cpu,omitempty" json:"cpu,omitempty"`
	VolumeMounts []VolumeMount `yaml:"volumeMounts,omitempty" json:"volumeMounts,omitempty"`
}

// Event is the representation of an event logged to etcd backends
type Event struct {
	Event     string             `json:"event,omitempty"`
	Manifest  *ContainerManifest `json:"manifest,omitempty"`
	Container *Container         `json:"container,omitempty"`
	Timestamp int64              `json:"timestamp"`
}

// The below types are used by kube_client and api_server.

// JSONBase is shared by all objects sent to, or returned from the client
type JSONBase struct {
	Kind              string `json:"kind,omitempty" yaml:"kind,omitempty"`
	ID                string `json:"id,omitempty" yaml:"id,omitempty"`
	CreationTimestamp string `json:"creationTimestamp,omitempty" yaml:"creationTimestamp,omitempty"`
	SelfLink          string `json:"selfLink,omitempty" yaml:"selfLink,omitempty"`
}

// TaskState is the state of a task, used as either input (desired state) or output (current state)
type TaskState struct {
	Manifest ContainerManifest `json:"manifest,omitempty" yaml:"manifest,omitempty"`
	Status   string            `json:"status,omitempty" yaml:"status,omitempty"`
	Host     string            `json:"host,omitempty" yaml:"host,omitempty"`
	HostIP   string            `json:"hostIP,omitempty" yaml:"hostIP,omitempty"`
	Info     interface{}       `json:"info,omitempty" yaml:"info,omitempty"`
}

type TaskList struct {
	JSONBase
	Items []Task `json:"items" yaml:"items,omitempty"`
}

// Task is a single task, used as either input (create, update) or as output (list, get)
type Task struct {
	JSONBase
	Labels       map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	DesiredState TaskState         `json:"desiredState,omitempty" yaml:"desiredState,omitempty"`
	CurrentState TaskState         `json:"currentState,omitempty" yaml:"currentState,omitempty"`
}

// ReplicationControllerState is the state of a replication controller, either input (create, update) or as output (list, get)
type ReplicationControllerState struct {
	Replicas      int               `json:"replicas" yaml:"replicas"`
	ReplicasInSet map[string]string `json:"replicasInSet,omitempty" yaml:"replicasInSet,omitempty"`
	TaskTemplate  TaskTemplate      `json:"taskTemplate,omitempty" yaml:"taskTemplate,omitempty"`
}

type ReplicationControllerList struct {
	JSONBase
	Items []ReplicationController `json:"items,omitempty" yaml:"items,omitempty"`
}

// ReplicationController represents the configuration of a replication controller
type ReplicationController struct {
	JSONBase
	DesiredState ReplicationControllerState `json:"desiredState,omitempty" yaml:"desiredState,omitempty"`
	Labels       map[string]string          `json:"labels,omitempty" yaml:"labels,omitempty"`
}

// TaskTemplate holds the information used for creating tasks
type TaskTemplate struct {
	DesiredState TaskState         `json:"desiredState,omitempty" yaml:"desiredState,omitempty"`
	Labels       map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
}

// ServiceList holds a list of services
type ServiceList struct {
	Items []Service `json:"items" yaml:"items"`
}

// Defines a service abstraction by a name (for example, mysql) consisting of local port
// (for example 3306) that the proxy listens on, and the labels that define the service.
type Service struct {
	JSONBase
	Port   int               `json:"port,omitempty" yaml:"port,omitempty"`
	Labels map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
}

// Defines the endpoints that implement the actual service, for example:
// Name: "mysql", Endpoints: ["10.10.1.1:1909", "10.10.2.2:8834"]
type Endpoints struct {
	Name      string
	Endpoints []string
}
