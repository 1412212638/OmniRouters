package setting

import (
	"fmt"
	"math"
	"sync"

	"github.com/QuantumNous/new-api/common"
)

// maxRateLimitDurationSeconds is the largest window the count cap is computed
// against (24h). Token-bucket capacity is count*duration; this keeps that
// product inside int64 when the window is at most a day.
const maxRateLimitDurationSeconds = 24 * 60 * 60

// maxModelRequestRateLimitCount is math.MaxInt64 / maxRateLimitDurationSeconds.
// It is the largest count that cannot overflow int64(count)*duration for a
// window of at most 24 hours.
const maxModelRequestRateLimitCount int64 = math.MaxInt64 / maxRateLimitDurationSeconds

var ModelRequestRateLimitEnabled = false
var ModelRequestRateLimitDurationMinutes = 1
var ModelRequestRateLimitCount = 0
var ModelRequestRateLimitSuccessCount = 1000
var ModelRequestRateLimitRPM = 0
var ModelRequestRateLimitTPM = 0
var ModelRequestRateLimitRPD = 0

// ModelRateLimit contains the public, independent rate-limit windows. A zero
// value means that the corresponding limit is unlimited.
type ModelRateLimit struct {
	RPM int `json:"rpm"`
	TPM int `json:"tpm"`
	RPD int `json:"rpd"`
}

var ModelRequestRateLimitGroup = map[string]ModelRateLimit{}
var ModelRequestRateLimitMutex sync.RWMutex

func ModelRequestRateLimitGroup2JSONString() string {
	ModelRequestRateLimitMutex.RLock()
	defer ModelRequestRateLimitMutex.RUnlock()

	jsonBytes, err := common.Marshal(ModelRequestRateLimitGroup)
	if err != nil {
		common.SysLog("error marshalling model ratio: " + err.Error())
	}
	return string(jsonBytes)
}

func UpdateModelRequestRateLimitGroupByJSONString(jsonStr string) error {
	parsed := make(map[string]commonRateLimitValue)
	if err := common.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		return err
	}
	groups := make(map[string]ModelRateLimit, len(parsed))
	for group, value := range parsed {
		limit := value.toModelRateLimit(ModelRequestRateLimitDurationMinutes)
		if err := validateModelRateLimit(group, limit); err != nil {
			return err
		}
		groups[group] = limit
	}

	ModelRequestRateLimitMutex.Lock()
	defer ModelRequestRateLimitMutex.Unlock()
	ModelRequestRateLimitGroup = groups
	return nil
}

// commonRateLimitValue accepts both the new object format and the legacy
// [maxRequests, maxSuccess] array format during upgrades.
type commonRateLimitValue struct {
	RPM int
	TPM int
	RPD int
	Legacy []int
}

func (v *commonRateLimitValue) UnmarshalJSON(data []byte) error {
	var object struct {
		RPM int `json:"rpm"`
		TPM int `json:"tpm"`
		RPD int `json:"rpd"`
	}
	if err := common.Unmarshal(data, &object); err == nil {
		v.RPM, v.TPM, v.RPD = object.RPM, object.TPM, object.RPD
		var probe []int
		if err := common.Unmarshal(data, &probe); err == nil {
			v.Legacy = probe
		}
		return nil
	}
	var legacy []int
	if err := common.Unmarshal(data, &legacy); err != nil {
		return err
	}
	v.Legacy = legacy
	return nil
}

func (v commonRateLimitValue) toModelRateLimit(durationMinutes int) ModelRateLimit {
	if len(v.Legacy) >= 2 {
		if durationMinutes <= 0 {
			durationMinutes = 1
		}
		return ModelRateLimit{
			RPM: int((int64(v.Legacy[0])*60 + int64(durationMinutes) - 1) / int64(durationMinutes)),
			RPD: int((int64(v.Legacy[1])*24*60 + int64(durationMinutes) - 1) / int64(durationMinutes)),
		}
	}
	return ModelRateLimit{RPM: v.RPM, TPM: v.TPM, RPD: v.RPD}
}

