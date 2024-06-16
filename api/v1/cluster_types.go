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

// see: https://github.com/provectus/kafka-ui/blob/master/kafka-ui-api/src/main/java/com/provectus/kafka/ui/config/ClustersProperties.java

// ClusterSpec defines the desired state of Cluster
type ClusterSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	ClusterProperties `json:",inline"`
}

type ClusterProperties struct {
	Name             string `json:"name" yaml:"name"`
	BootstrapServers string `json:"bootstrapServers" yaml:"bootstrapServers"`

	// +optional
	SchemaRegistry *string `json:"schemaRegistry,omitempty" yaml:"schemaRegistry,omitempty"`
	// +optional
	SchemaRegistryAuth *SchemaRegistryAuth `json:"schemaRegistryAuth,omitempty" yaml:"schemaRegistryAuth,omitempty"`
	// +optional
	SchemaRegistrySsl *KeystoreConfig `json:"schemaRegistrySsl,omitempty" yaml:"schemaRegistrySsl,omitempty"`

	// +optional
	KsqlDBServer *string `json:"ksqldbServer,omitempty" yaml:"ksqldbServer,omitempty"`
	// +optional
	KsqlDBServerAuth *KsqlDBServerAuth `json:"ksqldbServerAuth,omitempty" yaml:"ksqldbServerAuth,omitempty"`
	// +optional
	KsqlDBServerSsl *KeystoreConfig `json:"ksqldbServerSsl,omitempty" yaml:"ksqldbServerSsl,omitempty"`

	// +optional
	KafkaConnect []ConnectCluster `json:"kafkaConnect,omitempty" yaml:"kafkaConnect,omitempty"`

	// +optional
	Metrics *MetricsConfigData `json:"metrics,omitempty" yaml:"metrics,omitempty"`

	// +default=false
	// +optional
	ReadOnly *bool `json:"readOnly,omitempty" yaml:"readOnly,omitempty"`

	// +optional
	DefaultKeySerde *string `json:"defaultKeySerde,omitempty" yaml:"defaultKeySerde,omitempty"`
	// +optional
	DefaultValueSerde *string `json:"defaultValueSerde,omitempty" yaml:"defaultValueSerde,omitempty"`

	// +optional
	Masking []Masking `json:"masking,omitempty" yaml:"masking,omitempty"`

	// +optional
	PollingThrottleRate *int64 `json:"pollingThrottleRate,omitempty" yaml:"pollingThrottleRate,omitempty"`

	// +optional
	SSL *TruststoreConfig `json:"ssl,omitempty" yaml:"ssl,omitempty"`

	// +optional
	Audit *AuditProperties `json:"audit,omitempty" yaml:"audit,omitempty"`
}

type SchemaRegistryAuth struct {
	Username string `json:"username" yaml:"username"`
	Password string `json:"password" yaml:"password"`
}

type KeystoreConfig struct {
	KeystoreLocation string `json:"keystoreLocation" yaml:"keystoreLocation"`
	KeystorePassword string `json:"keystorePassword" yaml:"keystorePassword"`
}

type KsqlDBServerAuth struct {
	Username string `json:"username" yaml:"username"`
	Password string `json:"password" yaml:"password"`
}

type ConnectCluster struct {
	Name    string `json:"name" yaml:"name"`
	Address string `json:"address" yaml:"address"`
	// +optional
	Username *string `json:"username,omitempty" yaml:"username,omitempty"`
	// +optional
	Password *string `json:"password,omitempty" yaml:"password,omitempty"`
	// +optional
	KeystoreLocation *string `json:"keystoreLocation,omitempty" yaml:"keystoreLocation,omitempty"`
	// +optional
	KeystorePassword *string `json:"keystorePassword,omitempty" yaml:"keystorePassword,omitempty"`
}

type MetricsConfigData struct {
	Type string `json:"type" yaml:"type"`
	Port int32  `json:"port" yaml:"port"`
	// +optional
	SSL bool `json:"ssl,omitempty" yaml:"ssl,omitempty"`
	// +optional
	UserName string `json:"username,omitempty" yaml:"username,omitempty"`
	// +optional
	Password string `json:"password,omitempty" yaml:"password,omitempty"`
	// +optional
	KeystoreLocation string `json:"keystoreLocation,omitempty" yaml:"keystoreLocation,omitempty"`
	// +optional
	KeystorePassword string `json:"keystorePassword,omitempty" yaml:"keystorePassword,omitempty"`
}

type Masking struct {
	Type MaskingType `json:"type" yaml:"type"`
	// +optional
	Fields []string `json:"fields,omitempty" yaml:"fields,omitempty"`
	// +optional
	FieldsNamePattern *string `json:"fieldsNamePattern,omitempty" yaml:"fieldsNamePattern,omitempty"`
	// +optional
	MaskingCharsReplacement []string `json:"maskingCharsReplacement,omitempty" yaml:"maskingCharsReplacement,omitempty"`
	// +optional
	Replacement *string `json:"replacement,omitempty" yaml:"replacement,omitempty"`
	// +optional
	TopicKeyPattern *string `json:"topicKeyPattern,omitempty" yaml:"topicKeyPattern,omitempty"`
	// +optional
	TopicValuePattern *string `json:"topicValuePattern,omitempty" yaml:"topicValuePattern,omitempty"`
}

// MaskingType is the type of masking
// +enum
type MaskingType string

const (
	REMOVE  MaskingType = "remove"
	MASK    MaskingType = "mask"
	REPLACE MaskingType = "replace"
)

type TruststoreConfig struct {
	TruststoreLocation string `json:"truststoreLocation" yaml:"truststoreLocation"`
	TruststorePassword string `json:"truststorePassword" yaml:"truststorePassword"`
}

type AuditProperties struct {
	// +optional
	Topic *string `json:"topic,omitempty" yaml:"topic,omitempty"`
	// +optional
	AuditTopicPartitions *int32 `json:"auditTopicPartitions,omitempty" yaml:"auditTopicPartitions,omitempty"`
	// +optional
	TopicAuditEnabled *bool `json:"topicAuditEnabled,omitempty" yaml:"topicAuditEnabled,omitempty"`
	// +optional
	ConsoleAuditEnabled *bool `json:"consoleAuditEnabled,omitempty" yaml:"consoleAuditEnabled,omitempty"`
	// +optional
	LogLevel *AuditLogLevel `json:"logLevel,omitempty" yaml:"logLevel,omitempty"`
}

// AuditLogLevel is the type of audit log level
// +enum
type AuditLogLevel string

const (
	AUDIT_ALL  AuditLogLevel = "all"
	ALTER_ONLY AuditLogLevel = "alter_only"
)

// ClusterStatus defines the observed state of Cluster
type ClusterStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Cluster is the Schema for the clusters API
type Cluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterSpec   `json:"spec,omitempty"`
	Status ClusterStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ClusterList contains a list of Cluster
type ClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Cluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Cluster{}, &ClusterList{})
}
