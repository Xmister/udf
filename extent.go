package udf

type ExtentInterface interface {
	GetLocation() uint64
	GetLength() uint32
	SetLength(uint32)
	GetPartition() uint16
	IsNotRecorded() bool
	HasExtended() bool
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

func (e Extent) SetLength(length uint32)  {
	e.Length =  length
}

func (e Extent) IsNotRecorded() bool  {
	return (e.Length & UDF_EXTENT_FLAG_MASK) == EXT_NOT_RECORDED_ALLOCATED  || (e.Length & UDF_EXTENT_FLAG_MASK) == EXT_NOT_RECORDED_NOT_ALLOCATED
}

func (e Extent) HasExtended() bool  {
	return (e.Length >> 30) == 3
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

func (e ExtentSmall) SetLength(length uint32)  {
	e.Length = uint16(length)
}

func (e ExtentSmall) IsNotRecorded() bool  {
	return false
}

func (e ExtentSmall) HasExtended() bool  {
	return (e.Length >> 30) == 3
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

func (e ExtentLong) SetLength(length uint32)  {
	e.Length =  length
}

func (e ExtentLong) HasExtended() bool  {
	return (e.Length >> 30) == 3
}

func (e ExtentLong) IsNotRecorded() bool  {
	return (e.Length & UDF_EXTENT_FLAG_MASK) == EXT_NOT_RECORDED_ALLOCATED || (e.Length & UDF_EXTENT_FLAG_MASK) == EXT_NOT_RECORDED_NOT_ALLOCATED
}

func NewExtentLong(b []byte) ExtentLong {
	return ExtentLong{
		Length:   rl_u32(b[0:]),
		Location: new(LbAddr).FromBytes(b[4:]),
	}
}

type ExtentExtended struct {
	ExtentLength	uint32
	RecordedLength	uint32
	InfoLength		uint32
	Location		LbAddr
}

func (e ExtentExtended) GetPartition() uint16 {
	return e.Location.PartitionReferenceNumber
}

func (e ExtentExtended) GetLocation() uint64 {
	return uint64(e.Location.LogicalBlockNumber)
}

func (e ExtentExtended) GetLength() uint32 {
	return e.InfoLength
}

func (e ExtentExtended) SetLength(length uint32)  {
	e.InfoLength =  length
}

func (e ExtentExtended) HasExtended() bool  {
	return (e.GetLength() >> 30) == 3
}

func (e ExtentExtended) IsNotRecorded() bool  {
	return false
}

func NewExtentExtended(b []byte) ExtentExtended {
	return ExtentExtended{
		ExtentLength:   rl_u32(b[0:]),
		RecordedLength:   rl_u32(b[4:]),
		InfoLength:   rl_u32(b[8:]),
		Location: new(LbAddr).FromBytes(b[12:]),
	}
}

type AED struct {
	Descriptor	Descriptor
	PreviousAllocationExtentLocation uint32
	LengthOfAllocationDescriptors uint32
}

func (a *AED) FromBytes(b []byte) AED {
	a.Descriptor.FromBytes(b)
	a.PreviousAllocationExtentLocation = rl_u32(b[16:])
	a.LengthOfAllocationDescriptors = rl_u32(b[20:])
	return *a
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
