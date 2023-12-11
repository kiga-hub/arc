package leadelection

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kiga-hub/common/configuration"
)

// NacosLock is a lock using nacos as backend
type NacosLock struct {
	client   *configuration.NacosClient
	group    string
	dataID   string
	identity string
}

// NewNacosLock return a new NacosLock
func NewNacosLock(group, dataID, identity string, client *configuration.NacosClient) *NacosLock {
	return &NacosLock{
		client:   client,
		group:    group,
		dataID:   dataID,
		identity: identity,
	}
}

// Get returns the LeaderElectionRecord
func (lock NacosLock) Get(ctx context.Context) (*LeaderElectionRecord, []byte, error) {
	var record LeaderElectionRecord
	var err error
	recordStr, err := lock.client.Get(lock.dataID, lock.group)
	if err != nil {
		return nil, nil, err
	}
	if recordStr == "" {
		return nil, nil, ErrNotFound
	}
	recordBytes := []byte(recordStr)
	if err := json.Unmarshal(recordBytes, &record); err != nil {
		return nil, nil, err
	}
	return &record, recordBytes, nil
}

// Create attempts to create a LeaderElectionRecord
func (lock NacosLock) Create(ctx context.Context, ler LeaderElectionRecord) error {
	recordBytes, err := json.Marshal(ler)
	if err != nil {
		return err
	}
	_, err = lock.client.Publish(lock.dataID, lock.group, string(recordBytes))
	return err
}

// Update will update and existing LeaderElectionRecord
func (lock NacosLock) Update(ctx context.Context, ler LeaderElectionRecord) error {
	return lock.Create(ctx, ler)
}

// RecordEvent is used to record events
func (lock NacosLock) RecordEvent(string) {}

// Identity will return the locks Identity
func (lock NacosLock) Identity() string {
	return lock.identity
}

// Describe is used to convert details on current resource lock into a string
func (lock NacosLock) Describe() string {
	return fmt.Sprintf("%v/%v", lock.group, lock.dataID)
}
