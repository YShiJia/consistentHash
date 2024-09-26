/**
 * @author ysj
 * @email 2239831438@qq.com
 * @date 2024-09-26 19:06:53
 */

package csHash

type ConsistentHashOptions struct {
	//超时时间
	lockExpireSeconds int64
	//副本数量
	replicas int64
	//日志级别
	loggerLevel LoggerLevel
}
