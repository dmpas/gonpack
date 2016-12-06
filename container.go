package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"unicode/utf16"
)

type InvalidElementHeader struct {
}

func (err InvalidElementHeader) Error() string {
	return "Invalid element header!"
}

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

type ElemHeader struct {
	DateCreation     uint64
	DateModification uint64
	Reserved         int32
}

func getElemName(data []byte) (header ElemHeader, name string, err error) {

	sz := binary.Size(header)
	if sz >= len(data) {
		err = InvalidElementHeader{}
		return
	}

	inbuffer := bytes.NewBuffer(data)
	binary.Read(inbuffer, binary.LittleEndian, &header)

	off := data[sz:]
	len := len(off) / 2
	utf16Data := make([]uint16, len)
	for i := 0; i < len; i++ {
		utf16Data[i] = uint16(off[i*2]) | uint16(off[i*2+1])<<8
	}

	name = string(utf16.Decode(utf16Data))
	return
}

// UnpackToDirectoryNoLoad Unpacks
func UnpackToDirectoryNoLoad(dirname string, file *os.File, FileDataSize int64, fileProcessor DataProcessor) error {
	supplyDirectory(dirname)

	var header FileHeader
	header.Read(file)

	elemAddrs, err := ReadDataBlock(file)
	if err != nil {
		return err
	}
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
		elemHeaderData, err := ReadDataBlock(file)
		if err != nil {
			return err
		}

		_, elemName, err := getElemName(elemHeaderData)
		if err != nil {
			return err
		}

		fmt.Printf("found file %s/%s ", dirname, elemName)

		if addr.ElemDataAddr != magicUndefined {
			file.Seek(int64(addr.ElemDataAddr), os.SEEK_SET)
			fileProcessor.ProcessData(file, elemName)
		}

		fmt.Println()
	}

	return nil
}
