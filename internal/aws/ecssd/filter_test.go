package ecssd

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestTaskFilter_Filter(t *testing.T) {
	// We need this func because filter will update tasks in place to attach matcher information.
	// This is poor man's deep copy.
	genTasks := func() []*Task {
		return []*Task{
			{
				Task: &ecs.Task{
					TaskDefinitionArn: aws.String("t1"),
				},
				Definition: &ecs.TaskDefinition{
					ContainerDefinitions: []*ecs.ContainerDefinition{
						{
							Name: aws.String("c1-t1"),
						},
						{
							Name: aws.String("c2-t1"),
							PortMappings: []*ecs.PortMapping{
								{
									ContainerPort: aws.Int64(1234),
								},
							},
						},
					},
				},
				Service: &ecs.Service{
					ServiceName: aws.String("s1"),
				},
			},
			{
				Task: &ecs.Task{
					TaskDefinitionArn: aws.String("t2"),
				},
				Definition: &ecs.TaskDefinition{
					ContainerDefinitions: []*ecs.ContainerDefinition{
						{
							Name: aws.String("c1-t2"),
							DockerLabels: map[string]*string{
								"NOT_PORT": aws.String("just a value"),
							},
							PortMappings: []*ecs.PortMapping{
								{
									ContainerPort: aws.Int64(5678),
								},
							},
						},
						{
							Name: aws.String("c2-t2"),
							DockerLabels: map[string]*string{
								"PROMETHEUS_PORT": aws.String("2112"),
							},
						},
					},
				},
				Service: &ecs.Service{
					ServiceName: aws.String("s2"),
				},
			},
		}
	}

	t.Run("service only", func(t *testing.T) {
		t.Run("single service", func(t *testing.T) {
			c := Config{
				Services: []ServiceConfig{
					{
						NamePattern: "s1",
						CommonExporterConfig: CommonExporterConfig{
							MetricsPorts: []int{1234},
						},
					},
				},
			}
			tasks := genTasks()
			filtered := filterTasks(t, c, tasks)
			assert.Len(t, filtered, 1)
			assert.Equal(t, []MatchedContainer{
				{
					TaskIndex:      0,
					ContainerIndex: 0,
					// no targets because no port
				},
				{
					TaskIndex:      0,
					ContainerIndex: 1,
					Targets: []MatchedTarget{
						{
							MatcherType: MatcherTypeService,
							Port:        1234,
						},
					},
				},
			}, filtered[0].Matched)
			assert.Equal(t, tasks[0], filtered[0], "update inplace")
		})

		t.Run("multiple service", func(t *testing.T) {
			// NOTE: found a https://www.evanjones.ca/go-gotcha-loop-variables.html bug using this test ...
			c := Config{
				Services: []ServiceConfig{
					{
						NamePattern: "s1",
						CommonExporterConfig: CommonExporterConfig{
							MetricsPorts: []int{1234},
						},
					},
					{
						NamePattern: "s2",
						CommonExporterConfig: CommonExporterConfig{
							MetricsPorts: []int{5678},
						},
					},
				},
			}
			tasks := genTasks()
			require.Len(t, tasks, 2)
			filtered := filterTasks(t, c, tasks)
			assert.Len(t, filtered, 2)
			assert.Len(t, filtered[0].Matched, 2)
			assert.Equal(t, MatchedContainer{
				TaskIndex:      1,
				ContainerIndex: 0,
				Targets: []MatchedTarget{
					{
						MatcherType:  MatcherTypeService,
						MatcherIndex: 1,
						Port:         5678,
					},
				},
			}, filtered[1].Matched[0])
		})
	})

	t.Run("service and task definition", func(t *testing.T) {
		c := Config{
			Services: []ServiceConfig{
				{
					NamePattern: "s1",
				},
			},
			TaskDefinitions: []TaskDefinitionConfig{
				{
					ArnPattern: "t2",
				},
			},
		}
		tasks := genTasks()
		filtered := filterTasks(t, c, tasks)
		assert.Len(t, filtered, 2)
		assert.Equal(t, MatchedContainer{
			TaskIndex:      1,
			ContainerIndex: 1,
		}, filtered[1].Matched[1])
	})

	t.Run("service and task definition and docker label", func(t *testing.T) {
		c := Config{
			Services: []ServiceConfig{
				{
					NamePattern: "s1",
					CommonExporterConfig: CommonExporterConfig{
						MetricsPorts: []int{1234},
					},
				},
			},
			TaskDefinitions: []TaskDefinitionConfig{
				{
					ArnPattern: "t1",
					CommonExporterConfig: CommonExporterConfig{
						MetricsPorts: []int{1234},
					},
				},
			},
			DockerLabels: []DockerLabelConfig{
				{
					PortLabel: "PROMETHEUS_PORT",
				},
			},
		}
		tasks := genTasks()
		filtered := filterTasks(t, c, tasks)
		assert.Len(t, filtered, 2)
		// both service and task match t1, but match by service has higher priority.
		assert.Equal(t, MatchedContainer{
			TaskIndex:      0,
			ContainerIndex: 1,
			Targets: []MatchedTarget{
				{
					MatcherType: MatcherTypeService,
					Port:        1234,
				},
			},
		}, filtered[0].Matched[1])
		assert.Equal(t, MatchedContainer{
			TaskIndex:      1,
			ContainerIndex: 1,
			Targets: []MatchedTarget{
				{
					MatcherType: MatcherTypeDockerLabel,
					Port:        2112,
				},
			},
		}, filtered[1].Matched[0])
	})
}

// Util Start

func filterTasks(t *testing.T, c Config, tasks []*Task) []*Task {
	matchers, err := newMatchers(c, MatcherOptions{
		Logger: zap.NewExample(),
	})
	require.NoError(t, err)
	filter, err := NewTaskFilter(c, TaskFilterOptions{
		Logger:   zap.NewExample(),
		Matchers: matchers,
	})
	require.NoError(t, err)
	filtered, err := filter.Filter(tasks)
	require.NoError(t, err)
	return filtered
}

func newMatcher(t *testing.T, cfg MatcherConfig) Matcher {
	require.NoError(t, cfg.Init())
	m, err := cfg.NewMatcher(testMatcherOptions())
	require.NoError(t, err)
	return m
}

func newMatcherAndMatch(t *testing.T, cfg MatcherConfig, tasks []*Task) *MatchResult {
	m := newMatcher(t, cfg)
	res, err := matchContainers(tasks, m, 0)
	require.NoError(t, err)
	return res
}

func testMatcherOptions() MatcherOptions {
	return MatcherOptions{
		Logger: zap.NewExample(),
	}
}

// Util End
