package main

import (
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"raw2bmp/bmputil"
	"raw2bmp/files"
)

const hauteur, largeur uint32 = 960, 1280

func min(a, b uint32) uint32 {
	if a <= b {
		return a
	}
	return b
}
func max(a, b uint32) uint32 {
	if a >= b {
		return a
	}
	return b
}

func main() {
	rep := "./raw/"
	raws := files.GetRaws(rep)
	var width, height uint32
	var bayer []byte
	for _, name := range raws {
		raw, err := ioutil.ReadFile(rep + name)
		if err != nil {
			panic(err)
		}
		width, height = binary.BigEndian.Uint32(raw[20:24]), binary.BigEndian.Uint32(raw[24:28])
		bayer = raw[28:]
		fmt.Printf("%v : Size:%v×%v, len:%v\n", name, width, height, len(bayer))

		fmt.Println("Préparation table d'homographie...")
		//tbhomo := homographie.ConstituerMatriceDistortion()

		const rougeRef, vertRef, bleuRef byte = 201, 207, 197
		const ratioRV uint32 = (102400 * uint32(rougeRef)) / (uint32(vertRef))
		const ratioRB uint32 = (102400 * uint32(rougeRef)) / (uint32(bleuRef))
		const ratioRR uint32 = 1024
		step := width
		step2 := 2 * width
		image := make([]byte, hauteur*largeur*3)
		for y := uint32(0); y < hauteur; y++ {
			ofd := y * largeur
			for x := uint32(0); x < largeur; x++ {
				ofs := y*step2 + x*2 // tbhomo[y][x]
				if step2*2 < ofs && ofs < step*(height-2)-2 && (ofs%step) > 1 && (ofs%step) < step-2 {
					rouge := uint32(bayer[ofs])
					t0 := uint32((bayer[ofs-step-1] + bayer[ofs-step+1] + bayer[ofs+step-1] + bayer[ofs+step+1]) << 1)
					t0 -= uint32(((bayer[ofs-step2]+bayer[ofs-2]+bayer[ofs+2]+bayer[ofs+step2])*3 + 1) >> 1)
					t0 += uint32(rouge * 6)
					t1 := uint32((bayer[ofs-step] + bayer[ofs-1] + bayer[ofs+1] + bayer[ofs+step]) >> 1)
					t1 -= uint32((bayer[ofs-step2] + bayer[ofs-2] + bayer[ofs+2] + bayer[ofs+step2]))
					t1 += uint32(rouge << 2)
					bleu := max(0, min((t0+4)>>3, 255))
					vert := max(0, min((t1+4)>>3, 255))
					bleu = min(255, (bleu*ratioRB)>>10)
					vert = min(255, (vert*ratioRV)>>10)
					rouge = min(255, (rouge*ratioRR)>>10)
					image[(y*largeur+x)*3] = byte(bleu)
					image[(y*largeur+x)*3+1] = byte(vert)
					image[(y*largeur+x)*3+2] = byte(rouge)
				}
				ofd += 1
			}
		}
		var size bmputil.Size = bmputil.Size{largeur, hauteur}
		name = name[:len(name)-3] + "bmp"
		fmt.Printf("%v\n", image[:100])
		bmputil.SvBmp(name, &image, size)
	}
}
