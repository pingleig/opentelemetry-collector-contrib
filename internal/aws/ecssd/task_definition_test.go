package ecssd

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/stretchr/testify/assert"
)

func TestTaskDefinitionMatcher_Match(t *testing.T) {
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
									ContainerPort: aws.Int64(2112),
								},
								{
									ContainerPort: aws.Int64(2021),
								},
							},
						},
					},
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
							PortMappings: []*ecs.PortMapping{
								{
									ContainerPort: aws.Int64(2112),
								},
								{
									ContainerPort: aws.Int64(2021),
								},
							},
						},
					},
				},
			},
		}
	}
	t.Run("skip empty config", func(t *testing.T) {
		res := newMatcherAndMatch(t, &TaskDefinitionConfig{}, genTasks())
		assert.Len(t, res.Tasks, 0)
	})

	t.Run("skip container name only", func(t *testing.T) {
		cfg := TaskDefinitionConfig{
			ContainerNamePattern: "foo",
		}
		res := newMatcherAndMatch(t, &cfg, genTasks())
		assert.Len(t, res.Tasks, 0)
	})

	t.Run("task arn", func(t *testing.T) {
		cfg := TaskDefinitionConfig{
			ArnPattern: "^t1$",
		}
		res := newMatcherAndMatch(t, &cfg, genTasks())
		assert.Equal(t, &MatchResult{
			Tasks: []int{0},
			Containers: []MatchedContainer{
				{
					TaskIndex:      0,
					ContainerIndex: 0,
				},
				{
					TaskIndex:      0,
					ContainerIndex: 1,
				},
			},
		}, res)
	})

	t.Run("container name", func(t *testing.T) {
		cfg := TaskDefinitionConfig{
			ArnPattern:           "^t.*$",
			ContainerNamePattern: "^c2-t[0-9]$",
			CommonExporterConfig: CommonExporterConfig{
				MetricsPorts: []int{2112, 2021},
			},
		}
		res := newMatcherAndMatch(t, &cfg, genTasks())
		assert.Equal(t, &MatchResult{
			Tasks: []int{0, 1},
			Containers: []MatchedContainer{
				{
					TaskIndex:      0,
					ContainerIndex: 1,
					Targets: []MatchedTarget{
						{
							MatcherType: MatcherTypeTaskDefinition,
							Port:        2112,
						},
						{
							MatcherType: MatcherTypeTaskDefinition,
							Port:        2021,
						},
					},
				},
				{
					TaskIndex:      1,
					ContainerIndex: 1,
					Targets: []MatchedTarget{
						{
							MatcherType: MatcherTypeTaskDefinition,
							Port:        2112,
						},
						{
							MatcherType: MatcherTypeTaskDefinition,
							Port:        2021,
						},
					},
				},
			},
		}, res)
	})
}
