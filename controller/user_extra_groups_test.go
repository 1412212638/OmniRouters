package controller

import (
	"testing"

	"github.com/QuantumNous/new-api/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeUserExtraGroups(t *testing.T) {
	groups, err := normalizeUserExtraGroups(
		"default",
		model.StringList{"svip", " svip ", "default", "auto", ""},
	)

	require.NoError(t, err)
	assert.Equal(t, model.StringList{"svip"}, groups)
}

func TestNormalizeUserExtraGroupsRejectsUnknownGroup(t *testing.T) {
	_, err := normalizeUserExtraGroups(
		"default",
		model.StringList{"missing-group"},
	)

	require.Error(t, err)
}
