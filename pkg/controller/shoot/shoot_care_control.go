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

package shoot

import (
	"fmt"
	"sync"

	"github.com/gardener/gardener/pkg/apis/componentconfig"
	gardenv1beta1 "github.com/gardener/gardener/pkg/apis/garden/v1beta1"
	"github.com/gardener/gardener/pkg/apis/garden/v1beta1/helper"
	gardeninformers "github.com/gardener/gardener/pkg/client/garden/informers/externalversions/garden/v1beta1"
	"github.com/gardener/gardener/pkg/client/kubernetes"
	"github.com/gardener/gardener/pkg/logger"
	"github.com/gardener/gardener/pkg/operation"
	botanistpkg "github.com/gardener/gardener/pkg/operation/botanist"
	"github.com/gardener/gardener/pkg/operation/cloudbotanist"
	corev1 "k8s.io/api/core/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/cache"
)

func (c *Controller) shootCareAdd(obj interface{}) {
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		return
	}
	c.
		shootCareQueue.
		AddAfter(key, c.config.Controller.HealthCheckPeriod.Duration)
}

func (c *Controller) shootCareDelete(obj interface{}) {
	shoot, ok := obj.(*gardenv1beta1.Shoot)
	if shoot == nil || !ok {
		return
	}
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		return
	}
	c.
		shootCareQueue.
		Done(key)
}

func (c *Controller) reconcileShootCareKey(key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}
	shoot, err := c.shootLister.Shoots(namespace).Get(name)
	if apierrors.IsNotFound(err) {
		logger.Logger.Debugf("[SHOOT CARE] %s - skipping because Shoot has been deleted", key)
		return nil
	}
	if err != nil {
		logger.Logger.Infof("[SHOOT CARE] %s - unable to retrieve object from store: %v", key, err)
		return err
	}
	defer c.shootCareAdd(shoot)
	if operationOngoing(shoot) {
		logger.Logger.Debugf("[SHOOT CARE] %s - skipping because an operation in ongoing", key)
		return nil
	}
	return c.careControl.Care(shoot, key)
}

// CareControlInterface implements the control logic for caring for Shoots. It is implemented as an interface to allow
// for extensions that provide different semantics. Currently, there is only one implementation.
type CareControlInterface interface {
	Care(shoot *gardenv1beta1.Shoot, key string) error
}

// NewDefaultCareControl returns a new instance of the default implementation CareControlInterface that
// implements the documented semantics for caring for Shoots. updater is the UpdaterInterface used
// to update the status of Shoots. You should use an instance returned from NewDefaultCareControl() for any
// scenario other than testing.
func NewDefaultCareControl(k8sGardenClient kubernetes.Client, k8sGardenInformers gardeninformers.Interface, secrets map[string]*corev1.Secret, identity *gardenv1beta1.Gardener, config *componentconfig.ControllerManagerConfiguration, updater UpdaterInterface) CareControlInterface {
	return &defaultCareControl{k8sGardenClient, k8sGardenInformers, secrets, identity, config, updater}
}

type defaultCareControl struct {
	k8sGardenClient    kubernetes.Client
	k8sGardenInformers gardeninformers.Interface
	secrets            map[string]*corev1.Secret
	identity           *gardenv1beta1.Gardener
	config             *componentconfig.ControllerManagerConfiguration
	updater            UpdaterInterface
}

