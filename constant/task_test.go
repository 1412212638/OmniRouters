package constant

import "testing"

func TestNormalizeTaskActionPreservesPluginVocabularyAndLegacyCompatibility(t *testing.T) {
	tests := map[string]string{
		TaskActionGenerate:          TaskActionImageToVideo,
		TaskActionTextGenerate:      TaskActionTextToVideo,
		TaskActionFirstTailGenerate: TaskActionFirstTailToVideo,
		TaskActionReferenceGenerate: TaskActionReferenceToVideo,
		TaskActionRemix:             TaskActionRemixCanonical,
		"MUSIC":                     "MUSIC",
	}
	for input, want := range tests {
		if got := NormalizeTaskAction(input); got != want {
			t.Errorf("NormalizeTaskAction(%q) = %q, want %q", input, got, want)
		}
	}
}
