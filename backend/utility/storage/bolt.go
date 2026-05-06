package storage

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Zany2/browserflow/backend/internal/model"
	bolt "go.etcd.io/bbolt"
)

var (
	llmConfigsBucket       = []byte("llm_configs")               // llmConfigsBucket stores model configs 存储大模型配置
	chatSessionsBucket     = []byte("chat_sessions")             // chatSessionsBucket stores chat sessions 存储对话会话
	browserBucket          = []byte("browser_instances")         // browserBucket stores browser configs 存储浏览器配置
	automaSnapshotBucket   = []byte("automa_workflow_snapshots") // automaSnapshotBucket stores latest workflow snapshots 存储工作流快照
	automaWorkflowBucket   = []byte("automa_workflow_records")   // automaWorkflowBucket stores workflow records 存储工作流记录
	automaWorkflowSeqKey   = []byte("_seq")                      // automaWorkflowSeqKey stores workflow sequence 存储工作流自增序列
	automaWorkflowIndexKey = []byte("_idx_automa_id:")           // automaWorkflowIndexKey prefixes Automa id index Automa ID 索引前缀
)

// BoltDB local file database 本地文件数据库
type BoltDB struct {
	db *bolt.DB
}

// NewBoltDB opens local file database 打开并初始化本地文件数据库
func NewBoltDB(dbPath string) (*BoltDB, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return nil, err
	}

	db, err := bolt.Open(dbPath, 0o600, &bolt.Options{Timeout: 5 * time.Second})
	if err != nil {
		return nil, fmt.Errorf("open database failed: %w", err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		for _, bucket := range [][]byte{llmConfigsBucket, chatSessionsBucket, browserBucket, automaSnapshotBucket, automaWorkflowBucket} {
			if _, err := tx.CreateBucketIfNotExists(bucket); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		_ = db.Close()
		return nil, err
	}

	return &BoltDB{db: db}, nil
}

// Close closes database 关闭数据库
func (b *BoltDB) Close() error {
	return b.db.Close()
}

// SaveAutomaWorkflowSnapshot saves latest workflow snapshot 保存最新工作流快照
func (b *BoltDB) SaveAutomaWorkflowSnapshot(snapshot *model.AutomaWorkflowSnapshot) error {
	if snapshot.ID == "" {
		snapshot.ID = "latest"
	}
	snapshot.UpdatedAt = time.Now()

	return b.db.Update(func(tx *bolt.Tx) error {
		data, err := json.Marshal(snapshot)
		if err != nil {
			return err
		}
		return tx.Bucket(automaSnapshotBucket).Put([]byte(snapshot.ID), data)
	})
}

// GetAutomaWorkflowSnapshot gets workflow snapshot 获取工作流快照
func (b *BoltDB) GetAutomaWorkflowSnapshot(id string) (*model.AutomaWorkflowSnapshot, error) {
	if id == "" {
		id = "latest"
	}

	var snapshot model.AutomaWorkflowSnapshot
	err := b.db.View(func(tx *bolt.Tx) error {
		data := tx.Bucket(automaSnapshotBucket).Get([]byte(id))
		if data == nil {
			return errors.New("automa workflow snapshot not found")
		}
		return json.Unmarshal(data, &snapshot)
	})
	if err != nil {
		return nil, err
	}
	return &snapshot, nil
}

// SaveLLMConfig saves large model config 保存大模型配置
func (b *BoltDB) SaveLLMConfig(config *model.LLMConfig) error {
	now := time.Now()
	if config.CreatedAt.IsZero() {
		config.CreatedAt = now
	}
	config.UpdatedAt = now

	if config.IsDefault {
		if err := b.ClearDefaultLLMConfig(); err != nil {
			return err
		}
	}

	return b.db.Update(func(tx *bolt.Tx) error {
		data, err := json.Marshal(config)
		if err != nil {
			return err
		}
		return tx.Bucket(llmConfigsBucket).Put([]byte(config.ID), data)
	})
}

// GetLLMConfig gets large model config 获取大模型配置
func (b *BoltDB) GetLLMConfig(id string) (*model.LLMConfig, error) {
	var config model.LLMConfig
	err := b.db.View(func(tx *bolt.Tx) error {
		data := tx.Bucket(llmConfigsBucket).Get([]byte(id))
		if data == nil {
			return errors.New("llm config not found")
		}
		return json.Unmarshal(data, &config)
	})
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// ListLLMConfigs lists large model configs 列出大模型配置
func (b *BoltDB) ListLLMConfigs() ([]*model.LLMConfig, error) {
	configs := make([]*model.LLMConfig, 0)
	err := b.db.View(func(tx *bolt.Tx) error {
		return tx.Bucket(llmConfigsBucket).ForEach(func(_, v []byte) error {
			var config model.LLMConfig
			if err := json.Unmarshal(v, &config); err != nil {
				return err
			}
			configs = append(configs, &config)
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(configs, func(i, j int) bool { return configs[i].CreatedAt.After(configs[j].CreatedAt) })
	return configs, nil
}

// DeleteLLMConfig deletes large model config 删除大模型配置
func (b *BoltDB) DeleteLLMConfig(id string) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(llmConfigsBucket).Delete([]byte(id))
	})
}

// ClearDefaultLLMConfig clears default large model config 清空默认大模型配置
func (b *BoltDB) ClearDefaultLLMConfig() error {
	return b.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(llmConfigsBucket)
		return bucket.ForEach(func(k, v []byte) error {
			var config model.LLMConfig
			if err := json.Unmarshal(v, &config); err != nil {
				return err
			}
			if !config.IsDefault {
				return nil
			}
			config.IsDefault = false
			config.UpdatedAt = time.Now()
			data, err := json.Marshal(&config)
			if err != nil {
				return err
			}
			return bucket.Put(k, data)
		})
	})
}

// GetDefaultLLMConfig gets default or first active large model config 获取默认或首个启用配置
func (b *BoltDB) GetDefaultLLMConfig() (*model.LLMConfig, error) {
	configs, err := b.ListLLMConfigs()
	if err != nil {
		return nil, err
	}
	var firstActive *model.LLMConfig
	for _, config := range configs {
		if !config.IsActive {
			continue
		}
		if config.IsDefault {
			return config, nil
		}
		if firstActive == nil {
			firstActive = config
		}
	}
	if firstActive == nil {
		return nil, errors.New("no active llm config")
	}
	return firstActive, nil
}

// SaveChatSession saves chat session 保存对话会话
func (b *BoltDB) SaveChatSession(session *model.ChatSession) error {
	now := time.Now()
	if session.CreatedAt.IsZero() {
		session.CreatedAt = now
	}
	session.UpdatedAt = now

	return b.db.Update(func(tx *bolt.Tx) error {
		data, err := json.Marshal(session)
		if err != nil {
			return err
		}
		return tx.Bucket(chatSessionsBucket).Put([]byte(session.ID), data)
	})
}

// GetChatSession gets chat session 获取对话会话
func (b *BoltDB) GetChatSession(id string) (*model.ChatSession, error) {
	var session model.ChatSession
	err := b.db.View(func(tx *bolt.Tx) error {
		data := tx.Bucket(chatSessionsBucket).Get([]byte(id))
		if data == nil {
			return errors.New("chat session not found")
		}
		return json.Unmarshal(data, &session)
	})
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// ListChatSessions lists chat sessions 列出对话会话
func (b *BoltDB) ListChatSessions() ([]*model.ChatSession, error) {
	sessions := make([]*model.ChatSession, 0)
	err := b.db.View(func(tx *bolt.Tx) error {
		return tx.Bucket(chatSessionsBucket).ForEach(func(_, v []byte) error {
			var session model.ChatSession
			if err := json.Unmarshal(v, &session); err != nil {
				return err
			}
			sessions = append(sessions, &session)
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(sessions, func(i, j int) bool { return sessions[i].UpdatedAt.After(sessions[j].UpdatedAt) })
	return sessions, nil
}

// DeleteChatSession deletes chat session 删除对话会话
func (b *BoltDB) DeleteChatSession(id string) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(chatSessionsBucket).Delete([]byte(id))
	})
}

// SaveBrowserInstance saves browser config 保存浏览器配置
func (b *BoltDB) SaveBrowserInstance(instance *model.BrowserInstance) error {
	now := time.Now()
	if instance.CreatedAt.IsZero() {
		instance.CreatedAt = now
	}
	instance.UpdatedAt = now

	if instance.IsDefault {
		if err := b.ClearDefaultBrowserInstance(); err != nil {
			return err
		}
	}

	return b.db.Update(func(tx *bolt.Tx) error {
		data, err := json.Marshal(instance)
		if err != nil {
			return err
		}
		return tx.Bucket(browserBucket).Put([]byte(instance.ID), data)
	})
}

// GetBrowserInstance gets browser config 获取浏览器配置
func (b *BoltDB) GetBrowserInstance(id string) (*model.BrowserInstance, error) {
	var instance model.BrowserInstance
	err := b.db.View(func(tx *bolt.Tx) error {
		data := tx.Bucket(browserBucket).Get([]byte(id))
		if data == nil {
			return errors.New("browser instance not found")
		}
		return json.Unmarshal(data, &instance)
	})
	if err != nil {
		return nil, err
	}
	return &instance, nil
}

// ListBrowserInstances lists browser configs 列出浏览器配置
func (b *BoltDB) ListBrowserInstances() ([]*model.BrowserInstance, error) {
	instances := make([]*model.BrowserInstance, 0)
	err := b.db.View(func(tx *bolt.Tx) error {
		return tx.Bucket(browserBucket).ForEach(func(_, v []byte) error {
			var instance model.BrowserInstance
			if err := json.Unmarshal(v, &instance); err != nil {
				return err
			}
			instances = append(instances, &instance)
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(instances, func(i, j int) bool { return instances[i].CreatedAt.After(instances[j].CreatedAt) })
	return instances, nil
}

// DeleteBrowserInstance deletes browser config 删除浏览器配置
func (b *BoltDB) DeleteBrowserInstance(id string) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(browserBucket).Delete([]byte(id))
	})
}

// GetDefaultBrowserInstance gets default browser config 获取默认浏览器配置
func (b *BoltDB) GetDefaultBrowserInstance() (*model.BrowserInstance, error) {
	instances, err := b.ListBrowserInstances()
	if err != nil {
		return nil, err
	}
	var first *model.BrowserInstance
	for _, instance := range instances {
		if instance.IsDefault {
			return instance, nil
		}
		if first == nil {
			first = instance
		}
	}
	if first == nil {
		return nil, errors.New("browser instance not found")
	}
	return first, nil
}

// ClearDefaultBrowserInstance clears default browser config 清空默认浏览器配置
func (b *BoltDB) ClearDefaultBrowserInstance() error {
	return b.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(browserBucket)
		return bucket.ForEach(func(k, v []byte) error {
			var instance model.BrowserInstance
			if err := json.Unmarshal(v, &instance); err != nil {
				return err
			}
			if !instance.IsDefault {
				return nil
			}
			instance.IsDefault = false
			instance.UpdatedAt = time.Now()
			data, err := json.Marshal(&instance)
			if err != nil {
				return err
			}
			return bucket.Put(k, data)
		})
	})
}

// SaveAutomaWorkflowRecord saves workflow record 保存工作流记录
func (b *BoltDB) SaveAutomaWorkflowRecord(record *model.AutomaWorkflowRecord) error {
	now := time.Now()
	return b.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(automaWorkflowBucket)
		if record.ID <= 0 {
			nextID, err := nextAutomaWorkflowID(bucket)
			if err != nil {
				return err
			}
			record.ID = nextID
			record.CreatedAt = now
		}
		if record.CreatedAt.IsZero() {
			record.CreatedAt = now
		}
		record.UpdatedAt = now
		data, err := json.Marshal(record)
		if err != nil {
			return err
		}
		if err := bucket.Put(int64Key(record.ID), data); err != nil {
			return err
		}
		if strings.TrimSpace(record.AutomaID) != "" {
			return bucket.Put(indexKey(record.AutomaID), int64Key(record.ID))
		}
		return nil
	})
}

