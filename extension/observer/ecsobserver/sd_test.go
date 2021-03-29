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
	"context"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewDiscovery(t *testing.T) {
	t.Run("empty impl", func(t *testing.T) {
		_, err := NewDiscovery(ExampleConfig(), ServiceDiscoveryOptions{})
		require.NoError(t, err)
	})
	t.Run("for the coverage", func(t *testing.T) {
		d := ServiceDiscovery{}
		_, err := d.Discover(context.TODO())
		require.Error(t, err)
	})
}

// Util Start

func mustReadFile(t *testing.T, p string) []byte {
	b, err := ioutil.ReadFile(p)
	require.NoError(t, err, p)
	return b
}

func newMatcher(t *testing.T, cfg MatcherConfig) Matcher {
	require.NoError(t, cfg.Init())
	m, err := cfg.NewMatcher(testMatcherOptions())
	require.NoError(t, err)
	return m
}

func testMatcherOptions() MatcherOptions {
	return MatcherOptions{
		Logger: zap.NewExample(),
	}
}

// Util End
