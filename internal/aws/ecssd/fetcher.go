// Copyright  OpenTelemetry Authors
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

package ecssd

import (
	"context"

	"github.com/hashicorp/golang-lru/simplelru"
	"go.uber.org/zap"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
)

type Fetcher interface {
	// FetcherAndDecorate fetches all the tasks and attach addation information
	// like definition, serivces and container instances.
	FetchAndDecorate(ctx context.Context) ([]*Task, error)
}

const (
	// ECS Service Quota: https://docs.aws.amazon.com/AmazonECS/latest/developerguide/service-quotas.html
	taskDefCacheSize               = 2000
	ec2CacheSize                   = 2000
	describeServiceLimit           = 10
	describeContainerInstanceLimit = 100
	// NOTE: these constants are not defined in go sdk, there are three values for deployment status.
	deploymentStatusActive   = "ACTIVE"
	deploymentStatusPrimary  = "PRIMARY"
	deploymentStatusInActive = "INACTIVE"
)

type TaskFetcher struct {
	logger            *zap.Logger
	ecs               *ecs.ECS
	ec2               *ec2.EC2
	cluster           string
	serviceNameFilter ServiceNameFilter
	taskDefCache      simplelru.LRUCache
	ec2Cache          simplelru.LRUCache
}

type TaskFetcherOptions struct {
	Logger            *zap.Logger
	Cluster           string
	Region            string
	ServiceNameFilter ServiceNameFilter
}

func NewTaskFetcher(opts TaskFetcherOptions) (*TaskFetcher, error) {
	panic("not implemented")
}

func (f *TaskFetcher) FetchAndDecorate(ctx context.Context) ([]*Task, error) {
	panic("not implemented")
}
