package ecssd

import (
	"regexp"

	"github.com/aws/aws-sdk-go/service/ecs"
	"go.uber.org/zap"
)

type ServiceConfig struct {
	CommonExporterConfig `mapstructure:",squash" yaml:",inline"`

	// NamePattern is mandetory, empty string means service name based match is skipped.
	NamePattern string `mapstructure:"name_pattern" yaml:"name_pattern"`
	// ContainerNamePattern is optional, empty string means all containers in that service would be exported.
	// Otherwise both service and container name petterns need to metch.
	ContainerNamePattern string `mapstructure:"container_name_pattern" yaml:"container_name_pattern"`

	nameRegex          *regexp.Regexp
	containerNameRegex *regexp.Regexp
}

func (s *ServiceConfig) Init() error {
	panic("not implemented")
}

func (s *ServiceConfig) NewMatcher(opts MatcherOptions) (Matcher, error) {
	panic("not implemented")
}

func servicConfigsToMatchers(cfgs []ServiceConfig) []MatcherConfig {
	if len(cfgs) == 0 {
		return nil
	}
	panic("not implemented")
}

type ServiceNameFilter func(name string) bool

func serviceConfigsToFilter(cfgs []ServiceConfig) (ServiceNameFilter, error) {
	// If no service config, don't descibe any services
	if len(cfgs) == 0 {
		return func(name string) bool {
			return false
		}, nil
	}
	panic("not implemented")
}

type ServiceMatcher struct {
	logger *zap.Logger
	cfg    ServiceConfig
}

func (s *ServiceMatcher) Type() MatcherType {
	return MatcherTypeService
}

func (s *ServiceMatcher) ExporterConfig() CommonExporterConfig {
	return s.cfg.CommonExporterConfig
}

func (s *ServiceMatcher) MatchTargets(t *Task, c *ecs.ContainerDefinition) ([]MatchedTarget, error) {
	panic("not implemented")
}
