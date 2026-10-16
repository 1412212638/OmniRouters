package vertex

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildAPIBaseURLUsesMultiRegionHosts(t *testing.T) {
	tests := []struct {
		region, projectID, want string
	}{
		{"us", "project-1", "https://aiplatform.us.rep.googleapis.com/v1/projects/project-1/locations/us"},
		{"eu", "project-2", "https://aiplatform.eu.rep.googleapis.com/v1/projects/project-2/locations/eu"},
		{"us-central1", "project-3", "https://us-central1-aiplatform.googleapis.com/v1/projects/project-3/locations/us-central1"},
		{"global", "", "https://aiplatform.googleapis.com/v1"},
	}
	for _, test := range tests {
		assert.Equal(t, test.want, BuildAPIBaseURL("", DefaultAPIVersion, test.projectID, test.region))
	}
}
