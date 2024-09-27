/**
 * @author ysj
 * @email 2239831438@qq.com
 * @date 2024-09-27 00:20:27
 */

package redisHashRing

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/YShiJia/consistentHash"
	"github.com/demdxx/gocast"
	"github.com/xiaoxuxiansheng/redis_lock"
)

type RedisHashRing struct {
	//哈希环版本，本地保存一份
	version     int64
	key         string
	redisClient *Client
}

func NewRedisHashRing(key string, redisClient *Client) *RedisHashRing {
	return &RedisHashRing{
		key:         key,
		redisClient: redisClient,
	}
}

// 锁key
func (r *RedisHashRing) getLockKey() string {
	return fmt.Sprintf("redis:consistent_hash:ring:lock:%s", r.key)
}

// zset name
func (r *RedisHashRing) getRingKey() string {
	return fmt.Sprintf("redis:consistent_hash:ring:%s", r.key)
}

func (r *RedisHashRing) getTableVersionKey() string {
	return fmt.Sprintf("redis:consistent_hash:ring:version:%s", r.key)
}

func (r *RedisHashRing) getNodeReplicaKey() string {
	return fmt.Sprintf("redis:consistent_hash:ring:node:replica:%s", r.key)
}

func (r *RedisHashRing) Lock(ctx context.Context, expireSecond int64) error {
	lock := redis_lock.NewRedisLock(r.getLockKey(), r.redisClient, redis_lock.WithExpireSeconds(expireSecond))
	return lock.Lock(ctx)
}

func (r *RedisHashRing) Unlock(ctx context.Context) error {
	lock := redis_lock.NewRedisLock(r.getLockKey(), r.redisClient)
	return lock.Unlock(ctx)
}

func (r *RedisHashRing) AddVirtualNode(ctx context.Context, score int64, nodeID string) (version int64, err error) {
	hashScore, err := r.GetVirtualNode(ctx, score)
	if err != nil {
		//数据不存在
		if errors.Is(err, csHash.ErrVirtualNodeNotExists) {
			hashScore = &csHash.HashScore{
				Score:        score,
				VirtualNodes: make([]csHash.VirtualNode, 0),
			}
		} else {
			return 0, err
		}
	}
	//如果数据已经存在，直接返回
	for _, virtualNode := range hashScore.VirtualNodes {
		if virtualNode.VirtualNodeID == nodeID {
			return virtualNode.Version, nil
		}
	}

	if err = r.redisClient.ZRem(ctx, r.getRingKey(), score); err != nil {
		return 0, fmt.Errorf("redis ring zrem failed, err: %w", err)
	}

	hashRingVersion, err := r.GetVersion(ctx)
	if err != nil {
		return 0, err
	}
	r.version = max[int64](gocast.ToInt64(hashRingVersion), r.version)

	//TODO 后面想个办法解决一下数据溢出的问题，可以考虑使用英文进制，让字符串作为版本号
	hashScore.VirtualNodes = append(hashScore.VirtualNodes, csHash.VirtualNode{
		VirtualNodeID: nodeID,
		Version:       r.version + 1,
	})

	hsnData, _ := json.Marshal(hashScore)
	if err := r.redisClient.ZAdd(ctx, r.getRingKey(), score, string(hsnData)); err != nil {
		return 0, fmt.Errorf("redis ring zadd failed, err: %w", err)
	}
	//本地记录版本+1
	r.version++
	//更新redis 版本
	r.SetVersion(ctx, r.version)

	return r.version, nil
}

func (r *RedisHashRing) RemoveVirtualNode(ctx context.Context, score int64, nodeID string) error {
	hashScore, err := r.GetVirtualNode(ctx, score)
	if err != nil {
		//数据不存在
		if errors.Is(err, csHash.ErrVirtualNodeNotExists) {
			return nil
		} else {
			//真实报错
			return err
		}
	}
	index := 0
	for ; index < len(hashScore.VirtualNodes) && hashScore.VirtualNodes[index].VirtualNodeID != nodeID; index++ {
	}
	if index == len(hashScore.VirtualNodes) {
		return nil
	}

	//TODO 后续优化一下删除流程，不能让数据有删除了，但是没有上传的情况出现
	if err = r.redisClient.ZRem(ctx, r.getRingKey(), score); err != nil {
		return fmt.Errorf("redis ring zrem failed, err: %w", err)
	}

	//只有一个节点，那就是需要删除的节点，后面无需更新数据
	if len(hashScore.VirtualNodes) == 1 {
		return nil
	}
	hashScore.VirtualNodes = append(hashScore.VirtualNodes[:index], hashScore.VirtualNodes[index+1:]...)

	hsnData, _ := json.Marshal(hashScore)
	return r.redisClient.ZAdd(ctx, r.getRingKey(), score, string(hsnData))
}

