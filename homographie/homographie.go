package homographie

import (
	"log"
	"math"
)

type Point struct {
	X, Y float64
}

const wsp, hsp uint32 = 1280, 960
const degres uint32 = 4

func min(a, b uint32) uint32 {
	if a <= b {
		return a
	}
	return b
}

func coeffs() [5]float64 {
	return [5]float64{1.05, -2.468025e-06, -8.355279e-08, -5.611001e-11, 8.427150e-14}
}

func coins() [4]Point {
	return [4]Point{{64, 39}, {1256, 39}, {1259, 916}, {54, 917}}
}

func Norme(p1, p2 Point) float64 {
	return math.Sqrt(math.Pow(p1.X-p2.X, 2) + math.Pow(p1.Y-p2.Y, 2))
}

func MAT_GetPerspectiveTransform(P [4]Point) [9]float64 {
	var H [9]float64

	sx := (P[0].X - P[1].X) + (P[2].X - P[3].X)
	sy := (P[0].Y - P[1].Y) + (P[2].Y - P[3].Y)

	if sx == 0. || sy == 0. {
		H[0] = P[1].X - P[0].X
		H[1] = P[2].X - P[1].X
		H[2] = P[0].X
		H[3] = P[1].Y - P[0].Y
		H[4] = P[2].Y - P[1].Y
		H[5] = P[0].Y
		H[6] = 0.
		H[7] = 0.
	} else {
		dx1 := P[1].X - P[2].X
		dx2 := P[3].X - P[2].X
		dy1 := P[1].Y - P[2].Y
		dy2 := P[3].Y - P[2].Y

		z := dx1*dy1 - dy1*dx2
		g := (sx*dy2 - sy*dx2) / z
		h := (sy*dx1 - sx*dy1) / z

		H[0] = P[1].X - P[0].X + g*P[1].X
		H[1] = P[3].X - P[0].X + h*P[3].X
		H[2] = P[0].X
		H[3] = P[1].Y - P[0].Y + g*P[1].Y
		H[4] = P[3].Y - P[0].Y + h*P[3].Y
		H[5] = P[0].Y
		H[6] = g
		H[7] = h
		H[8] = 1.
	}
	return H
}

func MAT_Projective_mapping(u *float64, v *float64, H [9]float64) {
	x := (H[0]**u + H[1]**v + H[2]) / (H[6]**u + H[7]**v + 1.)
	y := (H[3]**u + H[4]**v + H[5]) / (H[6]**u + H[7]**v + 1.)
	*u = x
	*v = y
}

func ld_polynomial_evaluation(a [5]float64, na uint32, x float64) float64 {
	res := a[na]
	for i := int(na - 1); i >= 0; i-- {
		res = res*x + a[i]
	}
	return res
}

func ConstituerMatriceDistortion() [hsp][wsp]uint32 {
	var mire_w, mire_h uint32 = 209, 154
	var width, height uint32 = 2584, 1936
	//size := width * height
	var paspix uint32 = 2

	var tonneau [hsp][wsp]uint32
	for h := uint32(0); h < hsp; h++ {
		for w := uint32(0); w < wsp; w++ {
			tonneau[h][w] = 0xFFFFFFFF
		}
	}
	for i := uint32(0); i < hsp; i++ {
		for j := uint32(0); j < wsp; j++ {
			norme := math.Sqrt(math.Pow(float64(j-wsp/2), 2) + math.Pow(float64(i-hsp/2), 2))
			coef := ld_polynomial_evaluation(coeffs(), degres, norme)
			coef = 0.
			j2s := (float64(wsp/2) + float64(j-wsp/2)*coef) * float64(paspix)
			i2s := (float64(hsp/2) + float64(i-hsp/2)*coef) * float64(paspix)
			j3s := (float64(wsp/2) + float64(wsp/2-j)*coef) * float64(paspix)
			i3s := (float64(hsp/2) + float64(hsp/2-i)*coef) * float64(paspix)
			i2 := uint32(math.RoundToEven(i2s))
			j2 := uint32(math.RoundToEven(j2s))
			i3 := uint32(math.RoundToEven(i3s))
			j3 := uint32(math.RoundToEven(j3s))
			if !(0 <= i2 && i2 < height && 0 <= j2 && j2 < width) {
				continue
			}
			if !(0 <= i3 && i3 < height && 0 <= j3 && j3 < width) {
				continue
			}
			ofs := i2*width + j2
			ofs1 := i2*width + j3
			ofs2 := i3*width + j2
			ofs3 := i3*width + j3
			tonneau[i][j] = ofs
			tonneau[i][(wsp-1)-j] = ofs1
			tonneau[(hsp-1)-i][j] = ofs2
			tonneau[(hsp-1)-i][(wsp-1)-j] = ofs3
		}
	}
	var homographie [hsp][wsp]uint32
	H := MAT_GetPerspectiveTransform(coins())

	P0P1 := math.Round(Norme(coins()[0], coins()[1]))
	P3P2 := math.Round(Norme(coins()[2], coins()[3]))

	var rw uint32
	if P0P1+P3P2 != 0 {
		rw = uint32(P0P1+P3P2) / 2
	} else {
		rw = 1280
	}
	rw = min(rw, wsp)
	rh := uint32((rw * mire_h) / mire_w)

	ow, oh := (wsp-rw)/2, (hsp-rh)/2

	for i := oh - 63; i < rh+oh+64; i++ {
		for j := ow - 48; j < rw+ow+48; j++ {
			u, v := float64((j-ow)/rw), float64((i-oh)/rh)
			MAT_Projective_mapping(&u, &v, H)
			i2, j2 := uint32(math.Round(u)), uint32(math.Round(v))
			if 0 <= i2 && i2 < hsp && 0 <= j2 && j2 < wsp {
				if i < 0 || j < 0 {
					log.Fatal("Erreur", i, j)
				}
				homographie[i][j] = tonneau[i2][j2]
			}
		}
	}

	return homographie
}
