/**
 * @author ysj
 * @email 2239831438@qq.com
 * @date 2024-09-26 19:17:10
 */

package csHash

import (
	"math"

	"github.com/spaolacci/murmur3"
)

type HashEncryptor interface {
	Encrypt(origin string) int32
}

type murmurHasher32 struct{}

func NewMurmurHasher32() *murmurHasher32 {
	return &murmurHasher32{}
}

// 哈希出的范围为 [0, 1<<31 - 2]
func (m *murmurHasher32) Encrypt(origin string) int32 {
	hasher := murmur3.New32()
	_, _ = hasher.Write([]byte(origin))
	return int32(hasher.Sum32() % math.MaxInt32)
}
