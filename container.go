package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

type DataProcessor interface {
	ProcessData(f *os.File, filename string) error
}

// FileHeader is a file header
type FileHeader struct {
	NextPageAddr int32
	PageSize     int32
	StorageVer   int32
	Reserved     int32
}

type ElemAddr struct {
	ElemHeaderAddr int32
	ElemDataAddr   int32
	Reserved       int32
}

func ElemAddrSize() int {
	return 4 + 4 + 4
}

func (header *FileHeader) Read(r io.Reader) {
	binary.Read(r, binary.LittleEndian, header)
}

type Elem struct {
	DateCreation     uint64
	DateModification uint64
	Reserved         int32
}

func getElemName(data []byte) string {
	var e Elem
	sz := binary.Size(e)
	if sz >= len(data) {
		return ""
	}
	off := data[sz:]
	var buffer bytes.Buffer

	for i := 0; i < len(off); i += 2 {
		if off[i] >= 32 && off[i] < 128 {
			buffer.WriteByte(off[i])
		}
	}
	return buffer.String()
}

// UnpackToDirectoryNoLoad Unpacks
func UnpackToDirectoryNoLoad(dirname string, file *os.File, FileDataSize int64, fileProcessor DataProcessor) {
	supplyDirectory(dirname)

	var header FileHeader
	header.Read(file)

	elemAddrs := ReadDataBlock(file)
	elemsNum := len(elemAddrs) / ElemAddrSize()

	elemsReader := bytes.NewReader(elemAddrs)
	for i := 0; i < elemsNum; i++ {
		var addr ElemAddr
		binary.Read(elemsReader, binary.LittleEndian, &addr)

		if addr.Reserved != magicUndefined {
			fmt.Printf("Done at %d\n", i)
			break
		}

		file.Seek(int64(addr.ElemHeaderAddr), os.SEEK_SET)
		elemHeaderData := ReadDataBlock(file)
		elemName := getElemName(elemHeaderData)

		fmt.Printf("found file %s ", elemName)

		if addr.ElemDataAddr != magicUndefined {
			file.Seek(int64(addr.ElemDataAddr), os.SEEK_SET)
			fileProcessor.ProcessData(file, elemName)
		}

		fmt.Println()
	}
}
