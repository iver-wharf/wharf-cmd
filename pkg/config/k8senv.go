/*
Copyright 2015 The Kubernetes Authors.

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

// Originally taken from k8s.io/api@v0.23.3/core/v1/types.go
// This has since been modified to fit our use-case

package config

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// K8sEnvVar represents an environment variable present in a Container.
type K8sEnvVar struct {
	// Name of the environment variable. Must be a C_IDENTIFIER.
	Name string

	// Optional: no more than one of the following may be specified.

	// Variable references $(VAR_NAME) are expanded
	// using the previously defined environment variables in the container and
	// any service environment variables. If a variable cannot be resolved,
	// the reference in the input string will be unchanged. Double $$ are reduced
	// to a single $, which allows for escaping the $(VAR_NAME) syntax: i.e.
	// "$$(VAR_NAME)" will produce the string literal "$(VAR_NAME)".
	// Escaped references will never be expanded, regardless of whether the variable
	// exists or not.
	// Defaults to "".
	Value string
	// Source for the environment variable's value. Cannot be used if value is not empty.
	ValueFrom *K8sEnvVarSource
}

// AsV1 returns the Kubernetes k8s.io/api/core/v1 type for this config.
// An error is returned on any parse issues.
func (e *K8sEnvVar) AsV1() (*v1.EnvVar, error) {
	if e == nil {
		return nil, nil
	}
	valueFrom, err := e.ValueFrom.AsV1()
	if err != nil {
		return nil, err
	}
	return &v1.EnvVar{
		Name:      e.Name,
		Value:     e.Value,
		ValueFrom: valueFrom,
	}, nil
}

// K8sEnvVarSource represents a source for the value of an EnvVar.
type K8sEnvVarSource struct {
	// Selects a field of the pod: supports metadata.name, metadata.namespace, `metadata.labels['<KEY>']`, `metadata.annotations['<KEY>']`,
	// spec.nodeName, spec.serviceAccountName, status.hostIP, status.podIP, status.podIPs.
	FieldRef *K8sObjectFieldSelector
	// Selects a resource of the container: only resources limits and requests
	// (limits.cpu, limits.memory, limits.ephemeral-storage, requests.cpu, requests.memory and requests.ephemeral-storage) are currently supported.
	ResourceFieldRef *K8sResourceFieldSelector
	// Selects a key of a ConfigMap.
	ConfigMapKeyRef *K8sConfigMapKeySelector
	// Selects a key of a secret in the pod's namespace
	SecretKeyRef *K8sSecretKeySelector
}

// AsV1 returns the Kubernetes k8s.io/api/core/v1 type for this config.
// An error is returned on any parse issues.
func (e *K8sEnvVarSource) AsV1() (*v1.EnvVarSource, error) {
	if e == nil {
		return nil, nil
	}
	resourceFieldRef, err := e.ResourceFieldRef.AsV1()
	if err != nil {
		return nil, err
	}
	return &v1.EnvVarSource{
		FieldRef:         e.FieldRef.AsV1(),
		ResourceFieldRef: resourceFieldRef,
		ConfigMapKeyRef:  e.ConfigMapKeyRef.AsV1(),
		SecretKeyRef:     e.SecretKeyRef.AsV1(),
	}, nil
}

// K8sObjectFieldSelector selects an APIVersioned field of an object.
type K8sObjectFieldSelector struct {
	// Version of the schema the FieldPath is written in terms of, defaults to "v1".
	APIVersion string
	// Path of the field to select in the specified API version.
	FieldPath string
}

// AsV1 returns the Kubernetes k8s.io/api/core/v1 type for this config.
func (e *K8sObjectFieldSelector) AsV1() *v1.ObjectFieldSelector {
	if e == nil {
		return nil
	}
	return &v1.ObjectFieldSelector{
		APIVersion: e.APIVersion,
		FieldPath:  e.FieldPath,
	}
}

// K8sResourceFieldSelector represents container resources (cpu, memory) and their output format
type K8sResourceFieldSelector struct {
	// Container name: required for volumes, optional for env vars
	ContainerName string
	// Required: resource to select
	Resource string
	// Specifies the output format of the exposed resources, defaults to "1"
	Divisor string
}

// AsV1 returns the Kubernetes k8s.io/api/core/v1 type for this config.
// An error is returned on any parse issues.
func (e *K8sResourceFieldSelector) AsV1() (*v1.ResourceFieldSelector, error) {
	if e == nil {
		return nil, nil
	}
	divisor, err := resource.ParseQuantity(e.Divisor)
	if err != nil {
		return nil, err
	}
	return &v1.ResourceFieldSelector{
		ContainerName: e.ContainerName,
		Resource:      e.ContainerName,
		Divisor:       divisor,
	}, nil
}

// K8sConfigMapKeySelector selects a key from a ConfigMap.
type K8sConfigMapKeySelector struct {
	// Name of the referent.
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
	Name string
	// The key to select.
	Key string
	// Specify whether the ConfigMap or its key must be defined
	Optional *bool
}

// AsV1 returns the Kubernetes k8s.io/api/core/v1 type for this config.
func (e *K8sConfigMapKeySelector) AsV1() *v1.ConfigMapKeySelector {
	if e == nil {
		return nil
	}
	return &v1.ConfigMapKeySelector{
		LocalObjectReference: v1.LocalObjectReference{
			Name: e.Name,
		},
		Key:      e.Key,
		Optional: e.Optional,
	}
}

// K8sSecretKeySelector selects a key of a Secret.
type K8sSecretKeySelector struct {
	// Name of the referent.
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
	Name string
	// The key of the secret to select from.  Must be a valid secret key.
	Key string
	// Specify whether the Secret or its key must be defined
	Optional *bool
}

// AsV1 returns the Kubernetes k8s.io/api/core/v1 type for this config.
func (e *K8sSecretKeySelector) AsV1() *v1.SecretKeySelector {
	if e == nil {
		return nil
	}
	return &v1.SecretKeySelector{
		LocalObjectReference: v1.LocalObjectReference{
			Name: e.Name,
		},
		Key:      e.Key,
		Optional: e.Optional,
	}
}

// K8sLocalObjectReference contains enough information to let you locate the
// referenced object inside the same namespace.
type K8sLocalObjectReference struct {
	// Name of the referent.
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
	Name string
}

// AsV1 returns the Kubernetes k8s.io/api/core/v1 type for this config.
func (e *K8sLocalObjectReference) AsV1() *v1.LocalObjectReference {
	if e == nil {
		return nil
	}
	return &v1.LocalObjectReference{
		Name: e.Name,
	}
}
