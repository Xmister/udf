package udf

import (
	"errors"
	"io"
	"unsafe"
)

// Udf is a wrapper around an .iso file that allows reading its ISO-13346 "UDF" data
type Udf struct {
	r           io.ReaderAt
	isInited    bool
	pvd         *PrimaryVolumeDescriptor
	pd          map[uint16]*PartitionDescriptor
	lvd         *LogicalVolumeDescriptor
	fsd         *FileSetDescriptor
	root_fe     FileEntryInterface
	SECTOR_SIZE uint64
}

// NewUdfFromReader returns an Udf reader reading from a given file
func NewUdfFromReader(r io.ReaderAt) (*Udf, error) {
	udf := &Udf{
		r:        r,
		isInited: false,
		pd:       make(map[uint16]*PartitionDescriptor),
	}

	err := udf.init()
	return udf, err
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
	//fmt.Printf("udf.SECTOR_SIZE = %d\n", udf.SECTOR_SIZE)

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

	// DEBUGGING ONLY
	// udf.pvd.Show()
	// for i, pd := range udf.pd {
	// 	pd.Show(i)
	// }
	// udf.lvd.Show()
	// DEBUGGING ONLY - end

	for i, pMap := range udf.lvd.PartitionMaps {
		if pMap.PartitionMapType != 2 {
			// Check to error early if there is no match with a partition number
			if _, ok := udf.pd[pMap.PartitionNumber]; !ok {
				return errors.New("could not find partition number")
			}
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
	udf.root_fe = NewFileEntry(udf.lvd.LogicalVolumeContentsUse.GetPartition(), udf.ReadSector(udf.LogicalPartitionStart(rootICB.GetPartition())+rootICB.GetLocation()))

	udf.isInited = true
	return
}

func (udf *Udf) ReadSector(sectorNumber uint64) []byte {
	return udf.ReadSectors(sectorNumber, 1)
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
	var dummy *FileIdentifierDescriptor
	fidSize := int(unsafe.Sizeof(*dummy))
	for uint32(fdOff) < adPos.GetLength() {
		// Some Windows ISOs have some padding data that we can ignore?
		if len(fdBuf[fdOff:]) < fidSize {
			//fmt.Printf("WARNING: skipping incomplete data\n")
			break
		}
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

// XXX - unused
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
