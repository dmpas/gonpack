package main

import "fmt"
import "os"
import "io"
import "path"
import "compress/flate"

const exeVersion = "3.0.36"
const exeCopyright = "\n\t2008 Denis Demidov 2008-03-30\n\t2014-2016 Sergey Batanov"

type Extractor struct {
	dirname string
}

type Parser struct {
	dirname string
}

// ProcessData просто выводит данные без обработки
func (un Extractor) ProcessData(f *os.File, filename string) error {
	targetFilePath := path.Join(un.dirname, filename)
	fileOut, err := os.Create(targetFilePath)
	if err != nil {
		return err
	}
	defer fileOut.Close()

	blockReader := GetBlockReader(f)
	_, err = io.Copy(fileOut, blockReader)

	return err
}

func (pr Parser) ProcessData(f *os.File, filename string) error {

	blockReader := GetBlockReader(f)
	inflater := flate.NewReader(blockReader)

	targetFilePath := path.Join(pr.dirname, filename)
	fileOut, err := os.Create(targetFilePath)
	if err != nil {
		return err
	}
	defer fileOut.Close()

	_, err = io.Copy(fileOut, inflater)
	return err
}

func unpack(filename, dirname string, withParse bool) {
	fstat, err := os.Stat(filename)
	if err != nil {
		panic(err)
	}

	FileDataSize := fstat.Size()

	fileIn, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer fileIn.Close()

	supplyDirectory(dirname)

	var processor DataProcessor

	if withParse {
		var p Parser
		p.dirname = dirname
		processor = p
	} else {
		var p Extractor
		p.dirname = dirname
		processor = p
	}

	UnpackToDirectoryNoLoad(dirname, fileIn, FileDataSize, processor)
}

func inflate(packedFileName, unpackedFileName string) {

	fi, err := os.Open(packedFileName)
	if err != nil {
		panic(err)
	}
	defer fi.Close()

	fo, err := os.Create(unpackedFileName)
	if err != nil {
		panic(err)
	}
	defer fo.Close()

	inflater := flate.NewReader(fi)
	io.Copy(fo, inflater)
}

func run(args []string) {

	if len(args) < 2 {
		usage()
		return
	}

	if args[1] == "-parse" || args[1] == "-p" {

		unpack(args[2], args[3], true)

	} else if args[1] == "-unpack" || args[1] == "-u" {

		unpack(args[2], args[3], false)

	} else if args[1] == "-inflate" || args[1] == "-i" {

		inflate(args[2], args[3])

	} else {
		usage()
	}
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
	run(os.Args)
}
