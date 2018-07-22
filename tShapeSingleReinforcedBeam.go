package main

import (
	"fmt"
	"math"
	"os"
	"text/template"
)

type TShapeQuestion struct {
	Title    string
	B, H     float64 // b\times h
	Bff, Hff float64 // b'_f, h'_f
	C        Conc
	S        Steel
	M        float64
	A        float64 // a_s, h_0 = h - a_s
}

type TShapeSolve struct {
	Title            string
	B, H             float64
	Bff, Hff         float64
	C                Conc
	S                Steel
	A                float64 // a_s
	H_0              float64 // h_0, h_0 = h - a_s
	F_c, F_y         float64 // f_c, f_y
	Alpha_1          float64
	M                float64
	M_xEqHff         float64 // M correspond to x = h'_f,
	M_1              float64 // M_1 = \alpha_1 fc (b'_f -b) h'_f (h_0 - h'_f/2)
	M_2              float64 // M_2 = M - M_1
	Alpha_s, Xi      float64 // \alpha_s, \xi
	X                float64
	Xi_b, Alpha_smax float64 // 界限受压区高度
	A_s              float64 //纵向受拉钢筋面积
}

func NewTShapeSolve(q *TShapeQuestion) *TShapeSolve {
	s := TShapeSolve{
		B: q.B, H: q.H, Bff: q.Bff, Hff: q.Hff,
		A: q.A, M: q.M, C: q.C, S: q.S,
		Title: q.Title,
	}

	s.F_c = q.C.Fc()
	s.F_y = q.S.Fy()
	s.H_0 = q.H - q.A
	s.Xi_b = q.S.Xi_b(q.C)
	s.Alpha_1 = q.C.Alpha1()
	s.Alpha_smax = s.Xi_b * (1.0 - 0.5*s.Xi_b)
	s.M_xEqHff = 1.0 * s.F_c * s.Bff * s.Hff * (s.H_0 - 0.5*s.Hff) / 1e6
	return &s
}

func (s *TShapeSolve) SolveForSecondShape() {
	s.M_1 = 1.0 * s.F_c * (s.Bff - s.B) * s.Hff * (s.H_0 - 0.5*s.Hff) / 1e6
	s.M_2 = s.M - s.M_1
	s.Alpha_s = s.M_2 * 1e6 / (1.0 * s.F_c * s.B * s.H_0 * s.H_0)
	if s.Alpha_s > s.Alpha_smax { // 双筋T形截面梁未处理
		return
	} else { // 受拉纵筋屈服，继续计算
		s.Xi = 1.0 - math.Sqrt(1.0-2.0*s.Alpha_s)
		s.X = s.Xi * s.H_0
		s.A_s = (s.Alpha_1*s.F_c*(s.Bff-s.B)*s.Hff + s.Alpha_1*s.F_c*s.B*s.X) / s.F_y
		return
	}
}

func (s *TShapeSolve) SolveForFirstShape() {
	s.Alpha_s = s.M * 1e6 / (1.0 * s.F_c * s.Bff * s.H_0 * s.H_0)
	if s.Alpha_s > s.Alpha_smax {
		return
	}
	s.Xi = 1.0 - math.Sqrt(1.0-2.0*s.Alpha_s)
	s.X = s.Xi * s.H_0
	s.A_s = (s.Alpha_1 * s.F_c * s.Bff * s.X) / s.F_y
}

func MakeTShapeReinforcedBeamTex(texfile string) {
	file, err := os.Create(texfile)
	if err != nil {
		fmt.Printf("Can't open file: %s\n", texfile)
		return
	}
	Ques := TShapeQuestion{
		B: 250, H: 700, Bff: 600, Hff: 100,
		A: 60, M: 550, C: "C30", S: "HRB400",
	}
	Ms := []float64{550, 100}
	Ts := []string{"题1", "题2"}
	for i, M := range Ms {
		Ques.M = M
		Ques.Title = Ts[i]
		Solv := NewTShapeSolve(&Ques)
		if Solv.M > Solv.M_xEqHff {
			Solv.SolveForSecondShape()
			tmpl := template.Must(template.ParseFiles("Tshape.SecondShape.template.tex"))
			tmpl.Execute(file, Solv)
		} else {
			Solv.SolveForFirstShape()
			tmpl := template.Must(template.ParseFiles("Tshape.FirstShape.template.tex"))
			tmpl.Execute(file, Solv)
		}
	}
}
