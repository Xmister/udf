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

func (f *File) GetFileOffset() (res int64) {
	asd := f.FileEntry().GetAllocationDescriptors()[0]
	res = int64(f.Udf.LogicalPartitionStart(f.FileEntry().GetPartition())) + int64(asd.GetLocation())
	return int64(SECTOR_SIZE) * res
}

func (f *File) FileEntry() FileEntryInterface {
	if f.fe == nil {
		f.fileEntryPosition = uint64(f.Fid.ICB.GetLocation())
		meta := f.Udf.LogicalPartitionStart(f.Fid.ICB.GetPartition())
		f.fe = NewFileEntry(f.Fid.ICB.GetPartition(), f.Udf.ReadSector(meta + f.fileEntryPosition))
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
