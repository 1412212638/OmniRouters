package controller

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/model"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupSub2APIBalanceTest(t *testing.T) {
	t.Helper()
	previous := model.DB
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	model.DB = database
	require.NoError(t, database.AutoMigrate(&model.Channel{}))
	t.Cleanup(func() {
		model.DB = previous
		raw, _ := database.DB()
		_ = raw.Close()
	})
}

func TestUpdateChannelSub2APIBalancePrefersRemaining(t *testing.T) {
	setupSub2APIBalanceTest(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/usage", r.URL.Path)
		assert.Equal(t, "Bearer sk-test", r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"balance":100,"remaining":986.17,"isValid":true}`))
	}))
	defer server.Close()

	baseURL := server.URL
	channel := &model.Channel{Type: constant.ChannelTypeSub2API, BaseURL: &baseURL, Key: "sk-test"}
	require.NoError(t, model.DB.Create(channel).Error)

	balance, err := updateChannelSub2APIBalance(channel)
	require.NoError(t, err)
	require.InDelta(t, 986.17, balance, 1e-9)
	require.InDelta(t, 986.17, channel.Balance, 1e-9)
}

func TestUpdateChannelSub2APIBalanceRejectsInvalidKey(t *testing.T) {
	setupSub2APIBalanceTest(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"isValid":false}`))
	}))
	defer server.Close()

	baseURL := server.URL
	channel := &model.Channel{Type: constant.ChannelTypeSub2API, BaseURL: &baseURL, Key: "sk-test"}
	_, err := updateChannelSub2APIBalance(channel)
	require.EqualError(t, err, "Sub2API key is invalid (isValid=false)")
}
