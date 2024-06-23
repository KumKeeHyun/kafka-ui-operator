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
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var rolelog = logf.Log.WithName("role-resource")

// SetupWebhookWithManager will setup the manager to manage the webhooks
func (r *Role) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/validate-kafka-ui-kumkeehyun-github-com-v1-role,mutating=false,failurePolicy=fail,sideEffects=None,groups=kafka-ui.kumkeehyun.github.com,resources=roles,verbs=create;update,versions=v1,name=vrole.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Role{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Role) ValidateCreate() (admission.Warnings, error) {
	rolelog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return nil, r.validateRole()
}

func (r *Role) validateRole() error {
	var allErrs field.ErrorList

	if err := r.validateRoleSpec(); err != nil {
		allErrs = append(allErrs, err...)
	}

	if len(allErrs) != 0 {
		return apierrors.NewInvalid(
			schema.GroupKind{Group: "kafka-ui.kumkeehyun.github.com", Kind: "Role"},
			r.Name,
			allErrs,
		)
	}
	return nil
}

func (r *Role) validateRoleSpec() field.ErrorList {
	var specErrs field.ErrorList
	specPath := field.NewPath("spec")

	if err := r.validateRoleSpecPermissions(specPath.Child("permissions")); err != nil {
		specErrs = append(specErrs, err...)
	}

	return specErrs
}

func (r *Role) validateRoleSpecPermissions(fieldPath *field.Path) field.ErrorList {
	var permissionsErrs field.ErrorList

	for i, permission := range r.Spec.Permissions {
		if err := validateRolePermission(permission, fieldPath.Index(i)); err != nil {
			permissionsErrs = append(permissionsErrs, err)
		}
	}

	return permissionsErrs
}

func validateRolePermission(permission Permission, fieldPath *field.Path) *field.Error {
	if _, ok := ResourceActions[permission.Resource]; !ok {
		return field.Invalid(fieldPath.Child("resource"), permission.Resource, "invalid resource")
	}

	actionsPath := fieldPath.Child("actions")
	for i, action := range permission.Actions {
		if _, exists := ResourceActions[permission.Resource][action]; !exists {
			return field.Invalid(actionsPath.Index(i), action, "invalid action")
		}
	}

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Role) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	rolelog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil, r.validateRole()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Role) ValidateDelete() (admission.Warnings, error) {
	rolelog.Info("validate delete", "name", r.Name)

	return nil, nil
}

var (
	ResourceActions = map[Resource]map[Action]struct{}{
		APPLICATIONCONFIG: applicationConfigActions,
		CLUSTERCONFIG:     clusterConfigActions,
		TOPIC:             topicActions,
		CONSUMER:          consumerActions,
		SCHEMA:            schemaActions,
		CONNECT:           connectActions,
		KSQL:              ksqlActions,
		ACL:               aclActions,
		AUDIT:             auditActions,
	}
	applicationConfigActions = map[Action]struct{}{
		ALL:  {},
		VIEW: {},
		EDIT: {},
	}
	clusterConfigActions = map[Action]struct{}{
		ALL:  {},
		VIEW: {},
		EDIT: {},
	}
	topicActions = map[Action]struct{}{
		ALL:             {},
		VIEW:            {},
		EDIT:            {},
		CREATE:          {},
		DELETE:          {},
		MESSAGE_READ:    {},
		MESSAGE_PRODUCE: {},
		MESSAGE_DELETE:  {},
	}
	consumerActions = map[Action]struct{}{
		ALL:           {},
		VIEW:          {},
		DELETE:        {},
		RESET_OFFSETS: {},
	}
	schemaActions = map[Action]struct{}{
		ALL:                         {},
		VIEW:                        {},
		EDIT:                        {},
		CREATE:                      {},
		DELETE:                      {},
		MODIFY_GLOBAL_COMPATIBILITY: {},
	}
	connectActions = map[Action]struct{}{
		ALL:     {},
		VIEW:    {},
		EDIT:    {},
		CREATE:  {},
		RESTART: {},
	}
	ksqlActions = map[Action]struct{}{
		ALL:     {},
		EXECUTE: {},
	}
	aclActions = map[Action]struct{}{
		ALL:  {},
		VIEW: {},
		EDIT: {},
	}
	auditActions = map[Action]struct{}{
		ALL:  {},
		VIEW: {},
	}
)
