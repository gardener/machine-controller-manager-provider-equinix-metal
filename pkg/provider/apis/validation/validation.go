/*
Copyright (c) 2019 SAP SE or an SAP affiliate company. All rights reserved.

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

// Package validation - validation is used to validate cloud specific ProviderSpec
package validation

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	api "github.com/gardener/machine-controller-manager-provider-equinix-metal/pkg/provider/apis"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

const (
	nameFmt       string = `[-a-z0-9]+`
	nameMaxLength int    = 63
	// SecretFieldAPIKey is the field name containing the API token
	SecretFieldAPIKey = "apiToken"
	// SecretFieldUserData is the field name containing the userData for the VM
	SecretFieldUserData = "userData"
)

var (
	nameRegexp          = regexp.MustCompile("^" + nameFmt + "$")
	secretFieldDefaults = []string{SecretFieldAPIKey, SecretFieldUserData}
)

// ValidateProviderSpec validates provider spec to check if all fields are present and valid
func ValidateProviderSpec(spec *api.EquinixMetalProviderSpec, fldPath *field.Path) field.ErrorList {
	// Code for validation of providerSpec goes here
	var (
		allErrs = field.ErrorList{}
	)

	if "" == spec.OS {
		allErrs = append(allErrs, field.Required(fldPath.Child("os"), "OS is required"))
	}
	if "" == spec.MachineType {
		allErrs = append(allErrs, field.Required(fldPath.Child("machineType"), "Machine Type is required"))
	}
	if "" == spec.ProjectID {
		allErrs = append(allErrs, field.Required(fldPath.Child("projectID"), "Project ID is required"))
	}
	if "" == spec.Metro {
		allErrs = append(allErrs, field.Required(fldPath.Child("metro"), "Metro is required"))
	}

	allErrs = append(allErrs, validateTags(spec.Tags, field.NewPath("spec.tags"))...)

	return allErrs
}

// ValidateName validate that a name is valid
func ValidateName(name string) []error {
	var (
		errs []error
	)
	if name == "" {
		errs = append(errs, errors.New("name must not be blank"))
	}
	if len(name) > nameMaxLength {
		errs = append(errs, fmt.Errorf("name was length %d, more than the maximum %d", len(name), nameMaxLength))
	}
	if !nameRegexp.MatchString(name) {
		errs = append(errs, fmt.Errorf("name did not match allowed regex '%v'", nameRegexp))
	}

	return errs
}

func validateTags(tags []string, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	clusterName := ""
	nodeRole := ""

	for _, key := range tags {
		if strings.Contains(key, "kubernetes.io/cluster/") {
			clusterName = key
		} else if strings.Contains(key, "kubernetes.io/role/") {
			nodeRole = key
		}
	}

	if clusterName == "" {
		allErrs = append(allErrs, field.Required(fldPath.Child("kubernetes.io/cluster/"), "Tag required of the form kubernetes.io/cluster/****"))
	}
	if nodeRole == "" {
		allErrs = append(allErrs, field.Required(fldPath.Child("kubernetes.io/role/"), "Tag required of the form kubernetes.io/role/****"))
	}

	return allErrs
}

// ValidateSecret makes sure that the supplied secrets contains the required fields
func ValidateSecret(secret *corev1.Secret, fields ...string) field.ErrorList {
	var (
		allErrs     = field.ErrorList{}
		fldPath     = field.NewPath("secretRef")
		matchFields = fields
	)
	if len(matchFields) == 0 {
		matchFields = secretFieldDefaults
	}

	if secret == nil {
		allErrs = append(allErrs, field.Required(fldPath.Child(""), "secretRef is required"))
	} else {
		for _, fieldName := range matchFields {
			if fieldName == SecretFieldAPIKey && "" == string(secret.Data[api.APIKey]) && "" == string(secret.Data[api.AlternateAPIKey]) {
				allErrs = append(allErrs, field.Required(fldPath.Child(api.APIKey), fmt.Sprintf("Required Equinix Metal API Key one of '%s' or '%s'", api.APIKey, api.AlternateAPIKey)))
			}
			if fieldName == SecretFieldUserData && "" == string(secret.Data["userData"]) {
				allErrs = append(allErrs, field.Required(fldPath.Child("userData"), "Required userData"))
			}
		}
	}

	return allErrs
}
