/**
 * @author ysj
 * @email 2239831438@qq.com
 * @date 2024-09-26 19:06:13
 */

package csHash

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

type ConsistentHash struct {
	hashRing  HashRing
	encryptor HashEncryptor
	migrator  Migrator
	logger    Logger
	opts      ConsistentHashOptions
}

func NewConsistentHash(
	hashRing HashRing,
	encryptor HashEncryptor,
	migrator Migrator,
	logger Logger,
	opts ...ConsistentHashOption) *ConsistentHash {

	ch := ConsistentHash{
		hashRing:  hashRing,
		encryptor: encryptor,
		migrator:  migrator,
		logger:    logger,
	}

	for _, opt := range opts {
		opt(&ch.opts)
	}

	ch.opts.repair()
	return &ch
}

func repairWeight(weight int64) int64 {
	switch {
	case weight <= 0:
		weight = 1
	case weight > 10:
		weight = 10
	}
	return weight
}

func getVirtualNodeID(nodeName string, index int64) string {
	return fmt.Sprintf("%s_%d", nodeName, index)
}

func parseVirtualNodeID(virtualNodeID string) (string, int64, error) {
	index := strings.LastIndex(virtualNodeID, "_")
	if index == -1 {
		return "", int64(0), ErrInvalidVirtualNodeID
	}
	seg, err := strconv.Atoi(virtualNodeID[index+1:])
	if err != nil {
		return "", int64(0), ErrInvalidVirtualNodeID
	}

	return virtualNodeID[:index], int64(seg), nil
}

// 添加节点
func (c *ConsistentHash) AddNode(ctx context.Context, nodeName string, weight int64) error {
	if err := c.hashRing.Lock(ctx, c.opts.lockExpireSeconds); err != nil {
		return err
	}

	defer func() {
		_ = c.hashRing.Unlock(ctx)
	}()

	// 先判断RealNode中是否有该节点
	replicas, err := c.hashRing.GetRealNode(ctx, nodeName)
	if err != nil {
		return err
	}
	if replicas > 0 {
		return ErrNodeAlreadyExists
	}

	//需要映射的节点数量
	nodeReplicas := repairWeight(weight) * c.opts.replicas
	err = c.hashRing.AddRealNode(ctx, nodeName, nodeReplicas)
	if err != nil {
		return err
	}

	var migrateTasks []func()
	//将虚拟节点插入到hash环中
	for i := int64(1); i <= nodeReplicas; i++ {
		virtualNodeID := getVirtualNodeID(nodeName, i)
		virtualNodeScore := c.encryptor.Encrypt(virtualNodeID)
		if err := c.hashRing.AddVirtualNode(ctx, int64(virtualNodeScore), virtualNodeID); err != nil {
			return err
		}
		// 获取迁移参数
		data, fromRealNode, toRealNode, err := c.migrateIn(ctx, int64(virtualNodeScore), nodeName)
		if err != nil {
			return err
		}
		migrateTasks = append(migrateTasks, func() {
			_ = c.migrator(ctx, data, fromRealNode, toRealNode)
		})
	}
	c.batchExecuteMigrator(migrateTasks)

	return nil
}

// 删除节点
func (c *ConsistentHash) RemoveNode(ctx context.Context, nodeName string) error {
	if err := c.hashRing.Lock(ctx, c.opts.lockExpireSeconds); err != nil {
		return err
	}

	defer func() {
		_ = c.hashRing.Unlock(ctx)
	}()

	// 获取nodeName信息
	replicas, err := c.hashRing.GetRealNode(ctx, nodeName)
	if err != nil {
		return err
	}
	if replicas == 0 {
		// 没有节点，直接返回
		return nil
	}
	//先删除ReadNode的信息
	err = c.hashRing.RemoveRealNode(ctx, nodeName)
	if err != nil {
		return err
	}

	var migrateTasks []func()
	// 挨个删除
	for i := int64(1); i <= replicas; i++ {
		virtualNodeID := getVirtualNodeID(nodeName, i)
		virtualNodeScore := c.encryptor.Encrypt(virtualNodeID)

		data, fromRealNode, toRealNode, err := c.migrateOut(ctx, int64(virtualNodeScore), nodeName)
		if err != nil {
			return err
		}

		if err := c.hashRing.RemoveVirtualNode(ctx, int64(virtualNodeScore), virtualNodeID); err != nil {
			return err
		}

		if len(data) == 0 {
			continue
		}
		migrateTasks = append(migrateTasks, func() {
			_ = c.migrator(ctx, data, fromRealNode, toRealNode)
		})
	}
	c.batchExecuteMigrator(migrateTasks)

	return nil
}

// 获取节点
func (c *ConsistentHash) GetNode(ctx context.Context, dataKey string) (nodeName string, err error) {
	if err := c.hashRing.Lock(ctx, c.opts.lockExpireSeconds); err != nil {
		return "", err
	}

	defer func() {
		_ = c.hashRing.Unlock(ctx)
	}()

	dataScore := c.encryptor.Encrypt(dataKey)
	virtualNode, err := c.hashRing.FindDataToVirtualNode(ctx, int64(dataScore))
	if err != nil {
		return "", err
	}
	RealNode, _, err := parseVirtualNodeID(virtualNode)
	if err != nil {
		return "", err
	}
	return RealNode, nil
}

func (c *ConsistentHash) AddDataKeyToNode(ctx context.Context, dataKey string) (nodeName string, err error) {
	RealNode, err := c.GetNode(ctx, dataKey)
	if err != nil {
		return "", err
	}
	if err := c.hashRing.Lock(ctx, c.opts.lockExpireSeconds); err != nil {
		return "", err
	}
	defer func() {
		_ = c.hashRing.Unlock(ctx)
	}()

	if err = c.hashRing.AddDataToRealNode(ctx, RealNode, map[string]struct{}{dataKey: {}}); err != nil {
		return "", err
	}
	return RealNode, nil
}
