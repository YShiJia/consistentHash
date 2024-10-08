/**
 * @author ysj
 * @email 2239831438@qq.com
 * @date 2024-09-27 00:20:27
 */

package redisHashRing

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/demdxx/gocast"
	"github.com/xiaoxuxiansheng/redis_lock"
)

type RedisHashRing struct {
	key         string
	redisClient *Client
}

func NewRedisHashRing(key string, redisClient *Client) *RedisHashRing {
	return &RedisHashRing{
		key:         key,
		redisClient: redisClient,
	}
}

// lock key
func (r *RedisHashRing) getLockKey() string {
	return fmt.Sprintf("redis:consistent_hash:ring:lock:%s", r.key)
}

// zset name
func (r *RedisHashRing) getRingKey() string {
	return fmt.Sprintf("redis:consistent_hash:ring:%s", r.key)
}

// string name
func (r *RedisHashRing) getNodeReplicaKey() string {
	return fmt.Sprintf("redis:consistent_hash:ring:node:replica:%s", r.key)
}

// nodeData name
func (r *RedisHashRing) getNodeDataKey(nodeID string) string {
	return fmt.Sprintf("redis:consistent_hash:ring:node:data:%s", nodeID)
}

// expireSecond 获取锁最大尝试时间
func (r *RedisHashRing) Lock(ctx context.Context, expireSecond int64) error {
	// 创建一个锁: 阻塞+最大读取时间
	lock := redis_lock.NewRedisLock(r.getLockKey(), r.redisClient, redis_lock.WithBlock(), redis_lock.WithExpireSeconds(expireSecond))
	return lock.Lock(ctx)
}

func (r *RedisHashRing) Unlock(ctx context.Context) error {
	//取消锁
	lock := redis_lock.NewRedisLock(r.getLockKey(), r.redisClient)
	return lock.Unlock(ctx)
}

func (r *RedisHashRing) AddVirtualNode(ctx context.Context, score int64, virtualNode string) error {
	virtualNodes, err := r.GetVirtualNodes(ctx, score)
	if err != nil {
		return err
	}
	//如果数据已经存在，直接返回
	for _, vn := range virtualNodes {
		if vn == virtualNode {
			return nil
		}
	}
	// 先删除原来的数据
	if err = r.redisClient.ZRem(ctx, r.getRingKey(), score); err != nil {
		return fmt.Errorf("redis ring zrem failed, err: %w", err)
	}
	// 更新数据
	virtualNodes = append(virtualNodes, virtualNode)
	vnData, _ := json.Marshal(virtualNodes)
	// 上传数据
	if err := r.redisClient.ZAdd(ctx, r.getRingKey(), score, string(vnData)); err != nil {
		return fmt.Errorf("redis ring zadd failed, err: %w", err)
	}
	return nil
}

func (r *RedisHashRing) RemoveVirtualNode(ctx context.Context, score int64, nodeID string) error {
	virtualNodes, err := r.GetVirtualNodes(ctx, score)
	if err != nil {
		return err
	}
	// 更新数据
	for index := 1; index < len(virtualNodes); index++ {
		if virtualNodes[index] == nodeID {
			virtualNodes = append(virtualNodes[:index], virtualNodes[index+1:]...)
			break
		}
	}

	// 删除原来的数据
	if err = r.redisClient.ZRem(ctx, r.getRingKey(), score); err != nil {
		return fmt.Errorf("redis ring zrem failed, err: %w", err)
	}

	vnData, _ := json.Marshal(virtualNodes)
	return r.redisClient.ZAdd(ctx, r.getRingKey(), score, string(vnData))
}

func (r *RedisHashRing) GetVirtualNodes(ctx context.Context, score int64) ([]string, error) {
	scoreEntities, err := r.redisClient.ZRangeByScore(ctx, r.getRingKey(), score, score)
	if err != nil {
		return nil, err
	}
	if len(scoreEntities) != 1 {
		if len(scoreEntities) == 0 {
			//不存在数据，直接返回
			return []string{}, nil
		}
		return nil, fmt.Errorf("invalid entity len: %d", len(scoreEntities))
	}

	var virtualNodes []string
	if err = json.Unmarshal([]byte(scoreEntities[0].Val), &virtualNodes); err != nil {
		return nil, err
	}
	return virtualNodes, nil
}

