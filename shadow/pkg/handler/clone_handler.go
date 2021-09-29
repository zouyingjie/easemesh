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
	"log"

	"github.com/megaease/easemesh/mesh-shadow/pkg/object/v1beta1"
	appv1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (

	// Init container stuff.
	initContainerName = "initializer"

	agentVolumeName   = "agent-volume"
	sidecarVolumeName = "sidecar-volume"

	// Sidecar container stuff.
	sidecarContainerName = "easemesh-sidecar"

	shadowLabelKey            = "mesh.megaease.com/shadow-service"
	shadowAppContainerNameKey = "mesh.megaease.com/app-container-name"

	shadowDeploymentNameSuffix = "-shadow"

	databaseShadowConfigEnv      = "EASE_RESOURCE_DATABASE"
	kafkaShadowConfigEnv         = "EASE_RESOURCE_KAFKA"
	rabbitmqShadowConfigEnv      = "EASE_RESOURCE_RABBITMQ"
	redisShadowConfigEnv         = "EASE_RESOURCE_REDIS"
	elasticsearchShadowConfigEnv = "EASE_RESOURCE_ELASTICSEARCH"
)

type Cloner interface {
	Clone(obj interface{})
}

type ShadowServiceCloner struct {
	KubeClient    *kubernetes.Clientset
	RunTimeClient *client.Client
	CRDClient     *rest.RESTClient
}

func (cloner *ShadowServiceCloner) Clone(obj interface{}) {

	var err error
	block := obj.(ServiceCloneBlock)
	switch block.deployObj.(type) {
	case appv1.Deployment:
		deployment := block.deployObj.(appv1.Deployment)
		err = cloner.cloneDeployment(&deployment, &block.service)()
	case v1beta1.MeshDeployment:
		meshDeployment := block.deployObj.(v1beta1.MeshDeployment)
		err = cloner.cloneMeshDeployment(&meshDeployment, &block.service)()
	}
	if err != nil {
		log.Printf("Create shadow service failed. service: %s error: %s", block.service.ServiceName, err)
	}
}
