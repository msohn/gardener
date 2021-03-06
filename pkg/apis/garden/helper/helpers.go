// Copyright 2018 The Gardener Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package helper

import (
	"errors"

	"github.com/gardener/gardener/pkg/apis/garden"
)

// DetermineCloudProviderInProfile takes a CloudProfile specification and returns the cloud provider this profile is used for.
// If it is not able to determine it, an error will be returned.
func DetermineCloudProviderInProfile(spec garden.CloudProfileSpec) (garden.CloudProvider, error) {
	if spec.AWS != nil && spec.Azure == nil && spec.GCP == nil && spec.OpenStack == nil {
		return garden.CloudProviderAWS, nil
	}
	if spec.Azure != nil && spec.GCP == nil && spec.OpenStack == nil && spec.AWS == nil {
		return garden.CloudProviderAzure, nil
	}
	if spec.GCP != nil && spec.OpenStack == nil && spec.AWS == nil && spec.Azure == nil {
		return garden.CloudProviderGCP, nil
	}
	if spec.OpenStack != nil && spec.AWS == nil && spec.Azure == nil && spec.GCP == nil {
		return garden.CloudProviderOpenStack, nil
	}

	return "", errors.New("cloud profile must only contain exactly one field of aws/azure/gcp/openstack")
}
