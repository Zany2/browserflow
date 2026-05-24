package tasklock

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
)

const (
	// LeaseTTL is longer than the client heartbeat interval and short enough to self-heal stale locks.
	LeaseTTL = 10 * time.Minute
	// StaleAfter gives the timeout scanner a small buffer after the Redis lease can expire.
	StaleAfter = LeaseTTL + 2*time.Minute
	// KeyPrefix is the Redis prefix for client execution locks.
	KeyPrefix = "browserflow:client:task-lock:"
)

// LockInfo describes one client execution lease.
type LockInfo struct {
	ClientIP   string `json:"client_ip"`
	MachineID  string `json:"machine_id,omitempty"`
	NodeID     string `json:"node_id,omitempty"`
	TaskID     int64  `json:"task_id"`
	RecordID   int64  `json:"record_id"`
	WorkflowID string `json:"workflow_id"`
	CommandID  string `json:"command_id"`
	AcquiredAt int64  `json:"acquired_at"`
}

// Acquire creates a per-client task lease.
func Acquire(ctx context.Context, info LockInfo) (bool, *LockInfo, error) {
	info.ClientIP = strings.TrimSpace(info.ClientIP)
	info.MachineID = strings.TrimSpace(info.MachineID)
	info.NodeID = strings.TrimSpace(info.NodeID)
	info.WorkflowID = strings.TrimSpace(info.WorkflowID)
	info.CommandID = strings.TrimSpace(info.CommandID)
	if lockIdentity(info) == "" || info.CommandID == "" {
		return false, nil, nil
	}
	if info.AcquiredAt <= 0 {
		info.AcquiredAt = time.Now().UnixMilli()
	}

	body, err := json.Marshal(info)
	if err != nil {
		return false, nil, err
	}
	result, err := g.Redis().Do(ctx, "SET", clientLockKey(lockIdentity(info)), string(body), "NX", "EX", int(LeaseTTL.Seconds()))
	if err != nil {
		return false, nil, err
	}
	if strings.EqualFold(strings.TrimSpace(result.String()), "OK") {
		return true, &info, nil
	}

	current, ok, err := Get(ctx, lockIdentity(info))
	if err != nil {
		return false, nil, err
	}
	if !ok {
		return false, nil, nil
	}
	return false, &current, nil
}

// Renew extends a lease only when the command still owns it.
func Renew(ctx context.Context, clientIP string, commandID string) (bool, error) {
	commandID = strings.TrimSpace(commandID)
	if commandID == "" {
		return false, nil
	}
	result, err := g.Redis().Do(ctx, "EVAL", `
local value = redis.call("GET", KEYS[1])
if not value then return 0 end
local ok, data = pcall(cjson.decode, value)
if ok and data["command_id"] == ARGV[1] then
	return redis.call("EXPIRE", KEYS[1], ARGV[2])
end
return 0
`, 1, clientLockKey(clientIP), commandID, int(LeaseTTL.Seconds()))
	if err != nil {
		return false, err
	}
	return result.Int() > 0, nil
}

// RenewNode extends a lease for node identity, falling back to client ip.
func RenewNode(ctx context.Context, nodeID string, clientIP string, commandID string) (bool, error) {
	return Renew(ctx, firstNonEmpty(nodeID, clientIP), commandID)
}

// Release deletes a lease only when the command still owns it.
func Release(ctx context.Context, clientIP string, commandID string) error {
	commandID = strings.TrimSpace(commandID)
	if commandID == "" {
		return nil
	}
	_, err := g.Redis().Do(ctx, "EVAL", `
local value = redis.call("GET", KEYS[1])
if not value then return 0 end
local ok, data = pcall(cjson.decode, value)
if ok and data["command_id"] == ARGV[1] then
	return redis.call("DEL", KEYS[1])
end
return 0
`, 1, clientLockKey(clientIP), commandID)
	return err
}

// ReleaseNode deletes a lease for node identity, falling back to client ip.
func ReleaseNode(ctx context.Context, nodeID string, clientIP string, commandID string) error {
	return Release(ctx, firstNonEmpty(nodeID, clientIP), commandID)
}

// Get returns the active client lease.
func Get(ctx context.Context, clientIP string) (LockInfo, bool, error) {
	result, err := g.Redis().Do(ctx, "GET", clientLockKey(clientIP))
	if err != nil {
		return LockInfo{}, false, err
	}
	text := strings.TrimSpace(result.String())
	if text == "" {
		return LockInfo{}, false, nil
	}

	var info LockInfo
	if err = json.Unmarshal([]byte(text), &info); err != nil {
		return LockInfo{}, false, err
	}
	return info, true, nil
}

// GetNode returns an active node lease, falling back to client ip.
func GetNode(ctx context.Context, nodeID string, clientIP string) (LockInfo, bool, error) {
	return Get(ctx, firstNonEmpty(nodeID, clientIP))
}

// List scans active client execution leases.
func List(ctx context.Context) ([]LockInfo, error) {
	var (
		cursor = "0"
		locks  []LockInfo
	)

	for {
		result, err := g.Redis().Do(ctx, "SCAN", cursor, "MATCH", KeyPrefix+"*", "COUNT", 200)
		if err != nil {
			return nil, err
		}

		values := result.Array()
		if len(values) < 2 {
			return locks, nil
		}

		cursor = gconv.String(values[0])
		keys := gconv.Strings(values[1])
		for _, key := range keys {
			value, err := g.Redis().Do(ctx, "GET", key)
			if err != nil {
				return nil, err
			}
			text := strings.TrimSpace(value.String())
			if text == "" {
				continue
			}

			var info LockInfo
			if err = json.Unmarshal([]byte(text), &info); err != nil {
				return nil, err
			}
			identity := strings.TrimPrefix(key, KeyPrefix)
			if info.NodeID == "" && strings.HasPrefix(identity, "node") {
				info.NodeID = identity
			}
			if info.ClientIP == "" {
				info.ClientIP = identity
			}
			locks = append(locks, info)
		}

		if cursor == "0" {
			return locks, nil
		}
	}
}

// RecordIDFromCommand extracts task record id from a task-record command id.
func RecordIDFromCommand(commandID string) int64 {
	return gconv.Int64(strings.TrimPrefix(strings.TrimSpace(commandID), "task-record-"))
}

func clientLockKey(clientIP string) string {
	return KeyPrefix + strings.TrimSpace(clientIP)
}

func lockIdentity(info LockInfo) string {
	return firstNonEmpty(info.NodeID, info.ClientIP)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			return value
		}
	}
	return ""
}
