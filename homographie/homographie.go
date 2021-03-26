package homographie

import (
	"math"
)

type Point struct {
	X, Y float64
}

const wsp, hsp uint32 = 1280, 960

func min(a, b uint32) uint32 {
	if a <= b {
		return a
	}
	return b
}

func Norme(p1, p2 Point) float64 {
	return math.Sqrt(math.Pow(p1.X-p2.X, 2) + math.Pow(p1.Y-p2.Y, 2))
}

func init() {

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

		z := dx1*dy2 - dy1*dx2
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

func MAT_Projective_mappingFloat(u *float64, v *float64, H [9]float64) {
	x := (H[0]**u + H[1]**v + H[2]) / (H[6]**u + H[7]**v + 1.)
	y := (H[3]**u + H[4]**v + H[5]) / (H[6]**u + H[7]**v + 1.)
	*u = x
	*v = y
}

func MAT_Projective_mappingInt(u *int32, v *int32, H [9]float64) {
	h := make([]int32, 9)
	for i := 0; i < 9; i++ {
		h[i] = int32(H[i])
	}
	x := (h[0]**u + h[1]**v + h[2]) / (h[6]**u + h[7]**v + 1.)
	y := (h[3]**u + h[4]**v + h[5]) / (h[6]**u + h[7]**v + 1.)
	*u = x
	*v = y
}

func ld_polynomial_evaluation(a [5]float64, na int, x float64) float64 {
	res := a[na]
	for i := na - 1; i >= 0; i-- {
		res = res*x + a[i]
	}
	return res
}

func ConstituerMatriceDistortion(coins [4]Point, degre int, coefs [5]float64) [hsp][wsp]uint32 {
	var mire_w, mire_h uint32 = 209, 154
	var width, height uint32 = 2584, 1936
	//size := width * height
	var paspix float64 = 2.

	var tonneau [hsp][wsp]uint32
	for h := uint32(0); h < hsp; h++ {
		for w := uint32(0); w < wsp; w++ {
			tonneau[h][w] = 0xFFFFFFFF
		}
	}
	var norme, coef, j2s, i2s, j3s, i3s float64
	var i2, j2, i3, j3 uint32
	for i := uint32(0); i < hsp/2; i++ {
		for j := uint32(0); j < wsp/2; j++ {
			//norme = math.Sqrt(math.Pow(float64(j-wsp/2), 2) + math.Pow(float64(i-hsp/2), 2))
			norme = math.Sqrt(float64((j-wsp/2)*(j-wsp/2) + (i-hsp/2)*(i-hsp/2)))
			coef = ld_polynomial_evaluation(coefs, degre, norme)
			j2s = (float64(wsp)/2. + float64(j) - float64(wsp)/2.*coef) * paspix
			i2s = (float64(hsp)/2. + float64(i) - float64(hsp)/2.*coef) * paspix
			j3s = (float64(wsp)/2. + float64(wsp)/2. - float64(j)*coef) * paspix
			i3s = (float64(hsp)/2. + float64(hsp)/2. - float64(i)*coef) * paspix
			i2 = uint32(math.RoundToEven(i2s))
			j2 = uint32(math.RoundToEven(j2s))
			i3 = uint32(math.RoundToEven(i3s))
			j3 = uint32(math.RoundToEven(j3s))
			//fmt.Println(i, j, i2, j2, i3, j3, j2s, i2s, j3s, i3s, height, width)
			if i2 >= height || j2 >= width {
				continue
			}
			if i3 >= height || j3 >= width {
				continue
			}
			ofs := i2*width + j2
			ofs1 := i2*width + j3
			ofs2 := i3*width + j2
			ofs3 := i3*width + j3
			tonneau[i][j] = ofs
			tonneau[i][(wsp-1)-j] = ofs1
			tonneau[hsp-1-i][j] = ofs2
			tonneau[hsp-1-i][wsp-1-j] = ofs3
		}
	}
	var homographie [hsp][wsp]uint32
	for h := uint32(0); h < hsp; h++ {
		for w := uint32(0); w < wsp; w++ {
			homographie[h][w] = 0xFFFFFFFF
		}
	}
	H := MAT_GetPerspectiveTransform(coins)

	P0P1 := math.Round(Norme(coins[0], coins[1]))
	P3P2 := math.Round(Norme(coins[2], coins[3]))

	var rw, rh, ow, oh uint32
	if P0P1+P3P2 != 0 {
		rw = uint32(P0P1+P3P2) / 2
	} else {
		rw = wsp
	}
	rw = min(rw, wsp)
	rh = uint32((rw * mire_h) / mire_w)
	ow, oh = (wsp-rw)/2, (hsp-rh)/2

	for i := oh - min(oh, 64); i < min(hsp, rh+oh+64); i++ {
		for j := ow - min(ow, 48); j < min(wsp, rw+ow+48); j++ {
			//u, v := (float64(j)-float64(ow))/float64(rw), (float64(i)-float64(oh))/float64(rh)
			u, v := int32(10000*(j-ow)/rw), int32(10000*(i-oh)/rh)
			MAT_Projective_mappingInt(&u, &v, H)
			//i2, j2 := uint32(math.Round(u)), uint32(math.Round(v))
			i2, j2 := u, v
			if 0 <= i2 && i2 < int32(hsp) && 0 <= j2 && j2 < int32(wsp) {
				homographie[i][j] = tonneau[i2][j2]
			}
		}
	}

	return homographie
}
