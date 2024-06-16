package kafkaui

import (
	kafkauiv1 "github.com/KumKeeHyun/kafka-ui-operator/api/v1"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/third_party/forked/golang/reflect"
)

type KafkaProperties struct {
	Kafka ClustersProperties `json:"kafka" yaml:"kafka"`
}

type ClustersProperties struct {
	Clusters []kafkauiv1.ClusterProperties `json:"clusters" yaml:"clusters,omitempty"`
}

func NewKafkaPropertiesFromClusters(clusters []kafkauiv1.Cluster) *KafkaProperties {
	property := &KafkaProperties{}
	property.Kafka.Clusters = make([]kafkauiv1.ClusterProperties, 0, len(clusters))

	for _, cluster := range clusters {
		property.Kafka.Clusters = append(property.Kafka.Clusters, cluster.Spec.ClusterProperties)
	}
	return property
}

func (p *KafkaProperties) MustMarshalToYaml() string {
	yamlBytes, err := yaml.Marshal(p)
	if err != nil {
		panic(err)
	}
	return string(yamlBytes)
}

func (p *KafkaProperties) MustUnmarshalFromYaml(yamlStr string) {
	err := yaml.Unmarshal([]byte(yamlStr), p)
	if err != nil {
		panic(err)
	}
}

func (p *KafkaProperties) MatchWithClusters(clusters []kafkauiv1.Cluster) bool {
	if len(p.Kafka.Clusters) != len(clusters) {
		return false
	}

	e := reflect.Equalities{}
	for _, cluster := range clusters {
		found := false
		for _, propertyCluster := range p.Kafka.Clusters {
			if propertyCluster.Name == cluster.Spec.ClusterProperties.Name {
				found = e.DeepEqual(propertyCluster, cluster.Spec.ClusterProperties)
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

type RBACProperties struct {
	RBAC RolesProperties `json:"rbac" yaml:"rbac"`
}

type RolesProperties struct {
	Roles []kafkauiv1.RoleProperties `json:"roles" yaml:"roles,omitempty"`
}

func NewRBACPropertiesFromRoles(roles []kafkauiv1.Role) *RBACProperties {
	property := &RBACProperties{}
	property.RBAC.Roles = make([]kafkauiv1.RoleProperties, 0, len(roles))

	for _, role := range roles {
		property.RBAC.Roles = append(property.RBAC.Roles, role.Spec.RoleProperties)
	}
	return property
}

func (p *RBACProperties) MustMarshalToYaml() string {
	yamlBytes, err := yaml.Marshal(p)
	if err != nil {
		panic(err)
	}
	return string(yamlBytes)
}

func (p *RBACProperties) MustUnmarshalFromYaml(yamlStr string) {
	err := yaml.Unmarshal([]byte(yamlStr), p)
	if err != nil {
		panic(err)
	}
}

func (p *RBACProperties) MatchWithRoles(crdRoles []kafkauiv1.Role) bool {
	if len(p.RBAC.Roles) != len(crdRoles) {
		return false
	}

	e := reflect.Equalities{}
	for _, crdRole := range crdRoles {
		found := false
		for _, propertyRole := range p.RBAC.Roles {
			if propertyRole.Name == crdRole.Spec.Name {
				found = e.DeepEqual(propertyRole, crdRole.Spec.RoleProperties)
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
