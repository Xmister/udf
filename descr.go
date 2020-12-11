package udf

import (
	"fmt"
	"time"
)

const (
	DESCRIPTOR_PRIMARY_VOLUME            = 0x1
	DESCRIPTOR_ANCHOR_VOLUME_POINTER     = 0x2
	DESCRIPTOR_VOLUME_POINTER            = 0x3
	DESCRIPTOR_IMPLEMENTATION_USE_VOLUME = 0x4
	DESCRIPTOR_PARTITION                 = 0x5
	DESCRIPTOR_LOGICAL_VOLUME            = 0x6
	DESCRIPTOR_UNALLOCATED               = 0x7
	DESCRIPTOR_TERMINATING               = 0x8
	DESCRIPTOR_FILE_SET                  = 0x100
	DESCRIPTOR_IDENTIFIER                = 0x101
	DESCRIPTOR_ALLOCATION_EXTENT         = 0x102
	DESCRIPTOR_INDIRECT_ENTRY            = 0x103
	DESCRIPTOR_TERMINAL_ENTRY            = 0x104
	DESCRIPTOR_FILE_ENTRY                = 0x105
	UDF_EXTENT_FLAG_MASK                 = 0xC0000000
	EXT_NOT_RECORDED_ALLOCATED           = 0x40000000
	EXT_NOT_RECORDED_NOT_ALLOCATED       = 0x80000000
)

type Descriptor struct {
	TagIdentifier       uint16
	DescriptorVersion   uint16
	TagChecksum         uint8
	TagSerialNumber     uint16
	DescriptorCRC       uint16
	DescriptorCRCLength uint16
	TagLocation         uint32
	data                []byte
}

func (d *Descriptor) Data() []byte {
	buf := make([]byte, len(d.data))
	copy(buf, d.data[16:])
	return buf
}

func (d *Descriptor) Checksum() (checksum uint8) {
	for i := 0; i < 16; i++ {
		if i != 4 {
			checksum += d.data[i]
		}
	}
	return
}

func (d *Descriptor) FromBytes(b []byte) *Descriptor {
	d.TagIdentifier = rl_u16(b[0:])
	d.DescriptorVersion = rl_u16(b[2:])
	d.TagChecksum = r_u8(b[4:])
	d.TagSerialNumber = rl_u16(b[6:])
	d.DescriptorCRC = rl_u16(b[8:])
	d.DescriptorCRCLength = rl_u16(b[10:])
	d.TagLocation = rl_u32(b[12:])
	d.data = b[:]
	return d
}

func NewDescriptor(b []byte) *Descriptor {
	return new(Descriptor).FromBytes(b)
}

type AnchorVolumeDescriptorPointer struct {
	Descriptor                 Descriptor
	MainVolumeDescriptorSeq    Extent
	ReserveVolumeDescriptorSeq Extent
}

func (ad *AnchorVolumeDescriptorPointer) FromBytes(b []byte) *AnchorVolumeDescriptorPointer {
	ad.Descriptor.FromBytes(b)
	ad.MainVolumeDescriptorSeq = NewExtent(b[16:])
	ad.ReserveVolumeDescriptorSeq = NewExtent(b[24:])
	return ad
}

func NewAnchorVolumeDescriptorPointer(b []byte) *AnchorVolumeDescriptorPointer {
	return new(AnchorVolumeDescriptorPointer).FromBytes(b)
}

func (d *Descriptor) AnchorVolumeDescriptorPointer() *AnchorVolumeDescriptorPointer {
	return NewAnchorVolumeDescriptorPointer(d.data)
}

type PrimaryVolumeDescriptor struct {
	Descriptor                                  Descriptor
	VolumeDescriptorSequenceNumber              uint32
	PrimaryVolumeDescriptorNumber               uint32
	VolumeIdentifier                            string
	VolumeSequenceNumber                        uint16
	MaximumVolumeSequenceNumber                 uint16
	InterchangeLevel                            uint16
	MaximumInterchangeLevel                     uint16
	CharacterSetList                            uint32
	MaximumCharacterSetList                     uint32
	VolumeSetIdentifier                         string
	VolumeAbstract                              Extent
	VolumeCopyrightNoticeExtent                 Extent
	ApplicationIdentifier                       EntityID
	RecordingDateTime                           time.Time
	ImplementationIdentifier                    EntityID
	ImplementationUse                           []byte
	PredecessorVolumeDescriptorSequenceLocation uint32
	Flags                                       uint16
}

