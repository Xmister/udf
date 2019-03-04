package udf

type ExtentInterface interface {
	GetLocation() uint64
	GetLength() uint32
	GetPartition() uint16
}

type Extent struct {
	Length   uint32
	Location uint32
}

func (e Extent) GetPartition() uint16 {
	return 0
}

func (e Extent) GetLocation() uint64 {
	return uint64(e.Location)
}

func (e Extent) GetLength() uint32 {
	return e.Length
}

func NewExtent(b []byte) Extent {
	return Extent{
		Length:   rl_u32(b[0:]),
		Location: rl_u32(b[4:]),
	}
}

type ExtentSmall struct {
	Length   uint16
	Location uint64
}

func (e ExtentSmall) GetPartition() uint16 {
	return 0
}

func (e ExtentSmall) GetLocation() uint64 {
	return uint64(e.Location)
}

func (e ExtentSmall) GetLength() uint32 {
	return uint32(e.Length)
}

func NewExtentSmall(b []byte) ExtentSmall {
	return ExtentSmall{
		Length:   rl_u16(b[0:]),
		Location: rl_u48(b[2:]),
	}
}

type ExtentLong struct {
	Length   uint32
	Location LbAddr
}

func (e ExtentLong) GetPartition() uint16 {
	return e.Location.PartitionReferenceNumber
}

func (e ExtentLong) GetLocation() uint64 {
	return uint64(e.Location.LogicalBlockNumber)
}

func (e ExtentLong) GetLength() uint32 {
	return e.Length
}

func NewExtentLong(b []byte) ExtentLong {
	return ExtentLong{
		Length:   rl_u32(b[0:]),
		Location: new(LbAddr).FromBytes(b[4:]),
	}
}

type LbAddr struct {
	LogicalBlockNumber uint32
	PartitionReferenceNumber uint16
}

func (l *LbAddr) FromBytes(data []byte) LbAddr {
	l.LogicalBlockNumber = rl_u32(data[0:])
	l.PartitionReferenceNumber = rl_u16(data[4:])
	return *l
}
