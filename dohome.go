package main

import (
	"fmt"
	"math"
	"os"
	"text/template"
)

type Conc string

func (c *Conc) fc() (val float64) {
	switch *c {
	case "C30":
		val = 14.3
	case "C40":
		val = 19.1
	}
	return val
}

func (c *Conc) alpha1() (val float64) {
	switch *c {
	case "C50":
		val = 0.9
	default:
		val = 1.0
	}
	return val
}

type Steel string

func (s *Steel) fy() (val float64) {
	switch *s {
	case "HRB400":
		val = 360
	case "HRB335":
		val = 300

	}
	return val
}

type Problem struct {
	Title              string  //题目描述
	B, H, H_0          float64 //矩形截面尺寸及有效高度
	M, M_1, M_2, M_b   float64 //弯矩、受压筋分担弯矩、受压混凝土分担弯矩，平衡状态弯矩
	C                  Conc    // 混凝土标号
	S                  Steel   // 钢筋标号
	Fy, Fc             float64 // 纵筋屈服强度，混凝土轴心抗压强度
	QASc               float64 //题目给定的受压纵筋面积
	ASt, ASc           float64 //受拉纵筋、受压纵筋面积As, As'
	As, Ass, DoubleAss float64 // 纵筋合力点到边缘距离as, as'
	Xi, X, Xi_b        float64
	Alphas, Alphas_max float64
	AlphasGtAlphasmax  bool // \alpha_s > \alphas_{max}
	AlphasNegative     bool // \alpha_s < 0
	ResetX             bool
}

func (p *Problem) Prepare() {
	p.ResetX = false
	p.AlphasGtAlphasmax = false
	p.AlphasNegative = false
	p.H_0 = p.H - p.As
	p.DoubleAss = 2.0 * p.Ass
	p.Fy = p.S.fy()
	p.Fc = p.C.fc()
	p.Alphas_max = p.Xi_b * (1.0 - 0.5*p.Xi_b)
	p.M_b = 1.0 * p.Fc * p.B * p.H_0 * p.H_0 * p.Alphas_max / 1e6
}

func (p *Problem) SolveWithGivenAss() {
	p.Prepare()

	p.M_1 = p.S.fy() * p.QASc * (p.H_0 - p.Ass)
	p.M_2 = (p.M*1e6 - p.M_1) / 1e6
	p.Alphas = p.M_2 * 1e6 / (p.C.alpha1() * p.Fc * p.B * p.H_0 * p.H_0)

	if p.Alphas > p.Alphas_max {
		p.AlphasGtAlphasmax = true
		p.SolveForXiEqXi_b()
		return
	}
	//	if p.Alphas < 0 {
	//		p.AlphasNegative = true
	//		p.SolveForXLtDoubleAss()
	//		return
	//	}
	p.ASc = p.QASc
	p.Xi = 1.0 - math.Sqrt(1.0-2.0*p.Alphas)
	p.X = p.Xi * p.H_0
	if p.X < 2.0*p.Ass {
		p.SolveForXLtDoubleAss()
		return
	}
	p.ASt = p.C.alpha1()*p.C.fc()*p.B*p.X/p.S.fy() + p.ASc
	return
}

func (p *Problem) SolveForXLtDoubleAss() {
	p.ResetX = true
	p.ASt = p.M * 1e6 / (p.S.fy() * (p.H_0 - p.Ass))
}

func (p *Problem) SolveForXiEqXi_b() {
	p.ASc = (p.M*1e6 - p.M_b*1e6) / (p.Fy * (p.H_0 - p.Ass))
	p.ASt = (1.0*p.Fc*p.B*p.H_0*p.Xi_b + p.Fy*p.ASc) / p.Fy
}

func main() {
	file, err := os.Create("hm2.tex")
	if err != nil {
		fmt.Printf("Can't open file: %s\n", "hm2.tex")
		return
	}
	defer file.Close()

	problem := Problem{
		B: 200, H: 500, As: 60, Ass: 35, M: 330, C: "C40", S: "HRB400", Xi_b: 0.55,
	}
	As := []float64{226.2, 941, 2454}
	Ds := []string{"题1", "题2", "题3"}

	for i, _ := range As {
		problem.QASc = As[i]
		problem.Title = Ds[i]

		problem.SolveWithGivenAss()
		tmpl := template.Must(template.ParseFiles("DoubleReinforeBeam.template.tex"))
		tmpl.Execute(file, problem)
	}
}