func (pvd *PrimaryVolumeDescriptor) FromBytes(b []byte) *PrimaryVolumeDescriptor {
	pvd.Descriptor.FromBytes(b)
	pvd.VolumeDescriptorSequenceNumber = rl_u32(b[16:])
	pvd.PrimaryVolumeDescriptorNumber = rl_u32(b[20:])
	pvd.VolumeIdentifier = r_dstring(b[24:], 32)
	pvd.VolumeSequenceNumber = rl_u16(b[56:])
	pvd.MaximumVolumeSequenceNumber = rl_u16(b[58:])
	pvd.InterchangeLevel = rl_u16(b[60:])
	pvd.MaximumInterchangeLevel = rl_u16(b[62:])
	pvd.CharacterSetList = rl_u32(b[64:])
	pvd.MaximumCharacterSetList = rl_u32(b[68:])
	pvd.VolumeSetIdentifier = r_dstring(b[72:], 128)
	pvd.VolumeAbstract = NewExtent(b[328:])
	pvd.VolumeCopyrightNoticeExtent = NewExtent(b[336:])
	pvd.ApplicationIdentifier = NewEntityID(b[344:])
	pvd.RecordingDateTime = r_timestamp(b[376:])
	pvd.ImplementationIdentifier = NewEntityID(b[388:])
	pvd.ImplementationUse = b[420:484]
	pvd.PredecessorVolumeDescriptorSequenceLocation = rl_u32(b[484:])
	pvd.Flags = rl_u16(b[488:])
	return pvd
}

func (pvd *PrimaryVolumeDescriptor) Show() {
	fmt.Printf("----- PrimaryVolumeDescriptor ----\n")
	fmt.Printf("VolumeDescriptorSequenceNumber: %d\n", pvd.VolumeDescriptorSequenceNumber)
	fmt.Printf("PrimaryVolumeDescriptorNumber: %d\n", pvd.PrimaryVolumeDescriptorNumber)
	fmt.Printf("VolumeIdentifier: %s\n", pvd.VolumeIdentifier)
	fmt.Printf("VolumeSequenceNumber: %d\n", pvd.VolumeSequenceNumber)
	fmt.Printf("MaximumVolumeSequenceNumber: %d\n", pvd.MaximumVolumeSequenceNumber)
	fmt.Printf("InterchangeLevel: %d\n", pvd.InterchangeLevel)
	fmt.Printf("MaximumInterchangeLevel: %d\n", pvd.MaximumInterchangeLevel)
	fmt.Printf("CharacterSetList: %d\n", pvd.CharacterSetList)
	fmt.Printf("MaximumCharacterSetList: %d\n", pvd.MaximumCharacterSetList)
	fmt.Printf("VolumeSetIdentifier: %s\n", pvd.VolumeSetIdentifier)
	fmt.Printf("VolumeAbstract: %v\n", pvd.VolumeAbstract)
	fmt.Printf("VolumeCopyrightNoticeExtent: %v\n", pvd.VolumeCopyrightNoticeExtent)
	pvd.ApplicationIdentifier.Show("ApplicationIdentifier")
	fmt.Printf("RecordingDateTime: %v\n", pvd.RecordingDateTime)
	pvd.ImplementationIdentifier.Show("ImplementationIdentifier")
	fmt.Printf("ImplementationUse: %v\n", pvd.ImplementationUse)
	fmt.Printf("PredecessorVolumeDescriptorSequenceLocation: %v\n", pvd.PredecessorVolumeDescriptorSequenceLocation)
	fmt.Printf("Flags: %d\n", pvd.Flags)
	fmt.Printf("----- PrimaryVolumeDescriptor (end) ----\n")
}

func NewPrimaryVolumeDescriptor(b []byte) *PrimaryVolumeDescriptor {
	return new(PrimaryVolumeDescriptor).FromBytes(b)
}

func (d *Descriptor) PrimaryVolumeDescriptor() *PrimaryVolumeDescriptor {
	return NewPrimaryVolumeDescriptor(d.data)
}

type PartitionDescriptor struct {
	Descriptor                     Descriptor
	VolumeDescriptorSequenceNumber uint32
	PartitionFlags                 uint16
	PartitionNumber                uint16
	PartitionContents              EntityID
	PartitionContentsUse           []byte
	AccessType                     uint32
	PartitionStartingLocation      uint32
	PartitionLength                uint32
	ImplementationIdentifier       EntityID
	ImplementationUse              []byte
}