func validateModelRateLimit(group string, limit ModelRateLimit) error {
	if limit.RPM < 0 || limit.TPM < 0 || limit.RPD < 0 {
		return fmt.Errorf("group %s has negative rate limit values: [%d, %d, %d]", group, limit.RPM, limit.TPM, limit.RPD)
	}
	if int64(limit.RPM) > maxModelRequestRateLimitCount || int64(limit.TPM) > maxModelRequestRateLimitCount || int64(limit.RPD) > maxModelRequestRateLimitCount {
		return fmt.Errorf("group %s exceeds max rate limit %d", group, maxModelRequestRateLimitCount)
	}
	return nil
}

func CheckModelRequestRateLimitValue(value int) error {
	if value < 0 {
		return fmt.Errorf("rate limit must be non-negative")
	}
	if int64(value) > maxModelRequestRateLimitCount {
		return fmt.Errorf("rate limit exceeds max value %d", maxModelRequestRateLimitCount)
	}
	return nil
}

func GetModelRateLimit(group string) (ModelRateLimit, bool) {
	ModelRequestRateLimitMutex.RLock()
	defer ModelRequestRateLimitMutex.RUnlock()

	if ModelRequestRateLimitGroup == nil {
		return ModelRateLimit{}, false
	}
	limits, found := ModelRequestRateLimitGroup[group]
	return limits, found
}

func GetGroupRateLimit(group string) (totalCount, successCount int, found bool) {
	limits, found := GetModelRateLimit(group)
	return limits.RPM, limits.RPD, found
}

func GetModelRateLimits(groups []string) map[string]ModelRateLimit {
	ModelRequestRateLimitMutex.RLock()
	defer ModelRequestRateLimitMutex.RUnlock()

	global := EffectiveGlobalRateLimit()
	result := make(map[string]ModelRateLimit, len(groups))
	for _, group := range groups {
		if group == "" || group == "auto" {
			continue
		}
		limit := global
		if ModelRequestRateLimitEnabled {
			if configured, ok := ModelRequestRateLimitGroup[group]; ok {
				limit = configured
			}
		}
		result[group] = limit
	}
	return result
}

// EffectiveGlobalRateLimit exposes the new windows while keeping old installs
// visible until their legacy options are replaced in the settings UI.
func EffectiveGlobalRateLimit() ModelRateLimit {
	if !ModelRequestRateLimitEnabled {
		return ModelRateLimit{}
	}
	if ModelRequestRateLimitRPM != 0 || ModelRequestRateLimitTPM != 0 || ModelRequestRateLimitRPD != 0 {
		return ModelRateLimit{RPM: ModelRequestRateLimitRPM, TPM: ModelRequestRateLimitTPM, RPD: ModelRequestRateLimitRPD}
	}
	duration := ModelRequestRateLimitDurationMinutes
	if duration <= 0 {
		return ModelRateLimit{}
	}
	return ModelRateLimit{
		RPM: int((int64(ModelRequestRateLimitCount)*60 + int64(duration) - 1) / int64(duration)),
		RPD: int((int64(ModelRequestRateLimitSuccessCount)*24*60 + int64(duration) - 1) / int64(duration)),
	}
}

func CheckModelRequestRateLimitGroup(jsonStr string) error {
	checkModelRequestRateLimitGroup := make(map[string]commonRateLimitValue)
	err := common.Unmarshal([]byte(jsonStr), &checkModelRequestRateLimitGroup)
	if err != nil {
		return err
	}
	for group, value := range checkModelRequestRateLimitGroup {
		if err := validateModelRateLimit(group, value.toModelRateLimit(ModelRequestRateLimitDurationMinutes)); err != nil {
			return err
		}
	}

	return nil
}
