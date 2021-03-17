package ecssd

import (
	"regexp"

	"github.com/aws/aws-sdk-go/service/ecs"
	"go.uber.org/zap"
)

type TaskDefinitionConfig struct {
	CommonExporterConfig `mapstructure:",squash" yaml:",inline"`

	// ArnPattern is mandetory, empty string means arn based match is skipped.
	ArnPattern string `mapstructure:"arn_pattern" yaml:"arn_pattern"`
	// ContainerNamePattern is optional, empty string means all containers in that task definition would be exported.
	// Otherwise both service and container name petterns need to metch.
	ContainerNamePattern string `mapstructure:"container_name_pattern" yaml:"container_name_pattern"`

	arnRegex           *regexp.Regexp
	containerNameRegex *regexp.Regexp
}

func (t *TaskDefinitionConfig) Init() error {
	panic("not implemented")
}

func (t *TaskDefinitionConfig) NewMatcher(opts MatcherOptions) (Matcher, error) {
	panic("not implemented")
}

func taskDefintionConfigsToMatchers(cfgs []TaskDefinitionConfig) []MatcherConfig {
	if len(cfgs) == 0 {
		return nil
	}
	panic("not implemented")
}

type TaskDefinitionMatcher struct {
	logger *zap.Logger
	cfg    TaskDefinitionConfig
}

func (m *TaskDefinitionMatcher) Type() MatcherType {
	return MatcherTypeTaskDefinition
}

func (m *TaskDefinitionMatcher) ExporterConfig() CommonExporterConfig {
	return m.cfg.CommonExporterConfig
}

func (m *TaskDefinitionMatcher) MatchTargets(t *Task, c *ecs.ContainerDefinition) ([]MatchedTarget, error) {
	panic("not implemented")
}
