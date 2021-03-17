package ecssd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTask_AddMatchedContainer(t *testing.T) {
	task := Task{
		Matched: []MatchedContainer{
			{
				ContainerIndex: 0,
				Targets: []MatchedTarget{
					{
						MatcherType: MatcherTypeService,
						Port:        1,
					},
				},
			},
		},
	}

	// Different container
	task.AddMatchedContainer(MatchedContainer{
		ContainerIndex: 1,
		Targets: []MatchedTarget{
			{
				MatcherType: MatcherTypeDockerLabel,
				Port:        2,
			},
		},
	})
	assert.Equal(t, []MatchedContainer{
		{
			ContainerIndex: 0,
			Targets: []MatchedTarget{
				{
					MatcherType: MatcherTypeService,
					Port:        1,
				},
			},
		},
		{
			ContainerIndex: 1,
			Targets: []MatchedTarget{
				{
					MatcherType: MatcherTypeDockerLabel,
					Port:        2,
				},
			},
		},
	}, task.Matched)

	// Same container different metrics path
	task.AddMatchedContainer(MatchedContainer{
		ContainerIndex: 0,
		Targets: []MatchedTarget{
			{
				MatcherType: MatcherTypeTaskDefinition,
				Port:        1,
				MetricsPath: "/metrics2",
			},
		},
	})
	assert.Equal(t, []MatchedContainer{
		{
			ContainerIndex: 0,
			Targets: []MatchedTarget{
				{
					MatcherType: MatcherTypeService,
					Port:        1,
				},
				{
					MatcherType: MatcherTypeTaskDefinition,
					Port:        1,
					MetricsPath: "/metrics2",
				},
			},
		},
		{
			ContainerIndex: 1,
			Targets: []MatchedTarget{
				{
					MatcherType: MatcherTypeDockerLabel,
					Port:        2,
				},
			},
		},
	}, task.Matched)
}
