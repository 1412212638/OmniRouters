package relay

import (
	"strconv"
	"testing"

	"github.com/QuantumNous/new-api/constant"
	pluginruntime "github.com/QuantumNous/new-api/pkg/jsplugin"
	"github.com/QuantumNous/new-api/plugins"
	"github.com/QuantumNous/new-api/relay/channel"
	"github.com/QuantumNous/new-api/relay/channel/task/sora"
	"github.com/QuantumNous/new-api/relay/channel/task/suno"
)

func taskPluginTestRegistry(t *testing.T, key string) *pluginruntime.Registry {
	t.Helper()
	source, err := plugins.Source(key)
	if err != nil {
		t.Fatalf("read plugin %q: %v", key, err)
	}
	registry := pluginruntime.NewRegistry()
	if _, err = registry.RegisterFactory(source, pluginruntime.Options{Key: key}); err != nil {
		t.Fatalf("register plugin %q: %v", key, err)
	}
	return registry
}

func TestGetTaskAdaptorFallsBackWhenPluginsDisabled(t *testing.T) {
	registry := taskPluginTestRegistry(t, "sora")
	factoryCalled := false
	factory := func(_ *pluginruntime.LoadedPlugin) channel.TaskAdaptor {
		factoryCalled = true
		return &suno.TaskAdaptor{}
	}

	platform := constant.TaskPlatform(strconv.Itoa(constant.ChannelTypeSora))
	adaptor := getTaskAdaptor(platform, false, registry, factory)
	if _, ok := adaptor.(*sora.TaskAdaptor); !ok {
		t.Fatalf("expected legacy Sora adaptor, got %T", adaptor)
	}
	if factoryCalled {
		t.Fatal("plugin factory must not be called while plugins are disabled")
	}
}

func TestGetTaskAdaptorFallsBackWithoutFactory(t *testing.T) {
	registry := taskPluginTestRegistry(t, "sora")

	platform := constant.TaskPlatform(strconv.Itoa(constant.ChannelTypeSora))
	adaptor := getTaskAdaptor(platform, true, registry, nil)
	if _, ok := adaptor.(*sora.TaskAdaptor); !ok {
		t.Fatalf("expected legacy Sora adaptor, got %T", adaptor)
	}
}

func TestGetTaskAdaptorUsesResolvedPlugin(t *testing.T) {
	registry := taskPluginTestRegistry(t, "sora")
	var resolvedKey string
	factory := func(plugin *pluginruntime.LoadedPlugin) channel.TaskAdaptor {
		resolvedKey = plugin.Meta.Key
		return &suno.TaskAdaptor{}
	}

	platform := constant.TaskPlatform(strconv.Itoa(constant.ChannelTypeSora))
	adaptor := getTaskAdaptor(platform, true, registry, factory)
	if _, ok := adaptor.(*suno.TaskAdaptor); !ok {
		t.Fatalf("expected plugin adaptor marker, got %T", adaptor)
	}
	if resolvedKey != "sora" {
		t.Fatalf("expected sora plugin, got %q", resolvedKey)
	}
}

func TestGetTaskAdaptorFallsBackForMissingOrDisabledPlugin(t *testing.T) {
	registry := taskPluginTestRegistry(t, "sora")
	factoryCalled := false
	factory := func(_ *pluginruntime.LoadedPlugin) channel.TaskAdaptor {
		factoryCalled = true
		return &suno.TaskAdaptor{}
	}

	missing := getTaskAdaptor(constant.TaskPlatform("999999"), true, registry, factory)
	if missing != nil {
		t.Fatalf("expected no adaptor for unknown platform, got %T", missing)
	}

	registry.SetDisabledFactoryKeys([]string{"sora"})
	platform := constant.TaskPlatform(strconv.Itoa(constant.ChannelTypeSora))
	disabled := getTaskAdaptor(platform, true, registry, factory)
	if _, ok := disabled.(*sora.TaskAdaptor); !ok {
		t.Fatalf("expected legacy Sora adaptor for disabled plugin, got %T", disabled)
	}
	if factoryCalled {
		t.Fatal("plugin factory must not be called for missing or disabled plugins")
	}
}
