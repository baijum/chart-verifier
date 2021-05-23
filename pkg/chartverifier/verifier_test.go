/*
 * Copyright 2021 Red Hat
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package chartverifier

import (
	"context"
	"errors"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"helm.sh/helm/v3/pkg/cli"

	"github.com/redhat-certification/chart-verifier/pkg/chartverifier/checks"
	"github.com/redhat-certification/chart-verifier/pkg/testutil"
)

type ocVersionError struct{}

func (v *ocVersionError) getVersion(debug bool) (string, error) {
	return "", errors.New("error")
}

type ocVersionWithoutError struct{}

func (v *ocVersionWithoutError) getVersion(debug bool) (string, error) {
	return "4.9.7", nil
}

func (c *Report) isOk() bool {
	outcome := true
	for _, check := range c.Results {
		if !(check.Outcome == PassOutcomeType) {
			outcome = false
			break
		}
	}
	return outcome
}

func TestVerifier_Verify(t *testing.T) {

	addr := "127.0.0.1:9876"
	ctx, cancel := context.WithCancel(context.Background())

	require.NoError(t, testutil.ServeCharts(ctx, addr, "./checks/"))

	dummyCheckName := "dummy-check"

	erroredCheck := func(_ *checks.CheckOptions) (checks.Result, error) {
		return checks.Result{}, errors.New("artificial error")
	}

	negativeCheck := func(_ *checks.CheckOptions) (checks.Result, error) {
		return checks.Result{Ok: false}, nil
	}

	positiveCheck := func(_ *checks.CheckOptions) (checks.Result, error) {
		return checks.Result{Ok: true}, nil
	}

	validChartUri := "http://" + addr + "/charts/chart-0.1.0-v3.valid.tgz"

	verocVersionWithoutError := &ocVersionWithoutError{}
	verocVersionError := &ocVersionError{}

	t.Run("Should return error if check does not exist", func(t *testing.T) {
		c := &verifier{
			settings:       cli.New(),
			config:         viper.New(),
			registry:       checks.NewRegistry(),
			requiredChecks: []string{dummyCheckName},
			version:        verocVersionWithoutError,
		}

		r, err := c.Verify(validChartUri)
		require.Error(t, err)
		require.Nil(t, r)
	})

	t.Run("Should return error if check exists and returns error", func(t *testing.T) {
		c := &verifier{
			settings:       cli.New(),
			config:         viper.New(),
			registry:       checks.NewRegistry().Add(checks.Check{Name: dummyCheckName, Type: MandatoryCheckType, Func: erroredCheck}),
			requiredChecks: []string{dummyCheckName},
			version:        verocVersionWithoutError,
		}

		r, err := c.Verify(validChartUri)
		require.Error(t, err)
		require.Nil(t, r)
	})

	t.Run("Result should be negative if check exists and returns negative", func(t *testing.T) {

		c := &verifier{
			settings:         cli.New(),
			config:           viper.New(),
			registry:         checks.NewRegistry().Add(checks.Check{Name: dummyCheckName, Type: MandatoryCheckType, Func: negativeCheck}),
			requiredChecks:   []string{dummyCheckName},
			openshiftVersion: "4.9",
			version:          verocVersionWithoutError,
		}

		r, err := c.Verify(validChartUri)
		require.NoError(t, err)
		require.NotNil(t, r)
		require.False(t, r.isOk())
	})

	t.Run("Result should be positive if check exists and returns positive", func(t *testing.T) {
		c := &verifier{
			settings:         cli.New(),
			config:           viper.New(),
			registry:         checks.NewRegistry().Add(checks.Check{Name: dummyCheckName, Type: MandatoryCheckType, Func: positiveCheck}),
			requiredChecks:   []string{dummyCheckName},
			openshiftVersion: "4.9",
			version:          verocVersionWithoutError,
		}

		r, err := c.Verify(validChartUri)
		require.NoError(t, err)
		require.NotNil(t, r)
		require.True(t, r.isOk())
	})

	chartTestingCheckName := "chart-testing"

	t.Run("oc version error and wrong user input", func(t *testing.T) {
		c := &verifier{
			settings:         cli.New(),
			config:           viper.New(),
			registry:         checks.NewRegistry().Add(checks.Check{Name: chartTestingCheckName, Type: MandatoryCheckType, Func: positiveCheck}),
			requiredChecks:   []string{chartTestingCheckName},
			openshiftVersion: "NaN",
			version:          verocVersionError,
		}
		r, err := c.Verify(validChartUri)
		require.Error(t, err)
		require.Nil(t, r)
	})

	t.Run("oc version error and correct user input", func(t *testing.T) {
		c := &verifier{
			settings:         cli.New(),
			config:           viper.New(),
			registry:         checks.NewRegistry().Add(checks.Check{Name: chartTestingCheckName, Type: MandatoryCheckType, Func: positiveCheck}),
			requiredChecks:   []string{chartTestingCheckName},
			openshiftVersion: "4.9.7",
			version:          verocVersionError,
		}

		r, err := c.Verify(validChartUri)
		require.NoError(t, err)
		require.NotNil(t, r)
		require.True(t, r.isOk())
		require.Equal(t, "4.9.7", r.Metadata.ToolMetadata.CertifiedOpenShiftVersions)
	})

	t.Run("oc version and wrong user input", func(t *testing.T) {
		c := &verifier{
			settings:         cli.New(),
			config:           viper.New(),
			registry:         checks.NewRegistry().Add(checks.Check{Name: chartTestingCheckName, Type: MandatoryCheckType, Func: positiveCheck}),
			requiredChecks:   []string{chartTestingCheckName},
			openshiftVersion: "NaN",
			version:          verocVersionWithoutError,
		}

		r, err := c.Verify(validChartUri)
		require.NoError(t, err)
		require.NotNil(t, r)
		require.True(t, r.isOk())
		require.Equal(t, "4.9.7", r.Metadata.ToolMetadata.CertifiedOpenShiftVersions)
	})

	t.Run("oc version and correct user input", func(t *testing.T) {
		c := &verifier{
			settings:         cli.New(),
			config:           viper.New(),
			registry:         checks.NewRegistry().Add(checks.Check{Name: chartTestingCheckName, Type: MandatoryCheckType, Func: positiveCheck}),
			requiredChecks:   []string{chartTestingCheckName},
			openshiftVersion: "5.6.8",
			version:          verocVersionWithoutError,
		}

		r, err := c.Verify(validChartUri)
		require.NoError(t, err)
		require.NotNil(t, r)
		require.True(t, r.isOk())
		require.Equal(t, "4.9.7", r.Metadata.ToolMetadata.CertifiedOpenShiftVersions)
	})

	t.Run("oc version error and empty user input", func(t *testing.T) {
		c := &verifier{
			settings:         cli.New(),
			config:           viper.New(),
			registry:         checks.NewRegistry().Add(checks.Check{Name: chartTestingCheckName, Type: MandatoryCheckType, Func: positiveCheck}),
			requiredChecks:   []string{chartTestingCheckName},
			openshiftVersion: "",
			version:          verocVersionError,
		}

		r, err := c.Verify(validChartUri)
		require.Error(t, err)
		require.Nil(t, r)
	})

	t.Run("dummy-check oc version error and wrong user input", func(t *testing.T) {
		c := &verifier{
			settings:         cli.New(),
			config:           viper.New(),
			registry:         checks.NewRegistry().Add(checks.Check{Name: dummyCheckName, Type: MandatoryCheckType, Func: positiveCheck}),
			requiredChecks:   []string{dummyCheckName},
			openshiftVersion: "NaN",
			version:          verocVersionError,
		}
		r, err := c.Verify(validChartUri)
		require.Error(t, err)
		require.Nil(t, r)
	})

	t.Run("dummy-check oc version error and empty user input", func(t *testing.T) {
		c := &verifier{
			settings:         cli.New(),
			config:           viper.New(),
			registry:         checks.NewRegistry().Add(checks.Check{Name: dummyCheckName, Type: MandatoryCheckType, Func: positiveCheck}),
			requiredChecks:   []string{dummyCheckName},
			openshiftVersion: "",
			version:          verocVersionError,
		}
		r, err := c.Verify(validChartUri)
		require.NoError(t, err)
		require.NotNil(t, r)
		require.True(t, r.isOk())
		require.Equal(t, "", r.Metadata.ToolMetadata.CertifiedOpenShiftVersions)
	})

	cancel()
}