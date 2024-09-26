/**
 * @author ysj
 * @email 2239831438@qq.com
 * @date 2024-09-26 19:06:13
 */

package csHash

type ConsistentHash struct {
	hashRing  HashRing
	migrator  Migrator
	encryptor HashEncryptor
	logger    Logger
	opts      ConsistentHashOptions
}
