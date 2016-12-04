package main

import (
	"encoding/binary"
	"io"
	"os"
)

const magicUndefined int32 = 0x7fffffff

type BlockHeader struct {
	DataSize     int32
	PageSize     int32
	NextPageAddr int32
}

type Block struct {
	Header           BlockHeader
	r                *os.File
	currentPageData  []byte
	currentPageIndex int
	NextPageAddr     int32
}

type blockHeader struct {
	EOL1            [2]byte
	DataSizeHex     [8]byte
	Space1          byte
	PageSizeHex     [8]byte
	Space2          byte
	NextPageAddrHex [8]byte
	Space3          byte
	EOL2            [2]byte
}

func (block *Block) setR(r *os.File) {
	block.r = r
}

func (block *Block) Read(buf []byte) (n int, err error) {

	if block.currentPageData == nil {
		// нужно обновить данные

		if block.NextPageAddr != 0 {

			// если установлен адрес страницы, идём к нему
			// (в противном случае, мы на самом первом блоке)

			if block.NextPageAddr == magicUndefined {
				return 0, nil
			}
			_, err := block.r.Seek(int64(block.NextPageAddr), os.SEEK_SET)
			if err != nil {
				return 0, err
			}

			// update block info
			err = block.Header.Read(block.r)
			if err != nil {
				return 0, err
			}
		}

		block.currentPageIndex = 0
		block.currentPageData = make([]byte, block.Header.PageSize)
		block.NextPageAddr = block.Header.NextPageAddr

		_, err := io.ReadFull(block.r, block.currentPageData)
		if err != nil {
			return 0, err
		}
	}

	err = nil
	n = copy(buf, block.currentPageData[block.currentPageIndex:])
	block.currentPageIndex += n
	if block.currentPageIndex >= len(block.currentPageData) {
		block.currentPageData = nil
		if block.NextPageAddr == magicUndefined {
			err = io.EOF
		}
	}

	return
}

func (header *BlockHeader) Read(r io.Reader) error {
	var serializedHeader blockHeader

	err := binary.Read(r, binary.LittleEndian, &serializedHeader)
	if err != nil {
		return err
	}

	if !serializedHeader.IsTrueV8() {
		panic("213")
	}

	header.DataSize, header.PageSize, header.NextPageAddr = serializedHeader.Decompose()
	return nil
}

// ReadDataBlock считывает блок данных из текущей позиции в файле
func ReadDataBlock(r *os.File) []byte {
	var b Block
	b.Header.Read(r)
	b.setR(r)

	buf := make([]byte, b.Header.DataSize)
	_, err := io.ReadFull(&b, buf)

	if err != nil {
		panic(err)
	}
	return buf
}

// TransferDataBlock пересылает считываемые данные блочного файла r прямо в Писателя w
func TransferDataBlock(r *os.File, w io.Writer) error {
	var b Block
	b.Header.Read(r)
	b.setR(r)

	_, err := io.Copy(w, &b)
	return err
}

func (header blockHeader) IsTrueV8() bool {
	if header.EOL1[0] != 0x0d ||
		header.EOL1[1] != 0x0a ||
		header.Space1 != 0x20 ||
		header.Space2 != 0x20 ||
		header.Space3 != 0x20 ||
		header.EOL2[0] != 0x0d ||
		header.EOL2[1] != 0x0a {
		return false
	}

	return true
}

func _httoi(b [8]byte) (result int32) {
	result = 0
	for i := 0; i < len(b); i++ {
		result *= 16
		if b[i] >= '0' && b[i] <= '9' {
			result += int32(b[i] - '0')
		} else if b[i] >= 'a' && b[i] <= 'f' {
			result += int32(b[i] - 'a' + 10)
		} else if b[i] >= 'A' && b[i] <= 'F' {
			result += int32(b[i] - 'A' + 10)
		}
	}
	return
}

func (header blockHeader) Decompose() (DataSize int32, PageSize int32, NextPageAddr int32) {
	DataSize = _httoi(header.DataSizeHex)
	PageSize = _httoi(header.PageSizeHex)
	NextPageAddr = _httoi(header.NextPageAddrHex)
	return
}
