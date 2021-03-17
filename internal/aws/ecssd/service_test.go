package ecssd

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServiceNameFilter(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		f, err := serviceConfigsToFilter(nil)
		require.NoError(t, err)
		assert.False(t, f("darcy"))
	})

	t.Run("single config", func(t *testing.T) {
		cfgs := []ServiceConfig{
			{
				NamePattern: "^retail.*$",
			},
		}
		f, err := serviceConfigsToFilter(cfgs)
		require.NoError(t, err)
		assert.True(t, f("retail-bar"))
		assert.False(t, f("retai-bar"))
	})

	t.Run("multi config", func(t *testing.T) {
		cfgs := []ServiceConfig{
			{
				NamePattern: "^retail.*$",
			},
			{
				NamePattern: "darcy",
			},
		}
		f, err := serviceConfigsToFilter(cfgs)
		require.NoError(t, err)
		assert.True(t, f("retail-darcy"))
		assert.False(t, f("just don't match"))
	})
}

func TestServiceMatcher_Match(t *testing.T) {
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
							PortMappings: []*ecs.PortMapping{
								{
									ContainerPort: aws.Int64(2021),
								},
							},
						},
						{
							Name: aws.String("c2-t1"),
							PortMappings: []*ecs.PortMapping{
								{
									ContainerPort: aws.Int64(2022),
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
						},
						{
							Name: aws.String("c2-t2"),
						},
					},
				},
				Service: &ecs.Service{
					ServiceName: aws.String("s2"),
				},
			},
		}
	}

	t.Run("skip empty config", func(t *testing.T) {
		res := newMatcherAndMatch(t, &ServiceConfig{}, nil)
		assert.Len(t, res.Tasks, 0)
	})

	t.Run("skip contaienr name only", func(t *testing.T) {
		cfg := ServiceConfig{
			ContainerNamePattern: "foo",
		}
		res := newMatcherAndMatch(t, &cfg, genTasks())
		assert.Len(t, res.Tasks, 0)
	})

	t.Run("service name", func(t *testing.T) {
		cfg := ServiceConfig{
			NamePattern: "s1",
			CommonExporterConfig: CommonExporterConfig{
				MetricsPorts: []int{2021, 2022},
			},
		}
		res := newMatcherAndMatch(t, &cfg, genTasks())
		assert.Equal(t, &MatchResult{
			Tasks: []int{0},
			Containers: []MatchedContainer{
				{
					TaskIndex:      0,
					ContainerIndex: 0,
					Targets: []MatchedTarget{
						{
							MatcherType: MatcherTypeService,
							Port:        2021,
						},
					},
				},
				{
					TaskIndex:      0,
					ContainerIndex: 1,
					Targets: []MatchedTarget{
						{
							MatcherType: MatcherTypeService,
							Port:        2022,
						},
					},
				},
			},
		}, res)
	})

	t.Run("container name", func(t *testing.T) {
		cfg := ServiceConfig{
			NamePattern:          "s1",
			ContainerNamePattern: "c2",
			CommonExporterConfig: CommonExporterConfig{
				MetricsPorts: []int{2022},
			},
		}
		res := newMatcherAndMatch(t, &cfg, genTasks())
		assert.Equal(t, &MatchResult{
			Tasks: []int{0},
			Containers: []MatchedContainer{
				{
					TaskIndex:      0,
					ContainerIndex: 1,
					Targets: []MatchedTarget{
						{
							MatcherType: MatcherTypeService,
							Port:        2022,
						},
					},
				},
			},
		}, res)
	})
}
