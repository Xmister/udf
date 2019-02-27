package udf

import (
	"io"
	"os"
	"time"
)

type File struct {
	Udf               *Udf
	Fid               *FileIdentifierDescriptor
	fe                FileEntryInterface
	fileEntryPosition uint64
}

func (f *File) GetFileEntryPosition() int64 {
	return int64(f.fileEntryPosition)
}

func (f *File) GetFileOffset() int64 {
	return SECTOR_SIZE * (int64(f.FileEntry().GetAllocationDescriptors()[0].Location) + int64(f.Udf.PartitionStart(0)))
}

func (f *File) FileEntry() FileEntryInterface {
	if f.fe == nil {
		f.fileEntryPosition = uint64(f.Fid.ICB.Location.LogicalBlockNumber)
		f.fe = NewFileEntry(f.Udf.ReadSector(f.Udf.PartitionStart(0) + f.fileEntryPosition))
	}
	return f.fe
}

func (f *File) NewReader() *io.SectionReader {
	return io.NewSectionReader(f.Udf.r, f.GetFileOffset(), f.Size())
}

func (f *File) Name() string {
	return f.Fid.FileIdentifier
}

func (f *File) Mode() os.FileMode {
	var mode os.FileMode

	perms := os.FileMode(f.FileEntry().GetPermissions())
	mode |= ((perms >> 0) & 7) << 0
	mode |= ((perms >> 5) & 7) << 3
	mode |= ((perms >> 10) & 7) << 6

	if f.IsDir() {
		mode |= os.ModeDir
	}

	return mode
}

func (f *File) Size() int64 {
	return int64(f.FileEntry().GetInformationLength())
}

func (f *File) ModTime() time.Time {
	return f.FileEntry().GetModificationTime()
}

func (f *File) IsDir() bool {
	return f.FileEntry().GetICBTag().FileType == 4
}

func (f *File) Sys() interface{} {
	return f.Fid
}

func (f *File) ReadDir() []File {
	return f.Udf.ReadDir(f.FileEntry())
}
