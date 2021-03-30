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

const hauteur, largeur int32 = 960, 1280
const rep string = "./raw/"

var coins [4]homographie.Point = [4]homographie.Point{{64, 39}, {1256, 39}, {1259, 916}, {54, 917}}
var degre = 4
var coefs [5]float64 = [5]float64{1.05, -2.468025e-06, -8.355279e-08, -5.611001e-11, 8.427150e-14}

func min(a, b int32) int32 {
	if a <= b {
		return a
	}
	return b
}
func max(a, b int32) int32 {
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
	var width, height int32
	var bayer []byte
	fmt.Println(coins)
	fmt.Println("Préparation table d'homographie...")
	tbhomo := homographie.ConstituerMatriceDistortion(coins, degre, coefs)
	img := make([]byte, largeur*hauteur*4)
	for ofs := int32(0); ofs < hauteur*largeur; ofs++ {
		binary.LittleEndian.PutUint32(img[ofs*4:], uint32(tbhomo[ofs]))
	}
	bmputil.SvBmp("homographie.bmp", &img, bmputil.Size{uint32(largeur), uint32(hauteur)}, 32)
	for _, name := range raws {
		raw, err := ioutil.ReadFile(rep + name)
		if err != nil {
			panic(err)
		}
		width, height = int32(binary.BigEndian.Uint32(raw[20:24])), int32(binary.BigEndian.Uint32(raw[24:28]))
		bayer = raw[28:]
		fmt.Printf("%v : Size:%v×%v, len:%v\n", name, width, height, len(bayer))

		var rougeRef, vertRef, bleuRef byte = 201, 207, 197
		var ratioRV int32 = (102400 * int32(rougeRef)) / (int32(vertRef) * 100)
		var ratioRB int32 = (102400 * int32(rougeRef)) / (int32(bleuRef) * 100)
		var ratioRR int32 = 1024
		step := width
		step2 := 2 * width
		image := make([]byte, hauteur*largeur*3)
		var ofd, ofs int32
		var x, y int32
		for y = 2; y < hauteur-2; y++ {
			ofd = y * largeur
			for x = 0; x < largeur; x++ {
				ofs = tbhomo[ofd]
				if step2 < ofs && ofs < width*height-step2 && x > 1 && x < largeur-2 {
					rouge := int32(bayer[ofs])
					t0 := int32((bayer[ofs-step-1] + bayer[ofs-step+1] + bayer[ofs+step-1] + bayer[ofs+step+1]) << 1)
					t0 -= int32(((bayer[ofs-step2]+bayer[ofs-2]+bayer[ofs+2]+bayer[ofs+step2])*3 + 1) >> 1)
					t0 += int32(rouge * 6)
					t1 := int32((bayer[ofs-step] + bayer[ofs-1] + bayer[ofs+1] + bayer[ofs+step]) << 1)
					t1 -= int32((bayer[ofs-step2] + bayer[ofs-2] + bayer[ofs+2] + bayer[ofs+step2]))
					t1 += int32(rouge << 2)
					bleu := max(0, min((t0+4)>>3, 255))
					vert := max(0, min((t1+4)>>3, 255))
					bleu = min(255, (bleu*ratioRB)>>10)
					vert = min(255, (vert*ratioRV)>>10)
					rouge = min(255, (rouge*ratioRR)>>10)
					image[ofd*3] = byte(bleu)
					image[ofd*3+1] = byte(vert)
					image[ofd*3+2] = byte(rouge)
				}
				ofd++
			}
		}
		var size bmputil.Size = bmputil.Size{uint32(largeur), uint32(hauteur)}
		name = name[:len(name)-3] + "bmp"
		bmputil.SvBmp(name, &image, size, 24)
	}
}