func (pd *PartitionDescriptor) FromBytes(b []byte) *PartitionDescriptor {
	pd.Descriptor.FromBytes(b)
	pd.VolumeDescriptorSequenceNumber = rl_u32(b[16:])
	pd.PartitionFlags = rl_u16(b[20:])
	pd.PartitionNumber = rl_u16(b[22:])
	pd.PartitionContents = NewEntityID(b[24:])
	pd.PartitionContentsUse = b[56:184]
	pd.AccessType = rl_u32(b[184:])
	pd.PartitionStartingLocation = rl_u32(b[188:])
	pd.PartitionLength = rl_u32(b[192:])
	pd.ImplementationIdentifier = NewEntityID(b[196:])
	pd.ImplementationUse = b[228:356]
	return pd
}

func (pd *PartitionDescriptor) Show(i uint16) {
	fmt.Printf("----- PartitionDescriptor %d ----\n", i)
	fmt.Printf("VolumeDescriptorSequenceNumber: %d\n", pd.VolumeDescriptorSequenceNumber)
	fmt.Printf("PartitionFlags: %d\n", pd.PartitionFlags)
	fmt.Printf("PartitionNumber: %d\n", pd.PartitionNumber)
	pd.ImplementationIdentifier.Show("PartitionContents")
	fmt.Printf("PartitionContentsUse: %v\n", pd.PartitionContentsUse)
	fmt.Printf("AccessType: %d\n", pd.AccessType)
	fmt.Printf("PartitionStartingLocation: %d\n", pd.PartitionStartingLocation)
	fmt.Printf("PartitionLength: %d\n", pd.PartitionLength)
	pd.ImplementationIdentifier.Show("ImplementationIdentifier")
	fmt.Printf("ImplementationUse: %v\n", pd.ImplementationUse)
	fmt.Printf("----- PartitionDescriptor %d (end) ----\n", i)
}

func NewPartitionDescriptor(b []byte) *PartitionDescriptor {
	return new(PartitionDescriptor).FromBytes(b)
}

func (d *Descriptor) PartitionDescriptor() *PartitionDescriptor {
	return NewPartitionDescriptor(d.data)
}

type PartitionMap struct {
	PartitionMapType     uint8
	PartitionMapLength   uint8
	VolumeSequenceNumber uint16
	PartitionNumber      uint16
	PartitionStart       uint32
}

func (pm *PartitionMap) FromBytes(b []byte) *PartitionMap {
	pm.PartitionMapType = rb_u8(b[0:])
	pm.PartitionMapLength = rb_u8(b[1:])
	offset := 0
	switch pm.PartitionMapType {
	case 1:
		offset = 2
	case 2:
		offset = 36
	}
	pm.VolumeSequenceNumber = rb_u16(b[offset:])
	pm.PartitionNumber = rb_u16(b[offset+2:])
	pm.PartitionStart = uint32(pm.VolumeSequenceNumber)
	return pm
}

func (pm *PartitionMap) Show(i uint32) {
	fmt.Printf("----- PartitionMap %d ----\n", i)
	fmt.Printf("PartitionMapType: %d\n", pm.PartitionMapType)
	fmt.Printf("PartitionMapLength: %d\n", pm.PartitionMapLength)
	fmt.Printf("VolumeSequenceNumber: %d\n", pm.VolumeSequenceNumber)
	fmt.Printf("PartitionNumber: %d\n", pm.PartitionNumber)
	fmt.Printf("PartitionStart: %d\n", pm.PartitionStart)
	fmt.Printf("----- PartitionMap %d (end) ----\n", i)
}

type LogicalVolumeDescriptor struct {
	Descriptor                     Descriptor
	VolumeDescriptorSequenceNumber uint32
	LogicalVolumeIdentifier        string
	LogicalBlockSize               uint32
	DomainIdentifier               EntityID
	LogicalVolumeContentsUse       ExtentLong
	MapTableLength                 uint32
	NumberOfPartitionMaps          uint32
	ImplementationIdentifier       EntityID
	ImplementationUse              []byte
	IntegritySequenceExtent        Extent
	PartitionMaps                  []PartitionMap
}

