package ecssd

import (
	"os"
	"time"

	"gopkg.in/yaml.v2"
)

const (
	DefaultRefreshInterval = 30 * time.Second
	DefaultJobLabelName    = "prometheus_job"
	AWSRegionEnvKey        = "AWS_REGION"
)

type Config struct {
	// ClusterName is the target ECS cluster name for service discovery.
	ClusterName string `mapstructure:"cluster_name" yaml:"cluster_name"`
	// ClusterRegion is the target ECS cluster's AWS region.
	ClusterRegion string `mapstructure:"cluster_region" yaml:"cluster_region"`
	// RefreshInterval determines how frequency at which the observer
	// needs to poll for collecting information about new processes.
	RefreshInterval time.Duration `mapstructure:"refresh_interval" yaml:"refresh_interval"`
	// ResultFile is the output path of the discovered targets YAML file (optional).
	// This is mainly used in conjunction with the Prometheus receiver.
	ResultFile string `mapstructure:"result_file" yaml:"result_file"`
	// JobLabelName is the override for prometheus job label, using `job` literal will cause error
	// in otel prometheus receiver. See https://github.com/open-telemetry/opentelemetry-collector/issues/575
	JobLabelName string `mapstructure:"job_label_name" yaml:"job_label_name"`
	// Services is a list of service name patterns for filtering tasks.
	Services []ServiceConfig `mapstructure:"services" yaml:"services"`
	// TaskDefinitions is a list of task definition arn patterns for filtering tasks.
	TaskDefinitions []TaskDefinitionConfig `mapstructure:"task_definitions" yaml:"task_definitions"`
	// DockerLabels is a list of docker labels for filtering containers within tasks.
	DockerLabels []DockerLabelConfig `mapstructure:"docker_labels" yaml:"docker_labels"`
}

func (c *Config) MatcherConfigs() map[MatcherType][]MatcherConfig {
	// We can have a registry or factory methods etc. but since we only have three type of metchers in filter.
	return map[MatcherType][]MatcherConfig{
		MatcherTypeService:        servicConfigsToMatchers(c.Services),
		MatcherTypeTaskDefinition: taskDefintionConfigsToMatchers(c.TaskDefinitions),
		MatcherTypeDockerLabel:    dockerLabelConfigToMatchers(c.DockerLabels),
	}
}

// LoadConfig use yaml.v2 to decode the struct.
// It returns the yaml decode error directly.
func LoadConfig(b []byte) (Config, error) {
	var c Config
	if err := yaml.Unmarshal(b, &c); err != nil {
		return Config{}, err
	}
	return c, nil
}

// DefaultConfig does NOT work out of box because it has no filters.
func DefaultConfig() Config {
	return Config{
		ClusterName:     "default",
		ClusterRegion:   os.Getenv(AWSRegionEnvKey),
		ResultFile:      "/etc/ecs_sd_targets.yaml",
		RefreshInterval: DefaultRefreshInterval,
		JobLabelName:    DefaultJobLabelName,
	}
}

// ExampleConfig returns an example instance that matches testdata/config_example.yaml.
// It can be used to validate if the struct tags like mapstructure, yaml are working properly.
func ExampleConfig() Config {
	return Config{
		ClusterName:     "ecs-sd-test-1",
		ClusterRegion:   "us-west-2",
		ResultFile:      "/etc/ecs_sd_targets.yaml",
		RefreshInterval: 15 * time.Second,
		JobLabelName:    DefaultJobLabelName,
		Services: []ServiceConfig{
			{
				NamePattern: "^retail-.*$",
			},
		},
		TaskDefinitions: []TaskDefinitionConfig{
			{
				CommonExporterConfig: CommonExporterConfig{
					JobName:      "task_def_1",
					MetricsPath:  "/not/metrics",
					MetricsPorts: []int{9113, 9090},
				},
				ArnPattern: ".*:task-definition/nginx:[0-9]+",
			},
		},
		DockerLabels: []DockerLabelConfig{
			{
				PortLabel: "ECS_PROMETHEUS_EXPORTER_PORT",
			},
		},
	}
}
