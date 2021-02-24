package downloader

type Chunk struct {
	Start   int64
	End     int64
	Data    []byte
	Size    int
	FlowNum int64
}
