package main

import "fmt"
import "os"
import "io"
import "encoding/binary"
import "bytes"

const exeVersion = "3.0.36"
const exeCopyright = "\n\t2008 Denis Demidov 2008-03-30\n\t2014-2016 Sergey Batanov"

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

func SmartUnpackData(data []byte) {

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
func UnpackToDirectoryNoLoad(dirname string, file *os.File, FileDataSize int64) {
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

		fmt.Printf("found file %s", elemName)

		if addr.ElemDataAddr != magicUndefined {
			file.Seek(int64(addr.ElemDataAddr), os.SEEK_SET)
			elemDataData := ReadDataBlock(file)
			SmartUnpackData(elemDataData)

			fmt.Printf(" with size %d", len(elemDataData))
		}

		fmt.Println()
	}
}

func parse(filename, dirname string) {
	fstat, err := os.Stat(filename)
	if err != nil {
		panic(err)
	}

	FileDataSize := fstat.Size()

	fileIn, err := os.Open(filename)
	if err != nil {
		panic(err)
	}

	UnpackToDirectoryNoLoad(dirname, fileIn, FileDataSize)
}

func run(args []string) {
	parse(args[2], args[3])
}

func usage() {
	fmt.Println()
	fmt.Printf("V8Upack Version %s Copyright (c) %s\n", exeVersion, exeCopyright)
	fmt.Println()
	fmt.Println("Unpack, pack, deflate and inflate 1C v8 file (*.cf)")
	fmt.Println()
	fmt.Println("V8UNPACK")
	fmt.Println("  -U[NPACK]     in_filename.cf     out_dirname")
	fmt.Println("  -PA[CK]       in_dirname         out_filename.cf")
	fmt.Println("  -I[NFLATE]    in_filename.data   out_filename")
	fmt.Println("  -D[EFLATE]    in_filename        filename.data")
	fmt.Println("  -E[XAMPLE]")
	fmt.Println("  -BAT")
	fmt.Println("  -P[ARSE]      in_filename        out_dirname")
	fmt.Println("  -B[UILD]      in_dirname         out_filename")
	fmt.Println("  -V[ERSION]")
}

func version() {
	fmt.Println(exeVersion)
}

func main() {
	// usage()
	//
	// run(os.Args)
	run([]string{os.Args[0], "-parse", "test.epf", "testdir"})
}
