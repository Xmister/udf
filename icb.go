package udf

type AllocationType int

const (
	ShortDescriptors AllocationType = iota
	LongDescriptors
	ExtendedDescriptors
	Embedded
)

type ICBTag struct {
	PriorRecordedNumberOfDirectEntries uint32
	StrategyType                       uint16
	StrategyParameter                  uint16
	MaximumNumberOfEntries             uint16
	FileType                           uint8
	ParentICBLocation                  uint64
	Flags                              uint16
	AllocationType                     AllocationType
}

func (itag *ICBTag) FromBytes(b []byte) *ICBTag {
	itag.PriorRecordedNumberOfDirectEntries = rl_u32(b[0:])
	itag.StrategyType = rl_u16(b[4:])
	itag.StrategyParameter = rl_u16(b[4:])
	itag.MaximumNumberOfEntries = rl_u16(b[8:])
	itag.FileType = r_u8(b[11:])
	itag.ParentICBLocation = rl_u48(b[12:])
	itag.Flags = rl_u16(b[18:])
	itag.AllocationType = (AllocationType)(itag.Flags & 0x3)
	return itag
}

func NewICBTag(b []byte) *ICBTag {
	return new(ICBTag).FromBytes(b)
}
