package ecssd

import (
	"context"
	"io/ioutil"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestServiceDiscovery_RunAndWriteFile(t *testing.T) {
	genTasks := func() []*Task {
		return []*Task{
			{
				Task: &ecs.Task{
					TaskDefinitionArn: aws.String("t1"),
					Containers: []*ecs.Container{
						{
							Name: aws.String("c1-t1"),
						},
						{
							Name: aws.String("c2-t1"),
							NetworkBindings: []*ecs.NetworkBinding{
								{
									ContainerPort: aws.Int64(1008),
									HostPort:      aws.Int64(8008),
								},
							},
						},
					},
				},
				Definition: &ecs.TaskDefinition{
					NetworkMode: aws.String(ecs.NetworkModeBridge),
					ContainerDefinitions: []*ecs.ContainerDefinition{
						{
							Name: aws.String("c1-t1"),
						},
						{
							Name: aws.String("c2-t1"),
							PortMappings: []*ecs.PortMapping{
								{
									ContainerPort: aws.Int64(1008),
								},
							},
						},
					},
				},
				Service: &ecs.Service{
					ServiceName: aws.String("s1"),
				},
				EC2: &ec2.Instance{
					PrivateIpAddress: aws.String("172.168.0.1"),
				},
			},
			{
				Task: &ecs.Task{
					TaskDefinitionArn: aws.String("t2"),
					Attachments: []*ecs.Attachment{
						{
							Type: aws.String("ElasticNetworkInterface"),
							Details: []*ecs.KeyValuePair{
								{
									Name:  aws.String("privateIPv4Address"),
									Value: aws.String("172.168.1.1"),
								},
							},
						},
					},
				},
				Definition: &ecs.TaskDefinition{
					NetworkMode: aws.String(ecs.NetworkModeAwsvpc),
					ContainerDefinitions: []*ecs.ContainerDefinition{
						{
							Name: aws.String("c1-t2"),
							DockerLabels: map[string]*string{
								"NOT_PORT": aws.String("just a value"),
							},
						},
						{
							Name: aws.String("c2-t2"),
							DockerLabels: map[string]*string{
								"PROMETHEUS_PORT": aws.String("2112"),
							},
							PortMappings: []*ecs.PortMapping{
								{
									ContainerPort: aws.Int64(2112),
									HostPort:      aws.Int64(8112),
								},
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

	outputFile := "testdata/ut_targets.yaml"
	cfg := Config{
		ClusterName:     "ut-cluster-1",
		ClusterRegion:   "us-test-2",
		RefreshInterval: 100 * time.Millisecond,
		ResultFile:      outputFile,
		JobLabelName:    DefaultJobLabelName,
		Services: []ServiceConfig{
			{
				NamePattern: "s1",
				CommonExporterConfig: CommonExporterConfig{
					MetricsPorts: []int{1008},
					JobName:      "service-s1",
				},
			},
		},
		DockerLabels: []DockerLabelConfig{
			{
				PortLabel: "PROMETHEUS_PORT",
			},
		},
	}
	opts := ServiceDiscoveryOptions{
		Logger:          zap.NewExample(),
		FetcherOverride: newMockFetcher(genTasks),
	}
	sd, err := New(cfg, opts)
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()
	require.NoError(t, sd.RunAndWriteFile(ctx))
	assert.FileExists(t, outputFile)
	expectedFile := "testdata/ut_targets.expected.yaml"
	assert.Equal(t, string(mustReadFile(t, expectedFile)), string(mustReadFile(t, outputFile)))
}

func mustReadFile(t *testing.T, p string) []byte {
	b, err := ioutil.ReadFile(p)
	require.NoError(t, err, p)
	return b
}