func (r *RedisHashRing) GetVirtualNode(ctx context.Context, score int64) (hashScore *csHash.HashScore, err error) {
	scoreEntities, err := r.redisClient.ZRangeByScore(ctx, r.getRingKey(), score, score)
	if err != nil {
		return nil, err
	}
	if len(scoreEntities) != 1 {
		//不存在数据，直接返回
		if len(scoreEntities) == 0 {
			return nil, csHash.ErrVirtualNodeNotExists
		}
		return nil, fmt.Errorf("invalid entity len: %d", len(scoreEntities))
	}

	hs := csHash.HashScore{}
	if err = json.Unmarshal([]byte(scoreEntities[0].Val), &hs); err != nil {
		return nil, err
	}
	return &hs, nil
}

func (r *RedisHashRing) FindDataToVirtualNode(ctx context.Context, dataScore int64) (virtualNodeID string, err error) {
	scoreEntity, err := r.redisClient.Ceiling(ctx, r.getRingKey(), dataScore)
	//发生错误
	if err != nil && !errors.Is(err, ErrScoreNotExist) {
		return "", err
	}
	hashScore := csHash.HashScore{}
	//找到了数据
	if scoreEntity != nil {

		if err := json.Unmarshal([]byte(scoreEntity.Val), &hashScore); err != nil {
			return "", err
		}
		return hashScore.VirtualNodes[0].VirtualNodeID, nil
	}
	//寻找第一个score节点数据
	scoreEntity, err = r.redisClient.FirstOrLast(ctx, r.getRingKey(), true)
	if err != nil {
		if errors.Is(err, ErrScoreNotExist) {
			//节点不存在
			return "", csHash.ErrVirtualNodeNotExists
		} else {
			return "", err
		}
	}
	if err := json.Unmarshal([]byte(scoreEntity.Val), &hashScore); err != nil {
		return "", err
	}
	return hashScore.VirtualNodes[0].VirtualNodeID, nil
}

func (r *RedisHashRing) AddRealNode(ctx context.Context, nodeName string, replicas int64) (err error) {
	if err = r.redisClient.HSet(ctx, r.getNodeReplicaKey(), nodeName, gocast.ToString(replicas)); err != nil {
		return fmt.Errorf("redis ring add node to replica failed, err: %w", err)
	}
	return nil
}

func (r *RedisHashRing) GetRealNodes(ctx context.Context) (nodes map[string]int64, err error) {
	res, err := r.redisClient.HGetAll(ctx, r.getNodeReplicaKey())
	if err != nil {
		return nil, fmt.Errorf("redis ring nodes hgetall failed, err: %w", err)
	}
	nodes = make(map[string]int64, len(res))
	for k, v := range res {
		nodes[k] = gocast.ToInt64(v)
	}
	return nodes, err
}

func (r *RedisHashRing) GetRealNode(ctx context.Context, nodeName string) (replicas int64, err error) {
	nodes, err := r.GetRealNodes(ctx)
	if err != nil {
		return 0, err
	}
	return nodes[nodeName], nil
}

func (r *RedisHashRing) RemoveRealNode(ctx context.Context, nodeName string) (err error) {
	if err = r.redisClient.HDel(ctx, r.getNodeReplicaKey(), nodeName); err != nil {
		return fmt.Errorf("redis ring remove node from replica failed, err: %w", err)
	}
	return nil
}

func (r *RedisHashRing) GetVersion(ctx context.Context) (version int64, err error) {
	versionStr, err := r.redisClient.Get(ctx, r.getTableVersionKey())
	if err != nil {
		return 0, err
	}
	return gocast.ToInt64(versionStr), nil
}

func (r *RedisHashRing) SetVersion(ctx context.Context, version int64) (err error) {
	return r.redisClient.Set(ctx, r.getTableVersionKey(), fmt.Sprintf("%v", r.version))
}
