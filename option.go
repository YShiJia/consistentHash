/**
 * @author ysj
 * @email 2239831438@qq.com
 * @date 2024-09-26 19:06:53
 */

package csHash

type ConsistentHashOption func(*ConsistentHashOptions)

type ConsistentHashOptions struct {
	//超时时间
	lockExpireSeconds int64
	//副本数量
	replicas int64
	//日志级别
	loggerLevel LoggerLevel
}

// lockExpireSeconds 锁的过期时间，单位秒, 默认15秒
func WithLockExpireSeconds(seconds int64) ConsistentHashOption {
	return func(opts *ConsistentHashOptions) {
		opts.lockExpireSeconds = seconds
	}
}

// replicas 单个节点映射出来的副本数量，范围为[1,10]
func WithReplicas(replicas int64) ConsistentHashOption {
	return func(opts *ConsistentHashOptions) {
		opts.replicas = replicas
	}
}

func (opts *ConsistentHashOptions) repair() {
	//必须有超时时限
	if opts.lockExpireSeconds <= 0 {
		opts.lockExpireSeconds = 15
	}

	switch {
	case opts.replicas <= 0:
		opts.replicas = 5
	case opts.replicas > 10:
		opts.replicas = 10
	}
}
