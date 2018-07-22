package main

import (
	"fmt"
	"math"
	"os"
	"text/template"
)

type RectShapeBeamQues struct {
	Title     string
	B, H      float64
	D_s, D_ss float64 // as, a's
	C         Conc
	S         Steel
	Ass       float64
	M         float64
	GivenAss  bool
}

type RectShapeBeamSolve struct {
	Q                   RectShapeBeamQues
	H_0                 float64 //矩形截面尺寸及有效高度
	Fy, Fc              float64 // 纵筋屈服强度，混凝土轴心抗压强度
	Xi, X, Xi_b         float64
	M_1, M_2, M_b       float64 // M_1 = f'y A's (h_0 - a's)、M_2 = M - M_1，平衡状态弯矩
	As, Ass             float64 //计算所得受拉纵筋、受压纵筋面积As, As'
	DoubleD_ss          float64 //  2as'
	Alpha_s, Alpha_smax float64
	M_2Negative         bool // \alpha_s < 0?
	AsYield, AssYield   bool // 受拉纵筋屈服?、受压纵筋屈服?
	ResetX              bool
	ResetAss            bool // \alpha_s > \alpha_{s,max}?
}

func MakeRectShapeDoubleReinforcedBeamTex(texfile string) {
	file, err := os.Create(texfile)
	if err != nil {
		fmt.Printf("Can't open file: %s\n", texfile)
		return
	}
	defer file.Close()

	Ques := RectShapeBeamQues{
		B: 200, H: 500, D_s: 60, D_ss: 35, M: 330, C: "C40", S: "HRB400",
		GivenAss: true,
	}

	setOfAs := []float64{226.2, 941, 1964, 2454}
	Titles := []string{"双筋矩形截面梁题1", "双筋矩形截面梁题2", "双筋矩形截面梁题3", "双筋矩形截面梁题4"}

	for i, _ := range setOfAs {
		Ques.Ass = setOfAs[i]
		Ques.Title = Titles[i]
		solv := NewRectShapeBeamSolve(Ques)
		solv.Solve()
		if solv.Q.GivenAss == true {
			if solv.AsYield && solv.AssYield {
				tmpl := template.Must(template.ParseFiles("./RectShape.GivenAssAndAssYieldAsYield.template.tex"))
				tmpl.Execute(file, solv)
				continue
			}
			if solv.AsYield != true {
				tmpl := template.Must(template.ParseFiles("./RectShape.GivenAssAndAsNoYield.template.tex"))
				tmpl.Execute(file, solv)
				continue
			}
			if solv.AssYield != true {
				tmpl := template.Must(template.ParseFiles("./RectShape.GivenAssAndAssNoYield.template.tex"))
				tmpl.Execute(file, solv)
				continue
			}
		} else {
			continue
		}
	}
}

func NewRectShapeBeamSolve(Ques RectShapeBeamQues) *RectShapeBeamSolve {
	s := RectShapeBeamSolve{Q: Ques}
	s.AsYield = true
	s.AssYield = true
	s.ResetX = false
	s.ResetAss = false
	s.H_0 = s.Q.H - s.Q.D_s
	s.Fy = s.Q.S.Fy()
	s.Fc = s.Q.C.Fc()
	s.Xi_b = s.Q.S.Xi_b(s.Q.C)
	s.Alpha_smax = s.Xi_b * (1.0 - 0.5*s.Xi_b)
	s.DoubleD_ss = 2.0 * s.Q.D_ss
	s.M_b = s.Q.C.Alpha1() * s.Fc * s.Q.B * s.H_0 * s.H_0 * s.Alpha_smax / 1e6
	return &s
}

func (s *RectShapeBeamSolve) Solve() {
	if s.Q.GivenAss == true {
		s.SolveWithGivenAss()
	} else {
		s.SolveWithoutGivenAss()
	}
}

func (s *RectShapeBeamSolve) SolveWithoutGivenAss() {
}

func (p *RectShapeBeamSolve) SolveWithGivenAss() {
	p.M_1 = p.Fy * p.Q.Ass * (p.H_0 - p.Q.D_ss)
	p.M_2 = (p.Q.M*1e6 - p.M_1) / 1e6
	if p.M_2 < 0 {
		p.M_2Negative = true
		p.AssYield = false
		p.SolveForXLessThanDoubleD_ss()
		return
	}

	p.Alpha_s = p.M_2 * 1e6 / (p.Q.C.Alpha1() * p.Fc * p.Q.B * p.H_0 * p.H_0)
	if p.Alpha_s > p.Alpha_smax {
		p.AsYield = false
		p.SolveForXiEqualXi_b()
		return
	}

	p.Xi = 1.0 - math.Sqrt(1.0-2.0*p.Alpha_s)
	p.X = p.Xi * p.H_0
	if p.X < p.DoubleD_ss {
		p.AssYield = false
		p.SolveForXLessThanDoubleD_ss()
		return
	}
	p.Ass = p.Q.Ass
	p.As = p.Q.C.Alpha1()*p.Fc*p.Q.B*p.X/p.Fy + p.Ass
	return
}

func (p *RectShapeBeamSolve) SolveForXLessThanDoubleD_ss() {
	p.Ass = p.Q.Ass
	p.As = p.Q.M * 1e6 / (p.Fy * (p.H_0 - p.Q.D_ss))
}

func (p *RectShapeBeamSolve) SolveForXiEqualXi_b() {
	p.Ass = (p.Q.M*1e6 - p.M_b*1e6) / (p.Fy * (p.H_0 - p.Q.D_ss))
	p.As = (1.0*p.Fc*p.Q.B*p.H_0*p.Xi_b + p.Fy*p.Ass) / p.Fy
}
