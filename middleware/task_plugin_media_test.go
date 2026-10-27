package middleware

import (
	"testing"

	"github.com/QuantumNous/new-api/pkg/jsplugin"
	"github.com/stretchr/testify/require"
)

func TestTaskPluginRouteBodyAllowed(t *testing.T) {
	tests := []struct {
		name         string
		routeType    string
		bodyKind     string
		modelsScoped bool
		allowed      bool
	}{
		{"submit json", string(jsplugin.RouteTypeSubmit), string(jsplugin.BodyJSON), false, true},
		{"submit multipart", string(jsplugin.RouteTypeSubmit), string(jsplugin.BodyMultipart), false, true},
		{"scoped submit multipart", string(jsplugin.RouteTypeSubmit), string(jsplugin.BodyMultipart), true, false},
		{"query none", string(jsplugin.RouteTypeQuery), string(jsplugin.BodyNone), false, true},
		{"query json", string(jsplugin.RouteTypeQuery), string(jsplugin.BodyJSON), false, false},
		{"submit form", string(jsplugin.RouteTypeSubmit), string(jsplugin.BodyForm), false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.allowed, taskPluginRouteBodyAllowed(tt.routeType, tt.bodyKind, tt.modelsScoped))
		})
	}
}
