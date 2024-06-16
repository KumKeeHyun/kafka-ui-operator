package controller

import (
	"github.com/KumKeeHyun/kafka-ui-operator/pkg/funcutil"
	v1 "k8s.io/api/core/v1"
	"strings"
)

func findLatestClusterConfigMapOrEmpty(configMaps *v1.ConfigMapList) v1.ConfigMap {
	clusterConfigMaps := funcutil.Filter(configMaps.Items, func(cm v1.ConfigMap) bool {
		return strings.HasPrefix(cm.Name, "cluster")
	})
	if len(clusterConfigMaps) == 0 {
		return v1.ConfigMap{}
	}

	latestClusterConfigMap := clusterConfigMaps[0]
	for _, clusterConfigMap := range clusterConfigMaps {
		if clusterConfigMap.CreationTimestamp.After(latestClusterConfigMap.CreationTimestamp.Time) {
			latestClusterConfigMap = clusterConfigMap
		}
	}
	return latestClusterConfigMap
}

func findLatestRoleConfigMapOrEmpty(configMaps *v1.ConfigMapList) v1.ConfigMap {
	roleConfigMaps := funcutil.Filter(configMaps.Items, func(cm v1.ConfigMap) bool {
		return strings.HasPrefix(cm.Name, "role")
	})
	if len(roleConfigMaps) == 0 {
		return v1.ConfigMap{}
	}

	latestRoleConfigMap := roleConfigMaps[0]
	for _, roleConfigMap := range roleConfigMaps {
		if roleConfigMap.CreationTimestamp.After(latestRoleConfigMap.CreationTimestamp.Time) {
			latestRoleConfigMap = roleConfigMap
		}
	}
	return latestRoleConfigMap
}
