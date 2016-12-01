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
			block.currentPageIndex = 0
			block.currentPageData = make([]byte, block.Header.DataSize)
		}

		block.NextPageAddr = block.Header.NextPageAddr

		curn := 0
		for {
			ng, err := block.r.Read(block.currentPageData[curn:block.Header.PageSize])
			if err != nil {
				return 0, err
			}
			curn += ng
		}
	}

	err = nil
	dataLeft := len(block.currentPageData) - block.currentPageIndex
	if dataLeft >= len(buf) {
		n = len(buf)
	} else {
		n = dataLeft
	}

	copy(buf, block.currentPageData[block.currentPageIndex:])
	block.currentPageIndex += n
	if block.currentPageIndex == len(block.currentPageData) {
		block.currentPageData = nil
		block.currentPageIndex = 0
	}

	return
}

func (header *BlockHeader) Read(r io.Reader) error {
	var serializedHeader blockHeader

	err := binary.Read(r, binary.LittleEndian, &serializedHeader)
	if err != nil {
		return err
	}

	header.DataSize, header.PageSize, header.NextPageAddr = serializedHeader.Decompose()
	return nil
}

func ReadDataBlock(r io.Reader) []byte {
	var bh BlockHeader
	bh.Read(r)

	var b Block
	b.Header = bh

	buf := make([]byte, bh.DataSize)
	_, err := io.ReadFull(r, buf)

	if err != nil {
		panic(err)
	}
	return buf
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
		result = result * 16
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
