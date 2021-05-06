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

package ecsobserver

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

const (
	defaultMetricsPath = "/metrics"
)

// CommonExporterConfig should be embedded into filter config.
// They set labels like job, metrics_path etc. that can override prometheus default.
type CommonExporterConfig struct {
	JobName      string `mapstructure:"job_name" yaml:"job_name"`
	MetricsPath  string `mapstructure:"metrics_path" yaml:"metrics_path"`
	MetricsPorts []int  `mapstructure:"metrics_ports" yaml:"metrics_ports"`
}

// newExportSetting checks if there are duplicated metrics ports.
func (c *CommonExporterConfig) newExportSetting() (*commonExportSetting, error) {
	m := make(map[int]bool)
	for _, p := range c.MetricsPorts {
		if m[p] {
			return nil, fmt.Errorf("metrics_ports has duplicated port %d", p)
		}
		m[p] = true
	}
	return &commonExportSetting{CommonExporterConfig: *c, metricsPorts: m}, nil
}

// commonExportSetting is generated from CommonExportConfig with some util methods.
type commonExportSetting struct {
	CommonExporterConfig
	metricsPorts map[int]bool
}

func (s *commonExportSetting) hasContainerPort(containerPort int) bool {
	return s.metricsPorts[containerPort]
}

// taskExporter converts annotated Task into PrometheusECSTarget.
type taskExporter struct {
	logger  *zap.Logger
	cluster string
}

func newTaskExporter(logger *zap.Logger, cluster string) *taskExporter {
	return &taskExporter{
		logger:  logger,
		cluster: cluster,
	}
}

// ExportTasks loops a list of tasks and export prometheus scrape targets.
// It keeps track of error but does NOT stop when error occurs.
// The returned targets are all valid and the error(s) are mainly for generating metrics.
func (e *taskExporter) ExportTasks(tasks []*Task) ([]PrometheusECSTarget, error) {
	var merr error
	var allTargets []PrometheusECSTarget
	for _, t := range tasks {
		targets, err := e.ExportTask(t)
		multierr.AppendInto(&merr, err) // if err == nil, AppendInto does nothing
		// Even if there are error, returned targets are still valid.
		allTargets = append(allTargets, targets...)
	}
	return allTargets, merr
}

// ExportTask exports all the matched container within a single task.
// One task can contain multiple containers. One container can have more than one target
// if there are multiple ports in `metrics_port`.
func (e *taskExporter) ExportTask(task *Task) ([]PrometheusECSTarget, error) {
	// All targets in one task shares same IP.
	privateIP, err := task.PrivateIP()
	if err != nil {
		return nil, err
	}

	// Base for all the containers in this task, most attributes are same.
	baseTarget := PrometheusECSTarget{
		Source:                 aws.StringValue(task.Task.TaskArn),
		MetricsPath:            defaultMetricsPath,
		ClusterName:            e.cluster,
		TaskDefinitionFamily:   aws.StringValue(task.Definition.Family),
		TaskDefinitionRevision: int(aws.Int64Value(task.Definition.Revision)),
		TaskStartedBy:          aws.StringValue(task.Task.StartedBy),
		TaskLaunchType:         aws.StringValue(task.Task.LaunchType),
		TaskGroup:              aws.StringValue(task.Task.Group),
		TaskTags:               task.TaskTags(),
		HealthStatus:           aws.StringValue(task.Task.HealthStatus),
	}
	if task.Service != nil {
		baseTarget.ServiceName = aws.StringValue(task.Service.ServiceName)
	}
	if task.EC2 != nil {
		ec2 := task.EC2
		baseTarget.EC2InstanceID = aws.StringValue(ec2.InstanceId)
		baseTarget.EC2InstanceType = aws.StringValue(ec2.InstanceType)
		baseTarget.EC2Tags = task.EC2Tags()
		baseTarget.EC2VpcID = aws.StringValue(ec2.VpcId)
		baseTarget.EC2SubnetID = aws.StringValue(ec2.SubnetId)
		baseTarget.EC2PrivateIP = privateIP
		baseTarget.EC2PublicIP = aws.StringValue(ec2.PublicIpAddress)
	}

	var targetsInTask []PrometheusECSTarget
	var merr error
	for _, m := range task.Matched {
		container := task.Definition.ContainerDefinitions[m.ContainerIndex]
		// Shallow copy task level attributes
		containerTarget := baseTarget
		// Add container specific info
		containerTarget.ContainerName = aws.StringValue(container.Name)
		containerTarget.ContainerLabels = task.ContainerLabels(m.ContainerIndex)
		// Multiple targets for a single container
		for _, matchedTarget := range m.Targets {
			// Shallow copy from container
			target := containerTarget
			mappedPort, err := task.MappedPort(container, int64(matchedTarget.Port))
			// Skip this target and keep track of port error, does not abort.
			if multierr.AppendInto(&merr, err) {
				continue
			}
			target.Address = fmt.Sprintf("%s:%d", privateIP, mappedPort)
			if matchedTarget.MetricsPath != "" {
				target.MetricsPath = matchedTarget.MetricsPath
			}
			target.Job = matchedTarget.Job
			targetsInTask = append(targetsInTask, target)
		}
	}
	return targetsInTask, merr
}
