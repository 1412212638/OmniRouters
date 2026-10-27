package dto

import (
	"testing"

	"github.com/QuantumNous/new-api/relaykit/relayconvert/kitutil"
	"github.com/QuantumNous/new-api/relaykit/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmbeddingTokenCountMetaIncludesMultimodalParts(t *testing.T) {
	const imageData = "data:image/png;base64,iVBORw0KGgo="
	var request EmbeddingRequest
	require.NoError(t, kitutil.UnmarshalJsonStr(`{"model":"m","input":["a cat",`+
		`{"type":"text","text":"on a mat"},`+
		`{"type":"image_url","image_url":{"url":"`+imageData+`","detail":"low"}},`+
		`{"type":"video_url","video_url":{"url":"https://example.test/cat.mp4"}}]}`, &request))

	meta := request.GetTokenCountMeta()
	assert.Equal(t, "a cat\non a mat", meta.CombineText)
	assert.Equal(t, []*types.FileMeta{
		{FileType: types.FileTypeImage, Source: types.NewBase64FileSource(imageData, ""), Detail: "low"},
		{FileType: types.FileTypeVideo, Source: types.NewURLFileSource("https://example.test/cat.mp4")},
	}, meta.Files)
}