func (c *defaultCareControl) Care(shootObj *gardenv1beta1.Shoot, key string) error {
	var (
		shoot            = shootObj.DeepCopy()
		shootLogger      = logger.NewShootLogger(logger.Logger, shoot.Name, shoot.Namespace, "")
		healthCheckError = "ShootCareError"
		updateStatus     = func(conditionControlPlaneHealthy, conditionEveryNodeReady, conditionSystemComponentsHealthy *gardenv1beta1.ShootCondition) {
			if err := c.updateShootStatus(shoot, []gardenv1beta1.ShootCondition{*conditionControlPlaneHealthy, *conditionEveryNodeReady, *conditionSystemComponentsHealthy}); err != nil {
				shootLogger.Errorf("Could not update the Shoot status in care controller: %+v", err)
			}
		}
	)
	shootLogger.Debugf("[SHOOT CARE] %s", key)

	operation, err := operation.New(shoot, shootLogger, c.k8sGardenClient, c.k8sGardenInformers, c.identity, c.secrets)
	if err != nil {
		shootLogger.Errorf("could not initialize a new operation: %s", err.Error())
		return nil
	}

	// Initialize conditions based on the current status.
	conditionControlPlaneHealthy, conditionEveryNodeReady, conditionSystemComponentsHealthy := initConditions(shoot.Status.Conditions)

	botanist, err := botanistpkg.New(operation)
	if err != nil {
		message := fmt.Sprintf("Failed to create a botanist object to perform the care operations (%s).", err.Error())
		conditionControlPlaneHealthy = helper.ModifyCondition(conditionControlPlaneHealthy, corev1.ConditionUnknown, healthCheckError, message)
		conditionEveryNodeReady = helper.ModifyCondition(conditionEveryNodeReady, corev1.ConditionUnknown, healthCheckError, message)
		conditionSystemComponentsHealthy = helper.ModifyCondition(conditionSystemComponentsHealthy, corev1.ConditionUnknown, healthCheckError, message)
		operation.Logger.Error(message)
		updateStatus(conditionControlPlaneHealthy, conditionEveryNodeReady, conditionSystemComponentsHealthy)
		return nil
	}
	cloudBotanist, err := cloudbotanist.New(operation)
	if err != nil {
		message := fmt.Sprintf("Failed to create a Cloud Botanist to perform the care operations (%s).", err.Error())
		conditionControlPlaneHealthy = helper.ModifyCondition(conditionControlPlaneHealthy, corev1.ConditionUnknown, healthCheckError, message)
		conditionEveryNodeReady = helper.ModifyCondition(conditionEveryNodeReady, corev1.ConditionUnknown, healthCheckError, message)
		conditionSystemComponentsHealthy = helper.ModifyCondition(conditionSystemComponentsHealthy, corev1.ConditionUnknown, healthCheckError, message)
		operation.Logger.Error(message)
		updateStatus(conditionControlPlaneHealthy, conditionEveryNodeReady, conditionSystemComponentsHealthy)
		return nil
	}
	err = botanist.InitializeShootClients()
	if err != nil {
		message := fmt.Sprintf("Failed to create a K8SClient for the Shoot cluster to perform the care operations (%s).", err.Error())
		conditionEveryNodeReady = helper.ModifyCondition(conditionEveryNodeReady, corev1.ConditionUnknown, healthCheckError, message)
		conditionSystemComponentsHealthy = helper.ModifyCondition(conditionSystemComponentsHealthy, corev1.ConditionUnknown, healthCheckError, message)
		operation.Logger.Error(message)
		updateStatus(conditionControlPlaneHealthy, conditionEveryNodeReady, conditionSystemComponentsHealthy)
		return nil
	}

	// Trigger garbage collection
	garbageCollection(botanist)

	// Trigger health check
	conditionControlPlaneHealthy, conditionEveryNodeReady, conditionSystemComponentsHealthy = healthCheck(botanist, cloudBotanist, conditionControlPlaneHealthy, conditionEveryNodeReady, conditionSystemComponentsHealthy)

	// Update Shoot status
	updateStatus(conditionControlPlaneHealthy, conditionEveryNodeReady, conditionSystemComponentsHealthy)

	return nil
}

func (c *defaultCareControl) updateShootStatus(shoot *gardenv1beta1.Shoot, conditions []gardenv1beta1.ShootCondition) error {
	if shoot.Status.Conditions != nil && !apiequality.Semantic.DeepEqual(conditions, shoot.Status.Conditions) {
		return nil
	}
	shoot.Status.Conditions = conditions

	_, err := c.updater.UpdateShootStatusIfNoOperation(shoot)
	return err
}

