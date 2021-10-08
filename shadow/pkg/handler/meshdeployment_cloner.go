/*
 * Copyright (c) 2017, MegaEase
 * All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package handler

import (
	"github.com/megaease/easemesh/mesh-operator/pkg/api/v1beta1"
	"github.com/megaease/easemesh/mesh-shadow/pkg/object"
	"github.com/megaease/easemesh/mesh-shadow/pkg/utils"
	"github.com/pkg/errors"
	appsV1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CloneMeshDeploymentFunc func() error

type cloneMeshDeploymentSpecFunc func(ourceMeshDeployment *v1beta1.MeshDeployment, shadowService *object.ShadowService) *v1beta1.MeshDeployment

func (cloner *ShadowServiceCloner) cloneMeshDeployment(sourceMeshDeployment *v1beta1.MeshDeployment, shadowService *object.ShadowService) CloneMeshDeploymentFunc {
	shadowMeshDeployment := cloner.shadowMeshDeploymentBaseSpec(cloner.shadowMeshDeploymentInitialize(nil))(sourceMeshDeployment, shadowService)
	return func() error {
		err := utils.DeployMesheployment(shadowMeshDeployment.Namespace, shadowMeshDeployment, cloner.CRDClient)
		if err != nil {
			return errors.Wrapf(err, "Clone mesh deployment %s for service %s failed", sourceMeshDeployment.Name, shadowService.ServiceName)
		}
		return err
	}
}

func (cloner *ShadowServiceCloner) shadowMeshDeploymentInitialize(fn cloneMeshDeploymentSpecFunc) cloneMeshDeploymentSpecFunc {
	return func(sourceMeshDeployment *v1beta1.MeshDeployment, shadowService *object.ShadowService) *v1beta1.MeshDeployment {
		return &v1beta1.MeshDeployment{
			TypeMeta: sourceMeshDeployment.TypeMeta,
			ObjectMeta: metav1.ObjectMeta{
				Name:        shadowName(sourceMeshDeployment.Name),
				Namespace:   sourceMeshDeployment.Namespace,
				Labels:      sourceMeshDeployment.Labels,
				Annotations: sourceMeshDeployment.Annotations,
			},
		}
	}
}

func (cloner *ShadowServiceCloner) shadowMeshDeploymentBaseSpec(fn cloneMeshDeploymentSpecFunc) cloneMeshDeploymentSpecFunc {
	return func(sourceMeshDeployment *v1beta1.MeshDeployment, shadowService *object.ShadowService) *v1beta1.MeshDeployment {
		meshDeployment := fn(sourceMeshDeployment, shadowService)
		meshDeployment.Spec.Service = sourceMeshDeployment.Spec.Service

		labels := meshDeployment.Labels
		if labels == nil {
			labels = make(map[string]string)
		}

		shadowServiceLabels := shadowServiceLabels()
		for k, v := range shadowServiceLabels {
			labels[k] = v
		}
		meshDeployment.Labels = labels

		deployment := &appsV1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      sourceMeshDeployment.Name,
				Namespace: sourceMeshDeployment.Namespace,
			},
			Spec: sourceMeshDeployment.Spec.Deploy.DeploymentSpec,
		}
		if sourceMeshDeployment.Spec.Service.AppContainerName != "" {
			deployment.Annotations[shadowAppContainerNameKey] = sourceMeshDeployment.Spec.Service.AppContainerName
		}
		shadowDeployment := cloner.cloneDeploymentSpec(deployment, shadowService)
		meshDeployment.Spec.Deploy.DeploymentSpec = shadowDeployment.Spec

		return meshDeployment
	}
}
