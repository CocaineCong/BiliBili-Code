package bloomfilter

import (
	"hash"
	"math"
	"sync"

	"github.com/bits-and-blooms/bitset"
	"github.com/twmb/murmur3"
)

// BloomFilter 结构体
type BloomFilter struct {
	bitset    *bitset.BitSet // 使用第三方bitset包
	size      uint64         // 位集合大小
	hashFuncs []hash.Hash64  // 哈希函数集合
	mutex     sync.RWMutex   // 读写锁
	// hashMutex sync.Mutex     // 添加专门的哈希函数互斥锁
}

// NewWithFalsePositiveRate 根据预期元素数量和误判率创建布隆过滤器
func NewWithFalsePositiveRate(expectedItems uint64, falsePositiveRate float64) *BloomFilter {
	// 计算最优位数组大小和哈希函数数量
	m, k := optimalParameters(expectedItems, falsePositiveRate)
	funcs := make([]hash.Hash64, k)
	for i := 0; i < k; i++ {
		funcs[i] = murmur3.SeedNew64(uint64(i))
	}
	return &BloomFilter{
		bitset:    bitset.New(uint(m)),
		size:      m,
		hashFuncs: funcs,
		mutex:     sync.RWMutex{},
	}
}

// optimalParameters 计算最优参数
func optimalParameters(n uint64, p float64) (uint64, int) {
	m := uint64(math.Ceil(-float64(n) * math.Log(p) / (math.Ln2 * math.Ln2)))
	k := int(math.Ceil(math.Ln2 * float64(m) / float64(n)))
	return m, k
}

// createHashFunctions 创建哈希函数集合
// func createHashFunctions(k int) []hash.Hash64 {
// 	funcs := make([]hash.Hash64, k)
// 	for i := 0; i < k; i++ {
// 		funcs[i] = murmur3.SeedNew64(uint64(i))
// 	}
// 	return funcs
// }

// Add 添加元素到布隆过滤器
func (bf *BloomFilter) Add(item []byte) {
	bf.mutex.Lock()
	defer bf.mutex.Unlock()

	// bf.hashMutex.Lock()
	// defer bf.hashMutex.Unlock()

	for _, h := range bf.hashFuncs {
		h.Reset()
		_, _ = h.Write(item)
		index := h.Sum64() % bf.size
		bf.bitset.Set(uint(index))
	}
}

// Contains 检查元素是否可能在布隆过滤器中
func (bf *BloomFilter) Contains(item []byte) bool {
	bf.mutex.RLock()
	defer bf.mutex.RUnlock()

	// bf.hashMutex.Lock()
	// defer bf.hashMutex.Unlock()

	for _, h := range bf.hashFuncs {
		h.Reset()
		_, _ = h.Write(item)
		index := h.Sum64() % bf.size
		if !bf.bitset.Test(uint(index)) {
			return false
		}
	}
	return true
}
