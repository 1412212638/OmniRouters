package common

import (
	"io"

	"github.com/QuantumNous/new-api/common"
)

// NewOutboundJSONBody wraps an already-marshaled upstream request body into a
// BodyStorage. When disk cache is enabled and the payload exceeds the configured
// threshold, the data is written to a temp file and the original []byte can be
// garbage-collected while the upstream request is in flight.
//
// The caller must close the returned closer after the upstream call finishes.
// The reader is wrapped so net/http does not close the BodyStorage before the
// relay handler reaches its deferred cleanup.
func NewOutboundJSONBody(data []byte) (body io.Reader, size int64, closer io.Closer, err error) {
	storage, err := common.CreateBodyStorage(data)
	if err != nil {
		return nil, 0, nil, err
	}
	return common.ReaderOnly(storage), storage.Size(), storage, nil
}
