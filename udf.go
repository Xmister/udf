package udf

import (
	"io"
)

const SECTOR_SIZE = 2048

type Udf struct {
	r        io.ReaderAt
	isInited bool
	pvd      *PrimaryVolumeDescriptor
	pd       map[uint16]*PartitionDescriptor
	lvd      *LogicalVolumeDescriptor
	fsd      *FileSetDescriptor
	root_fe  FileEntryInterface
}

func (udf *Udf) PartitionStart(partition uint16) (physical uint64, meta uint64) {
	if udf.pd == nil {
		panic(udf)
	} else {
		physical = uint64(udf.pd[partition].PartitionStartingLocation)
		metaFile := NewFileEntry(udf.ReadSector(uint64(udf.pd[partition].PartitionStartingLocation)))
		if metaFile != nil && len(metaFile.GetAllocationDescriptors()) > 0 {
			meta = uint64(uint32(metaFile.GetAllocationDescriptors()[0].GetLocation()) + udf.pd[partition].PartitionStartingLocation)
		} else {
			meta = physical
		}
	}
	return
}

func (udf *Udf) GetReader() io.ReaderAt {
	return udf.r
}

func (udf *Udf) ReadSectors(sectorNumber uint64, sectorsCount uint64) []byte {
	buf := make([]byte, SECTOR_SIZE*sectorsCount)
	readed, err := udf.r.ReadAt(buf[:], int64(SECTOR_SIZE*sectorNumber))
	if err != nil {
		panic(err)
	}
	if readed != int(SECTOR_SIZE*sectorsCount) {
		panic(readed)
	}
	return buf[:]
}

func (udf *Udf) ReadSector(sectorNumber uint64) []byte {
	return udf.ReadSectors(sectorNumber, 1)
}

func (udf *Udf) init() {
	if udf.isInited {
		return
	}

	anchorDesc := NewAnchorVolumeDescriptorPointer(udf.ReadSector(256))
	if anchorDesc.Descriptor.TagIdentifier != DESCRIPTOR_ANCHOR_VOLUME_POINTER {
		panic(anchorDesc.Descriptor.TagIdentifier)
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

	_, partitionStart := udf.PartitionStart(0)

	udf.fsd = NewFileSetDescriptor(udf.ReadSector(partitionStart + uint64(udf.lvd.LogicalVolumeContentsUse.Location.LogicalBlockNumber)))
	udf.root_fe = NewFileEntry(udf.ReadSector(partitionStart + uint64(udf.fsd.RootDirectoryICB.Location.LogicalBlockNumber)))

	udf.isInited = true
}

func (udf *Udf) ReadDir(fe FileEntryInterface) []File {
	udf.init()

	if fe == nil {
		fe = udf.root_fe
	}

	_, ps := udf.PartitionStart(0)


	adPos := fe.GetAllocationDescriptors()[0]
	fdLen := uint64(adPos.GetLength())

	fdBuf := udf.ReadSectors(ps+uint64(adPos.GetLocation()), (fdLen+SECTOR_SIZE-1)/SECTOR_SIZE)
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

func NewUdfFromReader(r io.ReaderAt) *Udf {
	udf := &Udf{
		r:        r,
		isInited: false,
		pd:		  make(map[uint16]*PartitionDescriptor),
	}

	return udf
}
