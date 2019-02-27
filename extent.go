package udf

type Extent struct {
	Length   uint32
	Location uint32
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