func (lvd *LogicalVolumeDescriptor) FromBytes(b []byte) *LogicalVolumeDescriptor {
	lvd.Descriptor.FromBytes(b)
	lvd.VolumeDescriptorSequenceNumber = rl_u32(b[16:])
	lvd.LogicalVolumeIdentifier = r_dstring(b[84:], 128)
	lvd.LogicalBlockSize = rl_u32(b[212:])
	lvd.DomainIdentifier = NewEntityID(b[216:])
	lvd.LogicalVolumeContentsUse = NewExtentLong(b[248:])
	lvd.MapTableLength = rl_u32(b[264:])
	lvd.NumberOfPartitionMaps = rl_u32(b[268:])
	lvd.ImplementationIdentifier = NewEntityID(b[272:])
	lvd.ImplementationUse = b[304:432]
	lvd.IntegritySequenceExtent = NewExtent(b[432:])
	lvd.PartitionMaps = make([]PartitionMap, lvd.NumberOfPartitionMaps)
	for i := uint32(0); i < lvd.NumberOfPartitionMaps; i++ {
		lvd.PartitionMaps[i].FromBytes(b[440+i*6:])
	}
	return lvd
}

func (lvd *LogicalVolumeDescriptor) Show() {
	fmt.Printf("----- LogicalVolumeDescriptor ----\n")
	fmt.Printf("VolumeDescriptorSequenceNumber: %v\n", lvd.VolumeDescriptorSequenceNumber)
	fmt.Printf("LogicalVolumeIdentifier: %v\n", lvd.LogicalVolumeIdentifier)
	fmt.Printf("LogicalBlockSize: %v\n", lvd.LogicalBlockSize)
	lvd.DomainIdentifier.Show("DomainIdentifier")
	fmt.Printf("LogicalVolumeContentsUse: %v\n", lvd.LogicalVolumeContentsUse)
	fmt.Printf("MapTableLength: %v\n", lvd.MapTableLength)
	fmt.Printf("NumberOfPartitionMaps: %v\n", lvd.NumberOfPartitionMaps)
	lvd.ImplementationIdentifier.Show("ImplementationIdentifier")
	fmt.Printf("ImplementationUse: %v\n", lvd.ImplementationUse)
	fmt.Printf("IntegritySequenceExtent: %v\n", lvd.IntegritySequenceExtent)
	fmt.Printf("PartitionMaps: %v\n", lvd.PartitionMaps)
	for i := uint32(0); i < lvd.NumberOfPartitionMaps; i++ {
		lvd.PartitionMaps[i].Show(i)
	}
	fmt.Printf("----- LogicalVolumeDescriptor (end) ----\n")
}

func NewLogicalVolumeDescriptor(b []byte) *LogicalVolumeDescriptor {
	return new(LogicalVolumeDescriptor).FromBytes(b)
}

func (d *Descriptor) LogicalVolumeDescriptor() *LogicalVolumeDescriptor {
	return NewLogicalVolumeDescriptor(d.data)
}

type FileSetDescriptor struct {
	Descriptor              Descriptor
	RecordingDateTime       time.Time
	InterchangeLevel        uint16
	MaximumInterchangeLevel uint16
	CharacterSetList        uint32
	MaximumCharacterSetList uint32
	FileSetNumber           uint32
	FileSetDescriptorNumber uint32
	LogicalVolumeIdentifier string
	FileSetIdentifier       string
	CopyrightFileIdentifier string
	AbstractFileIdentifier  string
	RootDirectoryICB        ExtentLong
	DomainIdentifier        EntityID
	NexExtent               ExtentLong
}

func (fsd *FileSetDescriptor) FromBytes(b []byte) *FileSetDescriptor {
	fsd.Descriptor.FromBytes(b)
	fsd.RecordingDateTime = r_timestamp(b[16:])
	fsd.InterchangeLevel = rl_u16(b[28:])
	fsd.MaximumInterchangeLevel = rl_u16(b[30:])
	fsd.CharacterSetList = rl_u32(b[32:])
	fsd.MaximumCharacterSetList = rl_u32(b[36:])
	fsd.FileSetNumber = rl_u32(b[40:])
	fsd.FileSetDescriptorNumber = rl_u32(b[44:])
	fsd.LogicalVolumeIdentifier = r_dstring(b[112:], 128)
	fsd.FileSetIdentifier = r_dstring(b[304:], 32)
	fsd.CopyrightFileIdentifier = r_dstring(b[336:], 32)
	fsd.AbstractFileIdentifier = r_dstring(b[368:], 32)
	fsd.RootDirectoryICB = NewExtentLong(b[400:])
	fsd.DomainIdentifier = NewEntityID(b[416:])
	fsd.NexExtent = NewExtentLong(b[448:])
	return fsd
}