// GetAutomaWorkflowRecord gets workflow by local id or Automa id 获取工作流记录
func (b *BoltDB) GetAutomaWorkflowRecord(id string) (*model.AutomaWorkflowRecord, error) {
	var record model.AutomaWorkflowRecord
	err := b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(automaWorkflowBucket)
		key := []byte(strings.TrimSpace(id))
		if len(key) != 8 {
			if parsed, ok := parsePositiveInt64(id); ok {
				key = int64Key(parsed)
			}
		}
		if data := bucket.Get(key); data != nil {
			return json.Unmarshal(data, &record)
		}
		indexData := bucket.Get(indexKey(id))
		if indexData == nil {
			return errors.New("automa workflow not found")
		}
		data := bucket.Get(indexData)
		if data == nil {
			return errors.New("automa workflow not found")
		}
		return json.Unmarshal(data, &record)
	})
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// ListAutomaWorkflowRecords lists workflow records 列出工作流记录
func (b *BoltDB) ListAutomaWorkflowRecords() ([]*model.AutomaWorkflowRecord, error) {
	records := make([]*model.AutomaWorkflowRecord, 0)
	err := b.db.View(func(tx *bolt.Tx) error {
		return tx.Bucket(automaWorkflowBucket).ForEach(func(k, v []byte) error {
			if len(k) != 8 || strings.HasPrefix(string(k), "_") {
				return nil
			}
			var record model.AutomaWorkflowRecord
			if err := json.Unmarshal(v, &record); err != nil {
				return err
			}
			records = append(records, &record)
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(records, func(i, j int) bool { return records[i].UpdatedAt.After(records[j].UpdatedAt) })
	return records, nil
}

// DeleteAutomaWorkflowRecord deletes workflow record 删除工作流记录
func (b *BoltDB) DeleteAutomaWorkflowRecord(id string) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		record, err := getAutomaWorkflowRecordTx(tx, id)
		if err != nil {
			return nil
		}
		bucket := tx.Bucket(automaWorkflowBucket)
		if strings.TrimSpace(record.AutomaID) != "" {
			_ = bucket.Delete(indexKey(record.AutomaID))
		}
		return bucket.Delete(int64Key(record.ID))
	})
}

// getAutomaWorkflowRecordTx gets workflow record inside transaction 在事务中获取工作流记录
func getAutomaWorkflowRecordTx(tx *bolt.Tx, id string) (*model.AutomaWorkflowRecord, error) {
	bucket := tx.Bucket(automaWorkflowBucket)
	var key []byte
	if parsed, ok := parsePositiveInt64(id); ok {
		key = int64Key(parsed)
	} else {
		key = bucket.Get(indexKey(id))
	}
	if key == nil {
		return nil, errors.New("automa workflow not found")
	}
	data := bucket.Get(key)
	if data == nil {
		return nil, errors.New("automa workflow not found")
	}
	var record model.AutomaWorkflowRecord
	if err := json.Unmarshal(data, &record); err != nil {
		return nil, err
	}
	return &record, nil
}

// nextAutomaWorkflowID increments workflow id sequence 递增工作流 ID 序列
func nextAutomaWorkflowID(bucket *bolt.Bucket) (int64, error) {
	seq := int64(0)
	if data := bucket.Get(automaWorkflowSeqKey); len(data) == 8 {
		seq = int64(binary.BigEndian.Uint64(data))
	}
	seq++
	if seq <= 0 {
		seq = 1
	}
	return seq, bucket.Put(automaWorkflowSeqKey, int64Key(seq))
}

// int64Key encodes int64 as big-endian key 将 int64 编码为大端序键
func int64Key(value int64) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(value))
	return buf
}

// indexKey builds Automa id index key 构建 Automa ID 索引键
func indexKey(automaID string) []byte {
	return append(append([]byte{}, automaWorkflowIndexKey...), []byte(strings.TrimSpace(automaID))...)
}

// parsePositiveInt64 parses positive int64 解析正整数 ID
func parsePositiveInt64(value string) (int64, bool) {
	var parsed int64
	if _, err := fmt.Sscanf(strings.TrimSpace(value), "%d", &parsed); err != nil || parsed <= 0 {
		return 0, false
	}
	return parsed, true
}
