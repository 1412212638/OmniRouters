package service

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"

	"github.com/bytedance/gopkg/util/gopool"
)

const (
	ticketMaintenanceTickInterval = 5 * time.Minute
	ticketMaintenanceBatchSize    = 500
)

var (
	ticketMaintenanceOnce    sync.Once
	ticketMaintenanceRunning atomic.Bool
)

func StartTicketMaintenanceTask() {
	ticketMaintenanceOnce.Do(func() {
		if !common.IsMasterNode {
			return
		}
		gopool.Go(func() {
			logger.LogInfo(context.Background(), fmt.Sprintf("ticket maintenance task started: tick=%s", ticketMaintenanceTickInterval))
			ticker := time.NewTicker(ticketMaintenanceTickInterval)
			defer ticker.Stop()

			runTicketMaintenanceOnce()
			for range ticker.C {
				runTicketMaintenanceOnce()
			}
		})
	})
}

func runTicketMaintenanceOnce() {
	if common.TicketIdleCloseHours <= 0 {
		return
	}
	if !ticketMaintenanceRunning.CompareAndSwap(false, true) {
		return
	}
	defer ticketMaintenanceRunning.Store(false)

	ctx := context.Background()
	now := time.Now().Unix()
	cutoff := now - int64(common.TicketIdleCloseHours)*3600
	reopenUntil := int64(0)
	if window := ticketReopenWindowSeconds(); window > 0 {
		reopenUntil = now + window
	}

	totalClosed := 0
	for {
		n, err := model.CloseIdleAnsweredTickets(cutoff, now, reopenUntil, ticketMaintenanceBatchSize)
		if err != nil {
			logger.LogWarn(ctx, fmt.Sprintf("ticket maintenance task failed: %v", err))
			return
		}
		if n == 0 {
			break
		}
		totalClosed += n
		if n < ticketMaintenanceBatchSize {
			break
		}
	}
	if common.DebugEnabled && totalClosed > 0 {
		logger.LogDebug(ctx, "ticket maintenance: auto_closed_count=%d", totalClosed)
	}
}
