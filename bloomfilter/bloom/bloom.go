package bloomfilter

import (
	"math"
	"sync"
)

// BloomFilter 布隆过滤器结构体
type BloomFilter struct {
	bitArray  []byte       // 二进制位数组
	bitSize   int          // 位数组大小
	hashFuncs int          // 哈希函数数量
	mutex     sync.RWMutex // 互斥锁，保证并发安全
}

//// NewBloomFilter 创建一个新的布隆过滤器
//// expectedElements: 预计插入的元素数量
//// falsePositiveRate: 期望的误判率(0-1之间)
//func NewBloomFilter(expectedElements int, falsePositiveRate float64) *BloomFilter {
//	if expectedElements <= 0 {
//		expectedElements = 1000
//	}
//	if falsePositiveRate <= 0 || falsePositiveRate >= 1 {
//		falsePositiveRate = 0.01
//	}
//
//	// 计算最优位数组大小
//	bitSize := int(-float64(expectedElements) * math.Log(falsePositiveRate) / (math.Log(2) * math.Log(2)))
//	if bitSize < 1 {
//		bitSize = 1024
//	}
//
//	// 计算最优哈希函数数量
//	hashFuncs := int(float64(bitSize) / float64(expectedElements) * math.Log(2))
//	if hashFuncs < 1 {
//		hashFuncs = 1
//	}
//
//	return &BloomFilter{
//		bitArray:  make([]byte, (bitSize+7)/8), // 转换为字节数
//		bitSize:   bitSize,
//		hashFuncs: hashFuncs,
//	}
//}

// 替换原 NewBloomFilter，严格按公式计算最优参数
func NewBloomFilter(expectedElements int, falsePositiveRate float64) *BloomFilter {
	if expectedElements <= 0 {
		expectedElements = 1000
	}
	if falsePositiveRate <= 0 || falsePositiveRate >= 1 {
		falsePositiveRate = 0.01 // 兜底默认 1%
	}

	// 1. 计算最优位数组大小 m
	m := -int(math.Ceil(float64(expectedElements) * math.Log(falsePositiveRate) / math.Pow(math.Log(2), 2)))
	if m < 1 {
		m = 1024
	}

	// 2. 计算最优哈希函数数量 k
	k := int(math.Ceil(float64(m) / float64(expectedElements) * math.Log(2)))
	if k < 1 {
		k = 1
	}

	return &BloomFilter{
		bitArray:  make([]byte, (m+7)/8), // 转字节数组
		bitSize:   m,
		hashFuncs: k,
	}
}

// 添加元素到布隆过滤器
func (bf *BloomFilter) Add(item string) {
	bf.mutex.Lock()
	defer bf.mutex.Unlock()

	// 生成多个哈希值并设置对应位为1
	for i := 0; i < bf.hashFuncs; i++ {
		hash := bf.hash(item, i)
		bf.setBit(hash)
	}
}

// 检查元素是否可能在布隆过滤器中
func (bf *BloomFilter) MightContain(item string) bool {
	bf.mutex.RLock()
	defer bf.mutex.RUnlock()

	// 检查所有对应位是否都为1
	for i := 0; i < bf.hashFuncs; i++ {
		hash := bf.hash(item, i)
		if !bf.getBit(hash) {
			return false
		}
	}
	return true
}

// 计算元素的哈希值
//func (bf *BloomFilter) hash(item string, index int) int {
//	// 使用不同的种子生成不同的哈希函数
//	h := fnv.New64a()
//	h.Write([]byte(item))
//	h.Write([]byte(strconv.Itoa(index)))
//	return int(h.Sum64() % uint64(bf.bitSize))
//}

// 原： 用 fnv 单一哈希 + 加盐，冲突概率高。
// 改：用双哈希 + 组合生成多哈希（减少冲突）
func (bf *BloomFilter) hash(item string, index int) int {
	// 双哈希种子（可自定义）
	seed1 := uint64(0x123456789ABCDEF0)
	seed2 := uint64(0xFEDCBA9876543210)

	// 生成两个基础哈希
	h1 := murmurHash3(item, seed1)
	h2 := murmurHash3(item, seed2)

	// 组合生成第 k 个哈希（避免重复计算）
	return int((h1 + uint64(index)*h2) % uint64(bf.bitSize))
}

// 高效哈希：MurmurHash3（比 fnv 更适合布隆，冲突更低）
func murmurHash3(data string, seed uint64) uint64 {
	h := seed
	const c1 = uint64(0xcc9e2d51)
	const c2 = uint64(0x1b873593)
	const r1 = 15
	const r2 = 13
	const m = 5
	const n = 0xe6546b64

	bytes := []byte(data)
	length := len(bytes)
	blocks := length / 4

	for i := 0; i < blocks; i++ {
		block := uint64(bytes[i*4]) | uint64(bytes[i*4+1])<<8 |
			uint64(bytes[i*4+2])<<16 | uint64(bytes[i*4+3])<<24
		block *= c1
		block = (block << r1) | (block >> (64 - r1))
		block *= c2
		h ^= block
		h = (h << r2) | (h >> (64 - r2))
		h = h*m + n
	}

	// 处理剩余字节
	remainder := bytes[blocks*4:]
	var k uint64
	switch len(remainder) {
	case 3:
		k ^= uint64(remainder[2]) << 16
		fallthrough
	case 2:
		k ^= uint64(remainder[1]) << 8
		fallthrough
	case 1:
		k ^= uint64(remainder[0])
		k *= c1
		k = (k << r1) | (k >> (64 - r1))
		k *= c2
		h ^= k
	}

	h ^= uint64(length)
	h ^= h >> 16
	h *= 0x85ebca6b
	h ^= h >> 13
	h *= 0xc2b2ae35
	h ^= h >> 16

	return h
}

// 设置指定位为1
func (bf *BloomFilter) setBit(bitIndex int) {
	byteIndex := bitIndex / 8
	bitPosition := uint(bitIndex % 8)
	bf.bitArray[byteIndex] |= 1 << bitPosition
}

// 获取指定位的值
func (bf *BloomFilter) getBit(bitIndex int) bool {
	byteIndex := bitIndex / 8
	bitPosition := uint(bitIndex % 8)
	return (bf.bitArray[byteIndex] & (1 << bitPosition)) != 0
}
