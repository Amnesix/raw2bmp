package main

import (
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"log"
	"raw2bmp/bmputil"
	"raw2bmp/files"
	"raw2bmp/homographie"

	"gopkg.in/ini.v1"
)

const hauteur, largeur uint32 = 960, 1280
const rep string = "./raw/"

var coins [4]homographie.Point = [4]homographie.Point{{64, 39}, {1256, 39}, {1259, 916}, {54, 917}}
var degre = 4
var coefs [5]float64 = [5]float64{1.05, -2.468025e-06, -8.355279e-08, -5.611001e-11, 8.427150e-14}

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

func init() {
	cfg, err := ini.Load(rep + "camera-PARAMS.ini")
	if err != nil {
		log.Fatal("Il manque le fichier camera-PARAM.ini. ", err)
	}
	var key string
	var groupe string = "CoinsTonneau"
	for i := 0; i < 4; i++ {
		key = fmt.Sprintf("Coin%d_", i)
		coins[i].X = float64(cfg.Section(groupe).Key(key + "x").MustFloat64(coins[i].X))
		coins[i].Y = float64(cfg.Section(groupe).Key(key + "y").MustFloat64(coins[i].Y))
	}
	groupe = "Polynome"
	degre = cfg.Section(groupe).Key("Na").MustInt(degre)
	for i := 0; i < 5; i++ {
		key = fmt.Sprintf("a_%d", i)
		coefs[i] = cfg.Section(groupe).Key(key).MustFloat64(coefs[i])
	}
}

func main() {
	raws := files.GetRaws(rep)
	var width, height uint32
	var bayer []byte
	fmt.Println(coins)
	tbhomo := homographie.ConstituerMatriceDistortion(coins, degre, coefs)
	img := make([]byte, largeur*hauteur*4)
	for h := uint32(0); h < hauteur; h++ {
		for l := uint32(0); l < largeur; l++ {
			binary.LittleEndian.PutUint32(img[(h*hauteur+l)*4:], tbhomo[h][l])
		}
	}
	bmputil.SvBmp("homographie.bmp", &img, bmputil.Size{largeur, hauteur}, 32)
	for _, name := range raws {
		raw, err := ioutil.ReadFile(rep + name)
		if err != nil {
			panic(err)
		}
		width, height = binary.BigEndian.Uint32(raw[20:24]), binary.BigEndian.Uint32(raw[24:28])
		bayer = raw[28:]
		fmt.Printf("%v : Size:%v×%v, len:%v\n", name, width, height, len(bayer))

		fmt.Println("Préparation table d'homographie...")

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
				ofs := tbhomo[y][x]
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
		bmputil.SvBmp(name, &image, size, 24)
	}
}
