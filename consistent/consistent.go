package consistent


type Hasher interface {
	Sum64([]byte) uint64
}