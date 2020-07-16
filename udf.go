package udf

import (
	"errors"
	"io"
)

type Udf struct {
	r        io.ReaderAt
	isInited bool
	pvd      *PrimaryVolumeDescriptor
	pd       map[uint16]*PartitionDescriptor
	lvd      *LogicalVolumeDescriptor
	fsd      *FileSetDescriptor
	root_fe  FileEntryInterface
	SECTOR_SIZE uint64
}

func (udf *Udf) PhysicalPartitionStart(partition uint16) (physical uint64) {
	if udf.pd == nil {
		panic(udf)
	} else {
		return uint64(udf.pd[partition].PartitionStartingLocation)
	}
}

func (udf *Udf) LogicalPartitionStart(partition uint16) (logical uint64) {
	if udf.lvd == nil || len(udf.lvd.PartitionMaps) < 1 {
		panic(udf)
	} else {
		return uint64(udf.lvd.PartitionMaps[partition].PartitionStart)
	}
}

func (udf *Udf) GetReader() io.ReaderAt {
	return udf.r
}

func (udf *Udf) ReadSectors(sectorNumber uint64, sectorsCount uint64) []byte {
	buf := make([]byte, udf.SECTOR_SIZE*sectorsCount)
	read, err := udf.r.ReadAt(buf[:], int64(udf.SECTOR_SIZE*sectorNumber))
	if err != nil {
		panic(err)
	}
	/*if readed != int(udf.SECTOR_SIZE*sectorsCount) {
		panic(readed)
	}*/
	return buf[:read]
}

func (udf *Udf) ReadSector(sectorNumber uint64) []byte {
	return udf.ReadSectors(sectorNumber, 1)
}

func (udf *Udf) init() (err error) {
	if udf.isInited {
		return
	}

	var anchorDesc *AnchorVolumeDescriptorPointer

	for udf.SECTOR_SIZE = 512; udf.SECTOR_SIZE <= 32768; udf.SECTOR_SIZE <<= 1 {
		anchorDesc = NewAnchorVolumeDescriptorPointer(udf.ReadSector(256))
		if anchorDesc.Descriptor.TagIdentifier == DESCRIPTOR_ANCHOR_VOLUME_POINTER &&
			anchorDesc.Descriptor.TagChecksum == anchorDesc.Descriptor.Checksum() {
			break
		}
	}

	if anchorDesc.Descriptor.TagIdentifier != DESCRIPTOR_ANCHOR_VOLUME_POINTER ||
		anchorDesc.Descriptor.TagChecksum != anchorDesc.Descriptor.Checksum() {
		err = errors.New("couldn't find sector size")
		return
	}

	for sector := uint64(anchorDesc.MainVolumeDescriptorSeq.Location); ; sector++ {
		desc := NewDescriptor(udf.ReadSector(sector))
		if desc.TagIdentifier == DESCRIPTOR_TERMINATING {
			break
		}
		switch desc.TagIdentifier {
		case DESCRIPTOR_PRIMARY_VOLUME:
			udf.pvd = desc.PrimaryVolumeDescriptor()
		case DESCRIPTOR_PARTITION:
			pd := desc.PartitionDescriptor()
			udf.pd[pd.PartitionNumber] = pd
		case DESCRIPTOR_LOGICAL_VOLUME:
			udf.lvd = desc.LogicalVolumeDescriptor()
		}
	}

	for i, pMap := range udf.lvd.PartitionMaps {
		if pMap.PartitionMapType != 2 {
			udf.lvd.PartitionMaps[i].PartitionStart = udf.pd[pMap.PartitionNumber].PartitionStartingLocation
			continue
		}
		metaFile := NewFileEntry(0, udf.ReadSector(uint64(udf.pd[pMap.PartitionNumber].PartitionStartingLocation)))
		if metaFile != nil && len(metaFile.GetAllocationDescriptors()) > 0 {
			udf.lvd.PartitionMaps[i].PartitionStart = uint32(metaFile.GetAllocationDescriptors()[0].GetLocation()) + udf.pd[pMap.PartitionNumber].PartitionStartingLocation
		}
	}

	partitionStart := udf.LogicalPartitionStart(udf.lvd.LogicalVolumeContentsUse.GetPartition())

	udf.fsd = NewFileSetDescriptor(udf.ReadSector(partitionStart + uint64(udf.lvd.LogicalVolumeContentsUse.Location.LogicalBlockNumber)))
	rootICB := udf.fsd.RootDirectoryICB
	udf.root_fe = NewFileEntry(udf.lvd.LogicalVolumeContentsUse.GetPartition(), udf.ReadSector(udf.LogicalPartitionStart(rootICB.GetPartition()) + rootICB.GetLocation()))

	udf.isInited = true
	return
}

func (udf *Udf) ReadDir(fe FileEntryInterface) []File {
	udf.init()
	if fe == nil {
		fe = udf.root_fe
	}

	ps := udf.LogicalPartitionStart(fe.GetPartition())
	adPos := fe.GetAllocationDescriptors()[0]
	fdLen := uint64(adPos.GetLength())


	fdBuf := udf.ReadSectors(ps+adPos.GetLocation(), (fdLen+udf.SECTOR_SIZE-1)/udf.SECTOR_SIZE)
	fdOff := uint64(0)

	result := make([]File, 0)

	for uint32(fdOff) < adPos.GetLength() {
		fid := NewFileIdentifierDescriptor(fdBuf[fdOff:])
		if fid.FileIdentifier != "" {
			result = append(result, File{
				Udf: udf,
				Fid: fid,
			})
		}
		fdOff += fid.Len()
	}
	return result
}

func NewUdfFromReader(r io.ReaderAt) (udf *Udf, err error) {
	udf = &Udf{
		r:        r,
		isInited: false,
		pd:		  make(map[uint16]*PartitionDescriptor),
	}

	err = udf.init()
	return
}