// garbageCollection cleans the Seed and the Shoot cluster from unrequired objects.
// It receives a Garden object <garden> which stores the Shoot object.
func garbageCollection(botanist *botanistpkg.Botanist) {
	var wg sync.WaitGroup

	wg.Add(2)
	go func() {
		defer wg.Done()
		botanist.PerformGarbageCollectionSeed()
	}()
	go func() {
		defer wg.Done()
		botanist.PerformGarbageCollectionShoot()
	}()
	wg.Wait()

	botanist.Logger.Debugf("Successfully performed garbage collection for Shoot cluster '%s'", botanist.Shoot.Info.Name)
}

// healthCheck performs several health checks and updates the status conditions.
// It receives a Garden object <garden> which stores the Shoot object.
// The current Health check verifies that the control plane running in the Seed cluster is healthy, every
// node is ready and that all system components (pods running kube-system) are healthy.
func healthCheck(botanist *botanistpkg.Botanist, cloudBotanist cloudbotanist.CloudBotanist, conditionControlPlaneHealthy, conditionEveryNodeReady, conditionSystemComponentsHealthy *gardenv1beta1.ShootCondition) (*gardenv1beta1.ShootCondition, *gardenv1beta1.ShootCondition, *gardenv1beta1.ShootCondition) {
	var (
		currentlyScaling = false
		healthyInstances = 0
		wg               sync.WaitGroup
	)

	// We ask the Cloud Botanist whether the Shoot cluster is currently scaled or not to avoid problems of potentially
	// non-existing infrastructure resources like autoscaling groups.
	currentlyScaling, healthyInstances, _ = cloudBotanist.CheckIfClusterGetsScaled()

	wg.Add(3)
	go func() {
		defer wg.Done()
		conditionControlPlaneHealthy = botanist.CheckConditionControlPlaneHealthy(conditionControlPlaneHealthy)
	}()
	go func() {
		defer wg.Done()
		conditionEveryNodeReady = botanist.CheckConditionEveryNodeReady(conditionEveryNodeReady, currentlyScaling, healthyInstances)
	}()
	go func() {
		defer wg.Done()
		conditionSystemComponentsHealthy = botanist.CheckConditionSystemComponentsHealthy(conditionSystemComponentsHealthy)
	}()
	wg.Wait()

	botanist.Logger.Debugf("Successfully performed health check for Shoot cluster '%s'", botanist.Shoot.Info.Name)
	return conditionControlPlaneHealthy, conditionEveryNodeReady, conditionSystemComponentsHealthy
}

// initConditions initializes the Shoot conditions based on an existing list. If a condition type does not exist
// in the list yet, it will be set to default values.
func initConditions(conditions []gardenv1beta1.ShootCondition) (*gardenv1beta1.ShootCondition, *gardenv1beta1.ShootCondition, *gardenv1beta1.ShootCondition) {
	var (
		conditionControlPlaneHealthy     *gardenv1beta1.ShootCondition
		conditionEveryNodeReady          *gardenv1beta1.ShootCondition
		conditionSystemComponentsHealthy *gardenv1beta1.ShootCondition
	)

	// We retrieve the current conditions in order to update them appropriately.
	for _, condition := range conditions {
		if condition.Type == gardenv1beta1.ShootControlPlaneHealthy {
			c := condition
			conditionControlPlaneHealthy = &c
		}
		if condition.Type == gardenv1beta1.ShootEveryNodeReady {
			c := condition
			conditionEveryNodeReady = &c
		}
		if condition.Type == gardenv1beta1.ShootSystemComponentsHealthy {
			c := condition
			conditionSystemComponentsHealthy = &c
		}
	}

	// If the conditions have not been set yet for a cluster, we have to initialize them once.
	if conditionControlPlaneHealthy == nil {
		conditionControlPlaneHealthy = helper.InitCondition(gardenv1beta1.ShootControlPlaneHealthy, "", "")
	}
	if conditionEveryNodeReady == nil {
		conditionEveryNodeReady = helper.InitCondition(gardenv1beta1.ShootEveryNodeReady, "", "")
	}
	if conditionSystemComponentsHealthy == nil {
		conditionSystemComponentsHealthy = helper.InitCondition(gardenv1beta1.ShootSystemComponentsHealthy, "", "")
	}

	return conditionControlPlaneHealthy, conditionEveryNodeReady, conditionSystemComponentsHealthy
}
