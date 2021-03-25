package bmputil

import (
	"bufio"
	"encoding/binary"
	"log"
	"os"
)

type BmpHeader struct {
	BfType    uint16
	BfSize    uint32
	Rfu1      uint16
	Rfu2      uint16
	BfOffBits uint32

	BiSize          uint32
	BiWidth         uint32
	BiHeight        uint32
	BiPlanes        uint16
	BiBitCount      uint16
	BiStyle         uint32
	BiStyleImage    uint32
	BiXPelsPerMeter uint32
	BiYPelsPerMeter uint32
	BiClrUsed       uint32
	BiClrImportant  uint32
}

type Size struct {
	Largeur uint32
	Hauteur uint32
}

const biSize uint32 = 40
const bfSize uint32 = 54

func entete2byte(entete *BmpHeader) []byte {
	buf := make([]byte, bfSize)
	binary.LittleEndian.PutUint16(buf[0:], entete.BfType)
	binary.LittleEndian.PutUint32(buf[2:], entete.BfSize)
	binary.LittleEndian.PutUint16(buf[6:], entete.Rfu1)
	binary.LittleEndian.PutUint16(buf[8:], entete.Rfu2)
	binary.LittleEndian.PutUint32(buf[10:], entete.BfOffBits)

	binary.LittleEndian.PutUint32(buf[14:], entete.BiSize)
	binary.LittleEndian.PutUint32(buf[18:], entete.BiWidth)
	binary.LittleEndian.PutUint32(buf[22:], entete.BiHeight)
	binary.LittleEndian.PutUint16(buf[26:], entete.BiPlanes)
	binary.LittleEndian.PutUint16(buf[28:], entete.BiBitCount)
	binary.LittleEndian.PutUint32(buf[30:], entete.BiStyle)
	binary.LittleEndian.PutUint32(buf[34:], entete.BiStyleImage)
	binary.LittleEndian.PutUint32(buf[38:], entete.BiXPelsPerMeter)
	binary.LittleEndian.PutUint32(buf[42:], entete.BiYPelsPerMeter)
	binary.LittleEndian.PutUint32(buf[46:], entete.BiClrUsed)
	binary.LittleEndian.PutUint32(buf[50:], entete.BiClrImportant)
	return buf
}

func SvBmp(name string, data *[]byte, size Size) {
	nb_octet_par_ligne := size.Largeur * 3
	sizeImg := nb_octet_par_ligne * size.Hauteur
	var entete BmpHeader = BmpHeader{
		0x4d42,
		bfSize + sizeImg,
		0, 0,
		bfSize, biSize,
		size.Largeur, size.Hauteur,
		1, 24, 0,
		sizeImg,
		8000, 8000,
		0, 0}

	file, err := os.Create(name)
	if err != nil {
		log.Fatal("Erreur de création du fichier destination.", err)
	}
	defer file.Close()
	w := bufio.NewWriter(file)
	w.Write(entete2byte(&entete))
	w.Write(*data)
}