func NewFileSetDescriptor(b []byte) *FileSetDescriptor {
	return new(FileSetDescriptor).FromBytes(b)
}

func (d *Descriptor) FileSetDescriptor() *FileSetDescriptor {
	return NewFileSetDescriptor(d.data)
}

type FileIdentifierDescriptor struct {
	Descriptor                Descriptor
	FileVersionNumber         uint16
	FileCharacteristics       uint8
	LengthOfFileIdentifier    uint8
	ICB                       ExtentLong
	LengthOfImplementationUse uint16
	ImplementationUse         EntityID
	FileIdentifier            string
}

func (fid *FileIdentifierDescriptor) Len() uint64 {
	l := 38 + uint64(fid.LengthOfImplementationUse) + uint64(fid.LengthOfFileIdentifier)
	return 4 * ((l + 3) / 4) // padding = 4
}

func (fid *FileIdentifierDescriptor) FromBytes(b []byte) *FileIdentifierDescriptor {
	fid.Descriptor.FromBytes(b)
	fid.FileVersionNumber = rl_u16(b[16:])
	fid.FileCharacteristics = r_u8(b[18:])
	fid.LengthOfFileIdentifier = r_u8(b[19:])
	fid.ICB = NewExtentLong(b[20:])
	fid.LengthOfImplementationUse = rl_u16(b[36:])
	fid.ImplementationUse = NewEntityID(b[38:])
	identStart := 38 + fid.LengthOfImplementationUse
	fid.FileIdentifier = r_dcharacters(b[identStart : fid.LengthOfFileIdentifier+uint8(identStart)])
	return fid
}

func NewFileIdentifierDescriptor(b []byte) *FileIdentifierDescriptor {
	return new(FileIdentifierDescriptor).FromBytes(b)
}

func (d *Descriptor) FileIdentifierDescriptor() *FileIdentifierDescriptor {
	return NewFileIdentifierDescriptor(d.data)
}

type FileEntryInterface interface {
	GetAllocationDescriptors() []ExtentInterface
	GetPermissions() uint32
	GetInformationLength() uint64
	GetModificationTime() time.Time
	GetICBTag() *ICBTag
	GetPartition() uint16
}

type FileEntry struct {
	Descriptor                    Descriptor
	ICBTag                        *ICBTag
	Uid                           uint32
	Gid                           uint32
	Permissions                   uint32
	FileLinkCount                 uint16
	RecordFormat                  uint8
	RecordDisplayAttributes       uint8
	RecordLength                  uint32
	InformationLength             uint64
	LogicalBlocksRecorded         uint64
	AccessTime                    time.Time
	ModificationTime              time.Time
	AttributeTime                 time.Time
	Checkpoint                    uint32
	ExtendedAttributeICB          ExtentLong
	ImplementationIdentifier      EntityID
	UniqueId                      uint64
	LengthOfExtendedAttributes    uint32
	LengthOfAllocationDescriptors uint32
	ExtendedAttributes            []byte
	AllocationDescriptors         []byte
	// Manual field
	Partition uint16
}

type ExtendedFileEntry struct {
	FileEntry
	CreationTime       time.Time
	ObjectSize         uint64
	StreamDirectoryIcb ExtentLong
}

func (fe *FileEntry) FromBytes(b []byte) *FileEntry {
	fe.Descriptor.FromBytes(b)
	fe.ICBTag = NewICBTag(b[16:])
	fe.Uid = rl_u32(b[36:])
	fe.Gid = rl_u32(b[40:])
	fe.Permissions = rl_u32(b[44:])
	fe.FileLinkCount = rl_u16(b[48:])
	fe.RecordFormat = r_u8(b[50:])
	fe.RecordDisplayAttributes = r_u8(b[51:])
	fe.RecordLength = rl_u32(b[52:])
	fe.InformationLength = rl_u64(b[56:])
	fe.LogicalBlocksRecorded = rl_u64(b[64:])
	fe.AccessTime = r_timestamp(b[72:])
	fe.ModificationTime = r_timestamp(b[84:])
	fe.AttributeTime = r_timestamp(b[96:])
	fe.Checkpoint = rl_u32(b[108:])
	fe.ExtendedAttributeICB = NewExtentLong(b[112:])
	fe.ImplementationIdentifier = NewEntityID(b[128:])
	fe.UniqueId = rl_u64(b[160:])
	fe.LengthOfExtendedAttributes = rl_u32(b[168:])
	fe.LengthOfAllocationDescriptors = rl_u32(b[172:])
	allocDescStart := 176 + fe.LengthOfExtendedAttributes
	if allocDescStart > uint32(len(b)) {
		return nil
	}
	fe.ExtendedAttributes = b[176:allocDescStart]
	fe.AllocationDescriptors = b[allocDescStart : allocDescStart+fe.LengthOfAllocationDescriptors]
	return fe
}

