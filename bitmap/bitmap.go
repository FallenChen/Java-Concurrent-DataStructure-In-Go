package bitmap

// BitMap implement bitmap
type BitMap []byte

//数据范围在1到1亿之间，表达1千万的数据，只需要1亿个二进制位
//如果数据范围在1到10亿之间，表达1千万的数据，需要10亿个二进制位
//使用Bloom Filter判断，对一个数进行k个Hash函数，获得k个值，然后分别存入

// New create BitMap
func New(length uint) BitMap {
	return make([]byte, length/8+1)
}

// Set
func (b BitMap) Set(value uint) {
	byteIndex := value / 8
	if byteIndex >= uint(len(b)) {
		return
	}
	bitIndex := value % 8
	[]byte(b)[byteIndex] |= 1 << bitIndex
}

// Get
func (b BitMap) Get(value uint) bool {
	byteIndex := value / 8
	if byteIndex >= uint(len(b)) {
		return false
	}
	bitIndex := value % 8
	return []byte(b)[byteIndex]&(1<<bitIndex) != 0
}