func (r *RedisHashRing) FindDataToVirtualNode(ctx context.Context, dataScore int64) (virtualNode string, err error) {
	//由小到大查找
	scoreEntity, err := r.redisClient.Ceiling(ctx, r.getRingKey(), dataScore)
	//发生错误
	if err != nil {
		return "", err
	}
	var virtualNodes []string
	//找到了数据
	if scoreEntity != nil {
		if err := json.Unmarshal([]byte(scoreEntity.Val), &virtualNodes); err != nil {
			return "", err
		}
		return virtualNodes[0], nil
	}

	// dataScore 为最大节点，他的下一个节点为第一个节点
	scoreEntity, err = r.redisClient.FirstOrLast(ctx, r.getRingKey(), true)
	if err != nil {
		return "", err
	}
	if scoreEntity != nil {
		if err := json.Unmarshal([]byte(scoreEntity.Val), &virtualNodes); err != nil {
			return "", err
		}
		return virtualNodes[0], nil
	}
	return "", nil
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
	num, err := r.redisClient.HGet(ctx, r.getNodeReplicaKey(), nodeName)
	if err != nil {
		return 0, err
	}
	return gocast.ToInt64(num), nil
}

func (r *RedisHashRing) RemoveRealNode(ctx context.Context, nodeName string) (err error) {
	if err = r.redisClient.HDel(ctx, r.getNodeReplicaKey(), nodeName); err != nil {
		return fmt.Errorf("redis ring remove node from replica failed, err: %w", err)
	}
	return nil
}

func (r *RedisHashRing) AddDataToRealNode(ctx context.Context, RealNode string, data map[string]struct{}) error {
	tmp, err := r.GetDataFromRealNode(ctx, RealNode)
	if err != nil {
		return err
	}
	for k := range data {
		tmp[k] = struct{}{}
	}
	bytes, err := json.Marshal(tmp)
	return r.redisClient.Set(ctx, r.getNodeDataKey(RealNode), string(bytes))
}

func (r *RedisHashRing) GetDataFromRealNode(ctx context.Context, RealNode string) (data map[string]struct{}, err error) {
	str, err := r.redisClient.Get(ctx, r.getNodeDataKey(RealNode))
	tmp := make(map[string]struct{})
	if str != "" {
		if err := json.Unmarshal([]byte(str), &tmp); err != nil {
			return nil, err
		}
	}
	return tmp, nil
}

func (r *RedisHashRing) RemoveDataFromRealNode(ctx context.Context, RealNode string, data map[string]struct{}) error {
	tmp, err := r.GetDataFromRealNode(ctx, RealNode)
	if err != nil {
		return err
	}
	for k := range data {
		delete(tmp, k)
	}
	bytes, err := json.Marshal(tmp)
	return r.redisClient.Set(ctx, r.getNodeDataKey(RealNode), string(bytes))
}

// Floor 查找hashRing 的 score前面的第一个节点，不包括score
// 如果只有一个节点 == score， 返回当前score
func (r *RedisHashRing) Floor(ctx context.Context, score int64) (string, int64, error) {
	scoreEntity, err := r.redisClient.Floor(ctx, r.getRingKey(), score)
	if err != nil {
		return "", -1, err
	}
	if scoreEntity != nil {
		var virtualNodes []string
		if err := json.Unmarshal([]byte(scoreEntity.Val), &virtualNodes); err != nil {
			return "", -1, err
		}
		return virtualNodes[0], scoreEntity.Score, nil
	}
	//往前找没有，需要寻找最后一个节点
	scoreEntity, err = r.redisClient.FirstOrLast(ctx, r.getRingKey(), false)
	if err != nil {
		return "", -1, err
	}
	if scoreEntity == nil {
		return "", -1, nil
	}

	var virtualNodes []string
	if err := json.Unmarshal([]byte(scoreEntity.Val), &virtualNodes); err != nil {
		return "", -1, err
	}
	return virtualNodes[0], scoreEntity.Score, nil
}

// Ceiling 查找hashRing 的 score 后面的第一个节点，不包括score
// 如果只有一个节点 == score， 返回当前score
func (r *RedisHashRing) Ceiling(ctx context.Context, score int64) (string, int64, error) {
	scoreEntity, err := r.redisClient.Ceiling(ctx, r.getRingKey(), score)
	if err != nil {
		return "", -1, err
	}
	if scoreEntity != nil {
		var virtualNodes []string
		if err := json.Unmarshal([]byte(scoreEntity.Val), &virtualNodes); err != nil {
			return "", -1, err
		}
		return virtualNodes[0], scoreEntity.Score, nil
	}
	//往前找没有，需要寻找第一个节点
	scoreEntity, err = r.redisClient.FirstOrLast(ctx, r.getRingKey(), true)
	if err != nil {
		return "", -1, err
	}
	if scoreEntity == nil {
		return "", -1, nil
	}

	var virtualNodes []string
	if err := json.Unmarshal([]byte(scoreEntity.Val), &virtualNodes); err != nil {
		return "", -1, err
	}
	return virtualNodes[0], scoreEntity.Score, nil
}
