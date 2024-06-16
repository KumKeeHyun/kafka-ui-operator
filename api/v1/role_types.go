/*
Copyright 2024.

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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// RoleSpec defines the desired state of Role
type RoleSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of Role. Edit role_types.go to remove/update
	RoleProperties `json:",inline"`
}

type RoleProperties struct {
	Name        string       `json:"name" yaml:"name"`
	Clusters    []string     `json:"clusters" yaml:"clusters"`
	Subjects    []Subject    `json:"subjects" yaml:"subjects"`
	Permissions []Permission `json:"permissions" yaml:"permissions"`
}

type Subject struct {
	Provider Provider `json:"provider" yaml:"provider"`
	Type     string   `json:"type" yaml:"type"`
	Value    string   `json:"value" yaml:"value"`
}

// Provider is the type of authentication provider
// +enum
type Provider string

const (
	OAUTH         Provider = "oauth"
	OAUTH_GOOGLE  Provider = "oauth_google"
	OAUTH_GITHUB  Provider = "oauth_github"
	OAUTH_COGNITO Provider = "oauth_cognito"
	LDAP          Provider = "ldap"
	LDAP_AD       Provider = "ldap_ad"
)

type Permission struct {
	Resource Resource `json:"resource" yaml:"resource"`
	Actions  []Action `json:"actions" yaml:"actions"`
	// +optional
	Value string `json:"value,omitempty" yaml:"value,omitempty"`
}

// Resource is the type of resource
// +enum
type Resource string

const (
	APPLICATIONCONFIG Resource = "applicationconfig"
	CLUSTERCONFIG     Resource = "clusterconfig"
	TOPIC             Resource = "topic"
	CONSUMER          Resource = "consumer"
	SCHEMA            Resource = "schema"
	CONNECT           Resource = "connect"
	KSQL              Resource = "ksql"
	ACL               Resource = "acl"
	AUDIT             Resource = "audit"
)

// Action is the type of action
// +enum
type Action string

const (
	ALL Action = "all"

	// VIEW acl, application config, audit, cluster config, connect, consumer group, schema, topic
	VIEW Action = "view"
	// EDIT acl, application config, cluster config, connect, schema, topic
	EDIT Action = "edit"
	// CREATE connect, schema, topic
	CREATE Action = "create"
	// DELETE consumer group, schema, topic
	DELETE Action = "delete"
	// RESTART connect
	RESTART Action = "restart"
	// RESET_OFFSETS consumer group
	RESET_OFFSETS Action = "reset_offsets"
	// EXECUTE ksql
	EXECUTE Action = "execute"
	// MESSAGE_READ topic
	MESSAGE_READ Action = "message_read"
	// MESSAGE_PRODUCE topic
	MESSAGE_PRODUCE Action = "message_produce"
	// MESSAGE_DELETE topic
	MESSAGE_DELETE Action = "message_delete"
	// schema
	MODIFY_GLOBAL_COMPATIBILITY Action = "modify_global_compatibility"
)

// RoleStatus defines the observed state of Role
type RoleStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Role is the Schema for the roles API
type Role struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RoleSpec   `json:"spec,omitempty"`
	Status RoleStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// RoleList contains a list of Role
type RoleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Role `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Role{}, &RoleList{})
}
