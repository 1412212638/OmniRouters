package constant

type TaskPlatform string

const (
	TaskPlatformSuno       TaskPlatform = "suno"
	TaskPlatformMidjourney              = "mj"
	TaskPlatformOpenAIImage             = "openai_image"
)

const (
	SunoActionMusic  = "MUSIC"
	SunoActionLyrics = "LYRICS"

	// Canonical task-plugin actions. The legacy constants below remain stable
	// for native adaptors and persisted task rows.
	TaskActionImageToVideo     = "image_to_video"
	TaskActionTextToVideo      = "text_to_video"
	TaskActionFirstTailToVideo = "first_tail_to_video"
	TaskActionReferenceToVideo = "reference_to_video"
	TaskActionRemixCanonical  = "remix"

	TaskActionGenerate          = "generate"
	TaskActionTextGenerate      = "textGenerate"
	TaskActionFirstTailGenerate = "firstTailGenerate"
	TaskActionReferenceGenerate = "referenceGenerate"
	TaskActionRemix             = "remixGenerate"
)

var legacyTaskActionAliases = map[string]string{
	TaskActionGenerate:          TaskActionImageToVideo,
	TaskActionTextGenerate:      TaskActionTextToVideo,
	TaskActionFirstTailGenerate: TaskActionFirstTailToVideo,
	TaskActionReferenceGenerate: TaskActionReferenceToVideo,
	TaskActionRemix:             TaskActionRemixCanonical,
}

// NormalizeTaskAction maps legacy persisted/native action names to the
// canonical vocabulary exposed to task plugins. Unknown platform-specific
// actions pass through unchanged.
func NormalizeTaskAction(action string) string {
	if canonical, ok := legacyTaskActionAliases[action]; ok {
		return canonical
	}
	return action
}

var SunoModel2Action = map[string]string{
	"suno_music":  SunoActionMusic,
	"suno_lyrics": SunoActionLyrics,
}
