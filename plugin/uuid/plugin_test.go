package uuid

import (
	"testing"

	"github.com/cadence-workflow/starlark-worker/cadstar"
	"github.com/stretchr/testify/require"
)

func TestPluginFactory(t *testing.T) {
	stringDict := Plugin.Create(cadstar.RunInfo{})

	// Use require to validate the results
	require.NotNil(t, stringDict, "StringDict should not be nil")

	// Check that the "uuid" key exists in the dictionary
	require.Contains(t, stringDict, "uuid", "StringDict should contain 'uuid' key")

	// Assert that the value associated with "uuid" is of type *uuid.Module
	require.IsType(t, &Module{}, stringDict["uuid"], "The value associated with 'uuid' should be of type *uuid.Module")
}
