package model

import "testing"

func TestBuildAvailableVendorIDsMergesChannelSuppliers(t *testing.T) {
	meta := map[string]*Model{
		"glm-5.3": {Id: 1, VendorID: 7},
	}
	abilities := []AbilityWithChannel{
		{Ability: Ability{Model: "glm-5.3"}, ChannelVendorID: intPointer(9)},
		{Ability: Ability{Model: "glm-5.3"}, ChannelVendorID: intPointer(8)},
		{Ability: Ability{Model: "glm-5.3"}, ChannelVendorID: intPointer(9)},
	}

	got := buildAvailableVendorIDs(meta, map[int][]int{1: {2, 3}}, abilities)
	want := []int{2, 3, 9, 8}
	if len(got["glm-5.3"]) != len(want) {
		t.Fatalf("unexpected vendor count: %#v", got["glm-5.3"])
	}
	for i, id := range want {
		if got["glm-5.3"][i] != id {
			t.Fatalf("unexpected vendor order: %#v", got["glm-5.3"])
		}
	}
}

func TestBuildAvailableVendorIDsUsesLegacyVendorWhenUnconfigured(t *testing.T) {
	meta := map[string]*Model{
		"glm-5.3": {Id: 1, VendorID: 7},
	}
	got := buildAvailableVendorIDs(meta, map[int][]int{}, nil)
	if len(got["glm-5.3"]) != 1 || got["glm-5.3"][0] != 7 {
		t.Fatalf("unexpected legacy vendor fallback: %#v", got["glm-5.3"])
	}
}

func TestFilterAvailableVendorIDsSkipsDisabledVendors(t *testing.T) {
	available := map[string][]int{
		"glm-5.3": {2, 7, 9, 7},
		"deepseek-v4.1-flash": {9},
	}
	enabled := map[int]struct{}{2: {}, 9: {}}

	got := filterAvailableVendorIDs(available, enabled)
	want := map[string][]int{
		"glm-5.3":             {2, 9},
		"deepseek-v4.1-flash": {9},
	}
	for name, expected := range want {
		actual := got[name]
		if len(actual) != len(expected) {
			t.Fatalf("unexpected vendors for %s: %#v", name, actual)
		}
		for i, id := range expected {
			if actual[i] != id {
				t.Fatalf("unexpected vendors for %s: %#v", name, actual)
			}
		}
	}
	if _, exists := got["disabled-model"]; exists {
		t.Fatal("disabled-only model should not be added")
	}
}

func intPointer(value int) *int { return &value }
