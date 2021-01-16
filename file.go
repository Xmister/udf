package udf

import (
	"io"
	"os"
	"time"
)

// File is a os.FileInfo-compatible wrapper around an ISO-13346 "UDF" directory entry
type File struct {
	Udf               *Udf
	Fid               *FileIdentifierDescriptor
	fe                FileEntryInterface
	fileEntryPosition uint64
}

// IsDir returns true if the entry is a directory or false otherwise
func (f *File) IsDir() bool {
	return f.FileEntry().GetICBTag().FileType == 4
}

// ModTime returns the entry's recording time
func (f *File) ModTime() time.Time {
	return f.FileEntry().GetModificationTime()
}

// Mode returns os.FileMode flag set with the os.ModeDir flag enabled in case of directories
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

// Name returns the base name of the given entry
func (f *File) Name() string {
	return f.Fid.FileIdentifier
}

// Size returns the size in bytes of the extent occupied by the file or directory
func (f *File) Size() int64 {
	return int64(f.FileEntry().GetInformationLength())
}

func (f *File) Sys() interface{} {
	return f.Fid
}

// ReadDir returns the children entries in case of a directory
func (f *File) ReadDir() []File {
	return f.Udf.ReadDir(f.FileEntry())
}

func (f *File) GetFileEntryPosition() int64 {
	return int64(f.fileEntryPosition)
}

func (f *File) FileEntry() FileEntryInterface {
	if f.fe == nil {
		f.fileEntryPosition = uint64(f.Fid.ICB.GetLocation())
		meta := f.Udf.LogicalPartitionStart(f.Fid.ICB.GetPartition())
		f.fe = NewFileEntry(f.Fid.ICB.GetPartition(), f.Udf.ReadSector(meta+f.fileEntryPosition))
	}
	return f.fe
}

func (f *File) getReaders(descs []ExtentInterface, filePos int64) (readers []*sectionReader, finalFilePos int64) {
	finalFilePos = filePos
	for i := 0; i < len(descs); i++ {
		if descs[i].HasExtended() {
			extendData := f.Udf.ReadSector(f.Udf.LogicalPartitionStart(descs[i].GetPartition()) + descs[i].GetLocation())
			aed := new(AED).FromBytes(extendData)
			var subReaders []*sectionReader
			subReaders, finalFilePos = f.getReaders(GetAllocationDescriptors(f.FileEntry().GetICBTag().AllocationType, extendData[24:], aed.LengthOfAllocationDescriptors), finalFilePos)
			readers = append(readers, subReaders...)
		} else if !descs[i].IsNotRecorded() {
			readers = append(readers, newSectionReader(finalFilePos, f.Udf.r, int64(f.Udf.SECTOR_SIZE)*int64(f.Udf.LogicalPartitionStart(descs[i].GetPartition())+descs[i].GetLocation()), int64(descs[i].GetLength())))
		}
		finalFilePos += int64(descs[i].GetLength())
	}
	return
}

func (f *File) NewReader() *MultiSectionReader {
	descs := f.FileEntry().GetAllocationDescriptors()
	readers, _ := f.getReaders(descs, 0)
	return newMultiSectionReader(readers)
}

type sectionReader struct {
	*io.SectionReader
	start int64
	size  int64
}

func newSectionReader(inFileStart int64, reader io.ReaderAt, start int64, size int64) *sectionReader {
	return &sectionReader{
		io.NewSectionReader(reader, start, size),
		inFileStart,
		size,
	}
}

type MultiSectionReader struct {
	readers []*sectionReader
	pos     int64
	size    int64
	index   int
}

func newMultiSectionReader(readers []*sectionReader) *MultiSectionReader {
	var size int64
	for _, reader := range readers {
		size += reader.size
	}
	return &MultiSectionReader{
		readers: readers,
		size:    size,
	}
}

func (r *MultiSectionReader) Read(p []byte) (n int, err error) {
	if r.index > len(r.readers)-1 {
		return 0, io.EOF
	}
	n, err = r.readers[r.index].Read(p)
	r.pos += int64(n)
	if err != nil {
		if r.index+1 < len(r.readers) {
			r.index++
			r.readers[r.index].Seek(0, 0)
			if n == 0 {
				n, err = r.Read(p)
			}
		}
	}
	if n > 0 {
		err = nil
	}
	return
}

func (r *MultiSectionReader) Seek(offset int64, whence int) (n int64, err error) {
	switch whence {
	case io.SeekStart:
		n = offset
	case io.SeekCurrent:
		n = offset + r.pos
	case io.SeekEnd:
		n = offset + r.size
	}
	if n > r.size {
		return 0, io.EOF
	}
	r.pos = n
	for i, reader := range r.readers {
		if reader.start <= n && reader.start+reader.size >= n {
			_, err = reader.Seek(n-reader.start, io.SeekStart)
			r.index = i
			break
		}
	}
	return
}

func (r *MultiSectionReader) ReadAt(p []byte, off int64) (n int, err error) {
	var read int
	err = os.ErrNotExist // No readers
	for _, reader := range r.readers {
		if reader.start <= off+int64(n) && reader.start+reader.size >= off+int64(n) {
			read, err = reader.ReadAt(p[n:], off+int64(n)-reader.start)
			n += read
			if (err != nil && err != io.EOF) || cap(p) == n {
				return
			}
		}
	}
	return
}

func (r *MultiSectionReader) Size() int64 {
	return r.size
}
