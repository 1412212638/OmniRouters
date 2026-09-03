package common

import (
	"context"
	"fmt"

	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/relaykit/types"
)

const maxConversionDiagnostics = 32

type conversionDiagnosticKey struct {
	code     string
	path     string
	severity types.ConversionDiagnosticSeverity
	from     types.RelayFormat
	to       types.RelayFormat
}

func (info *RelayInfo) RecordConversionDiagnostics(ctx context.Context, diagnostics []types.ConversionDiagnostic) {
	if info == nil || len(diagnostics) == 0 {
		return
	}
	if info.conversionDiagnosticKeys == nil {
		info.conversionDiagnosticKeys = make(map[conversionDiagnosticKey]struct{})
	}
	for _, diagnostic := range diagnostics {
		key := conversionDiagnosticKey{diagnostic.Code, diagnostic.Path, diagnostic.Severity, diagnostic.From, diagnostic.To}
		if _, ok := info.conversionDiagnosticKeys[key]; ok {
			continue
		}
		if len(info.conversionDiagnostics) >= maxConversionDiagnostics {
			info.conversionDiagnosticsTruncated = true
			continue
		}
		info.conversionDiagnosticKeys[key] = struct{}{}
		info.conversionDiagnostics = append(info.conversionDiagnostics, diagnostic)
		logger.LogWarn(ctx, fmt.Sprintf("conversion diagnostic: code=%q severity=%q from=%q to=%q path=%q message=%q", diagnostic.Code, diagnostic.Severity, diagnostic.From, diagnostic.To, diagnostic.Path, diagnostic.Message))
	}
}

func (info *RelayInfo) ConversionDiagnostics() []types.ConversionDiagnostic {
	if info == nil {
		return nil
	}
	return append([]types.ConversionDiagnostic(nil), info.conversionDiagnostics...)
}

func (info *RelayInfo) ConversionDiagnosticsTruncated() bool {
	return info != nil && info.conversionDiagnosticsTruncated
}
