package fop

type op uint8

const (
	Read op = iota
	Write

)
type Entry struct {
	Op op
	fileName string
	Off *uint64
	Size *uint64
	Data []byte
}
type CEntry struct {
	Entry Entry
	Completed chan bool
}
