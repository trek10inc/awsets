package cache

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/trek10inc/awsets"
	"github.com/trek10inc/awsets/resource"
	"go.etcd.io/bbolt"
)

type BoltCache struct {
	db      *bbolt.DB
	account string
	refresh bool
}

func NewBoltCache(account string, refresh bool) (*BoltCache, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w\n", err)
	}
	db, err := bbolt.Open(filepath.Join(home, ".awsets_cache"), 0666, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open cache: %w\n", err)
	}
	err = db.Update(func(tx *bbolt.Tx) error {
		_, err = tx.CreateBucketIfNotExists([]byte(account))
		if err != nil {
			return fmt.Errorf("failed to create or get bucket: %w\n", err)
		}
		return nil
	})
	return &BoltCache{
		db:      db,
		account: account,
		refresh: refresh,
	}, err
}

func (c *BoltCache) IsCached(accountId, region string, kind awspelunk.ListerName) bool {
	if c.refresh {
		return false
	}
	err := c.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(accountId))
		res := bucket.Get([]byte(fmt.Sprintf("%s_%s", region, kind)))
		if res == nil {
			return errors.New("no cache")
		}
		t, err := time.Parse(time.RFC3339, string(res))
		if err != nil {
			return fmt.Errorf("failed to parse time %s: %w\n", string(res), err)
		}
		if time.Now().Sub(t) > 24*time.Hour {
			return errors.New("cache has expired")
		}
		return nil
	})
	if err != nil {
		//fmt.Printf("failed: %v\n", err)
		return false
	}
	return true
}

func (c *BoltCache) SaveGroup(rg *resource.Group, region string, kind awspelunk.ListerName) error {
	data, err := rg.JSON()
	if err != nil {
		return err
	}

	err = c.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(c.account))
		return b.Put([]byte(fmt.Sprintf("%s_%s_data", region, kind)), []byte(data))
	})
	if err != nil {
		return fmt.Errorf("failed to save data: %w", err)
	}

	return c.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(c.account))
		return b.Put([]byte(fmt.Sprintf("%s_%s", region, kind)), []byte(time.Now().Format(time.RFC3339)))
	})
}

func (c *BoltCache) LoadGroup(region string, kind awspelunk.ListerName) (*resource.Group, error) {
	rg := resource.NewGroup()
	err := c.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(c.account))
		data := b.Get([]byte(fmt.Sprintf("%s_%s_data", region, kind)))
		var resources []resource.Resource
		err := json.Unmarshal(data, &resources)
		if err != nil {
			return fmt.Errorf("failed to unmarshall value of size %d: %w\n", len(data), err)
		}
		for i := range resources {
			rg.AddResource(resources[i])
		}
		return nil
	})
	return rg, err
}
