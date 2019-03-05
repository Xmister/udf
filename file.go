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

type sectionReader struct {
	*io.SectionReader
	start int64
	size int64
}

func newSectionReader(base int64, reader io.ReaderAt, start int64, size int64) *sectionReader {
	return &sectionReader{
		io.NewSectionReader(reader, start, size),
		start-base,
		size,
	}
}

type multiSectionReader struct {
	readers []*sectionReader
	pos int64
	size int64
	index int
}

func newMultiSectionReader(readers []*sectionReader) *multiSectionReader {
	var start, end int64
	for _, reader := range readers {
		if reader.start < start {
			start = reader.start
		}
		if reader.size+reader.start > end {
			end = reader.size+reader.start
		}
	}
	return &multiSectionReader{
		readers: readers,
		size: end-start,
	}
}

func (r *multiSectionReader) Read(p []byte) (n int, err error) {
	var read int
	n, err = io.ReadFull(r.readers[r.index], p)
	if err == io.ErrUnexpectedEOF {
		if r.index+1 < len(r.readers) {
			r.index++
			read, err = r.Read(p[n:])
			n+=read
		}
	}
	r.pos += int64(n)
	return
}

func (r *multiSectionReader) Seek(offset int64, whence int) (n int64, err error) {
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

func (r *multiSectionReader) ReadAt(p []byte, off int64) (n int, err error) {
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

func (f *File) GetFileEntryPosition() int64 {
	return int64(f.fileEntryPosition)
}

func (f *File) GetFileOffset(descriptor int) (res int64) {
	asd := f.FileEntry().GetAllocationDescriptors()[descriptor]
	partition := f.FileEntry().GetPartition()
	if f.FileEntry().GetICBTag().AllocationType == LongDescriptors {
		partition = asd.GetPartition()
	}
	res = int64(f.Udf.LogicalPartitionStart(partition) + asd.GetLocation())
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

func (f *File) NewReader() *multiSectionReader {
	descs := f.FileEntry().GetAllocationDescriptors()
	readers := make([]*sectionReader, len(descs))
	for i:=0; i<len(descs); i++ {
		readers[i] = newSectionReader(f.GetFileOffset(0), f.Udf.r, f.GetFileOffset(i), int64(descs[i].GetLength()))
	}
	return newMultiSectionReader(readers)
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