func (fe *FileEntry) GetPartition() uint16 {
	if fe.ICBTag.AllocationType == LongDescriptors {
		return fe.GetAllocationDescriptors()[0].GetPartition()
	}
	return fe.Partition
}

func GetAllocationDescriptor(t AllocationType, b []byte) ExtentInterface {
	switch t {
	case ShortDescriptors:
		return NewExtent(b)
	case LongDescriptors:
		return NewExtentLong(b)
	case ExtendedDescriptors:
		return NewExtentExtended(b)
	}
	return nil
}

func GetAllocationDescriptors(t AllocationType, b []byte, len uint32) (list []ExtentInterface) {
	var descLen uint32
	switch t {
	case ShortDescriptors:
		descLen = 8
	case LongDescriptors:
		descLen = 16
	case ExtendedDescriptors:
		descLen = 24
	default:
		return
	}
	list = make([]ExtentInterface, len/descLen)
	for i := range list {
		list[i] = GetAllocationDescriptor(t, b[uint32(i)*descLen:])
	}

	return
}

func (fe *FileEntry) GetAllocationDescriptors() (list []ExtentInterface) {
	return GetAllocationDescriptors(fe.ICBTag.AllocationType, fe.AllocationDescriptors, fe.LengthOfAllocationDescriptors)
}

func (fe *FileEntry) GetPermissions() uint32 {
	return fe.Permissions
}

func (fe *FileEntry) GetInformationLength() uint64 {
	return fe.InformationLength
}

func (fe *FileEntry) GetModificationTime() time.Time {
	return fe.ModificationTime
}

func (fe *FileEntry) GetICBTag() *ICBTag {
	return fe.ICBTag
}

func NewFileEntry(partition uint16, b []byte) (fe FileEntryInterface) {
	if e := new(FileEntry).FromBytes(b); e == nil {
		ee := new(ExtendedFileEntry).FromBytes(b)
		ee.Partition = partition
		fe = ee
	} else {
		e.Partition = partition
		fe = e
	}
	return
}

func (fe *ExtendedFileEntry) FromBytes(b []byte) *ExtendedFileEntry {
	fe.Descriptor.FromBytes(b)
	fe.ICBTag = NewICBTag(b[16:])
	fe.Uid = rl_u32(b[36:])
	fe.Gid = rl_u32(b[40:])
	fe.Permissions = rl_u32(b[44:])
	fe.FileLinkCount = rl_u16(b[48:])
	fe.RecordFormat = r_u8(b[50:])
	fe.RecordDisplayAttributes = r_u8(b[51:])
	fe.RecordLength = rl_u32(b[52:])
	fe.InformationLength = rl_u64(b[56:])
	fe.ObjectSize = rl_u64(b[64:])
	fe.LogicalBlocksRecorded = rl_u64(b[72:])
	fe.AccessTime = r_timestamp(b[80:])
	fe.ModificationTime = r_timestamp(b[92:])
	fe.CreationTime = r_timestamp(b[104:])
	fe.AttributeTime = r_timestamp(b[116:])
	fe.Checkpoint = rl_u32(b[128:])
	fe.ExtendedAttributeICB = NewExtentLong(b[136:])
	fe.StreamDirectoryIcb = NewExtentLong(b[152:])
	fe.ImplementationIdentifier = NewEntityID(b[168:])
	fe.UniqueId = rl_u64(b[200:])
	fe.LengthOfExtendedAttributes = rl_u32(b[208:])
	fe.LengthOfAllocationDescriptors = rl_u32(b[212:])
	allocDescStart := 216 + fe.LengthOfExtendedAttributes
	if allocDescStart > uint32(len(b)) {
		return nil
	}
	fe.ExtendedAttributes = b[216:allocDescStart]
	fe.AllocationDescriptors = b[allocDescStart : allocDescStart+fe.LengthOfAllocationDescriptors]
	return fe
}

func (d *Descriptor) FileEntry() FileEntryInterface {
	return NewFileEntry(0, d.data)
}
