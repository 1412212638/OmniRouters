package common

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRateLimitReservationReleaseCommitAndExpiry(t *testing.T) {
	var limiter InMemoryRateLimiter
	limiter.Init(0)
	first := limiter.Reserve("user", 1, 60)
	require.NotNil(t, first)
	require.Nil(t, limiter.Reserve("user", 1, 60))
	first.Complete(false)
	first.Complete(true) // Deferred cleanup must not complete twice.
	second := limiter.Reserve("user", 1, 60)
	require.NotNil(t, second)
	second.Complete(true)
	second.Complete(false)
	require.Empty(t, limiter.reservations)
	require.Nil(t, limiter.Reserve("user", 1, 60))
	limiter.store["user"].requests.head.timestamp = time.Now().Unix() - 60
	next := limiter.Reserve("user", 1, 60)
	require.NotNil(t, next)
	next.Complete(false)
	var absent *RateLimitReservation
	absent.Complete(true)
}

func TestRateLimitReservationSurvivesIdleEviction(t *testing.T) {
	for _, success := range []bool{false, true} {
		var limiter InMemoryRateLimiter
		limiter.Init(0)
		limiter.expirationDuration = time.Second
		reservation := limiter.Reserve("slow", 1, 60)
		require.NotNil(t, reservation)
		limiter.deleteExpiredEntries(time.Now().Add(time.Minute))
		require.Empty(t, limiter.store)
		require.Nil(t, limiter.Reserve("slow", 1, 60), "in-flight request retains admission")
		reservation.Complete(success)
		require.Empty(t, limiter.reservations)
		next := limiter.Reserve("slow", 1, 60)
		if success {
			require.Nil(t, next)
		} else {
			require.NotNil(t, next)
			next.Complete(false)
		}
	}
}

func TestRateLimitReservationConcurrentAdmission(t *testing.T) {
	var limiter InMemoryRateLimiter
	limiter.Init(0)
	var wg sync.WaitGroup
	admitted := make(chan *RateLimitReservation, 100)
	for range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if reservation := limiter.Reserve("shared", 7, 60); reservation != nil {
				admitted <- reservation
			}
		}()
	}
	wg.Wait()
	close(admitted)
	require.Len(t, admitted, 7)
	for reservation := range admitted {
		reservation.Complete(false)
	}
	require.Empty(t, limiter.reservations)
}

func TestRateLimitRejectsNonPositiveCapacityWithoutAllocation(t *testing.T) {
	var limiter InMemoryRateLimiter
	limiter.Init(0)
	for _, capacity := range []int{0, -1} {
		require.False(t, limiter.Request("invalid", capacity, 60))
		require.Nil(t, limiter.Reserve("invalid", capacity, 60))
	}
	require.Empty(t, limiter.store)
}
