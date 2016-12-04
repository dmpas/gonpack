package main

import "fmt"
import "os"
import "path"

const exeVersion = "3.0.36"
const exeCopyright = "\n\t2008 Denis Demidov 2008-03-30\n\t2014-2016 Sergey Batanov"

type Unpacker struct {
	dirname string
}

type Parser struct {
	dirname string
}

// ProcessData просто выводит данные без обработки
func (un Unpacker) ProcessData(f *os.File, filename string) error {
	targetFilePath := path.Join(un.dirname, filename)
	fileOut, err := os.Create(targetFilePath)
	if err != nil {
		return err
	}
	return TransferDataBlock(f, fileOut)
}

func (pr Parser) ProcessData(f *os.File, filename string) error {
	return nil
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

	supplyDirectory(dirname)

	var p Unpacker
	p.dirname = dirname
	UnpackToDirectoryNoLoad(dirname, fileIn, FileDataSize, p)
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
