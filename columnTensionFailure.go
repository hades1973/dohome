package main

import (
	"fmt"
	"math"
	"os"
	"text/template"
)

type RectShapeColumnQues struct {
	Title       string
	B, H        float64
	D_s, D_ss   float64 // as, a's
	C           Conc
	S           Steel
	Ass         float64
	GivenAss    bool
	M_t, M_b, N float64
	L_c         float64
	Sym         bool
}

type RectShapeColumnSolve struct {
	Q           RectShapeColumnQues
	H_0         float64 //矩形截面尺寸及有效高度
	Fy, Fyy, Fc float64 // 纵筋受拉屈服强度，受压屈服强度、混凝土轴心抗压强度
	DoubleD_ss  float64 //  2as'

	M_1, M_2                        float64 // M_1 = min(M_t, M_b), M_2=max(M_1, M_2)平衡状态弯矩
	M_1DividedByM_2                 float64
	M_1DividedByM_2GreatThanDotNine bool
	I, Lambda, LambdaLimit          float64 // i = h/sqrt(12), \lambda=L_c/I
	LambdaGreatThanLambdaLimit      bool
	NDividedByFcA                   float64
	NDividedByFcAGreatThanDotNine   bool

	Zeta_c float64
	Eta_ns float64
	Cm     float64
	M      float64

	E_0, E_a, E_i, E, Ee float64 // e_a, e_i, e, e'

	Xi, X, Xi_b, X_b    float64
	XiGreatThanXi_b     bool
	XLessThanDoubleD_ss bool

	As, Ass                         float64 //计算所得受拉纵筋、受压纵筋面积As, As'
	AsPlusAss                       float64
	OneSideRho_min, TwoSidesRho_min float64
	As_min, AsPlusAss_min           float64
	AsGreatThanAs_min               bool
	AssGreatThanAs_min              bool
	AsPlusAssGreatThanAsPlusAss_min bool
}

func MakeRectColumnTex(texfile string) {
	file, err := os.Create(texfile)
	if err != nil {
		fmt.Printf("Can't open file: %s\n", texfile)
		return
	}
	defer file.Close()

	Ques := RectShapeColumnQues{
		B: 300, H: 400, D_s: 40, D_ss: 40, L_c: 3.5,
		M_t: -200, M_b: 70, N: 750,
		C: "C30", S: "HRB400",
		GivenAss: false,
		Sym:      true,
	}

	Titles := []string{"对称配筋柱正截面计算题例1", "对称配筋柱正截面计算题2", "对称配筋柱正截面计算题3", "对称配筋柱计算题4"}
	M_ts := []float64{-200, -200, 200, 200}
	M_bs := []float64{70, 70, -70, -70}
	Ns := []float64{750, 750, 250, 2500}
	Lcs := []float64{3.5, 4.5, 3.5, 3.5}

	for i, _ := range Titles {
		Ques.Title = Titles[i]
		Ques.M_b = M_bs[i]
		Ques.M_t = M_ts[i]
		Ques.N = Ns[i]
		Ques.L_c = Lcs[i]
		solv := NewRectShapeColumnSolve(Ques)
		solv.Solve()
		tmpl := template.Must(template.ParseFiles("./Column.template.tex"))
		tmpl.Execute(file, solv)
	}
}

func NewRectShapeColumnSolve(Ques RectShapeColumnQues) *RectShapeColumnSolve {
	s := RectShapeColumnSolve{Q: Ques}
	s.Fy = s.Q.S.Fy()
	s.Fyy = s.Q.S.Fyy()
	s.Fc = s.Q.C.Fc()
	s.DoubleD_ss = 2.0 * s.Q.D_ss
	s.H_0 = s.Q.H - s.Q.D_s
	s.Xi_b = s.Q.C.Xi_b(s.Q.S)
	s.X_b = s.Xi_b * s.H_0

	// caculate M_1, M_2
	s.M_1 = math.Min(math.Abs(s.Q.M_t), math.Abs(s.Q.M_b))
	s.M_2 = math.Max(math.Abs(s.Q.M_t), math.Abs(s.Q.M_b))
	if s.Q.M_t*s.Q.M_b < 0 {
		s.M_1 *= -1
	}
	s.M_1DividedByM_2 = s.M_1 / s.M_2
	if s.M_1DividedByM_2 > 0.9 {
		s.M_1DividedByM_2GreatThanDotNine = true
	} else {
		s.M_1DividedByM_2GreatThanDotNine = false
	}

	// 计算长细比相关参数
	s.I = s.Q.H / math.Sqrt(12)
	s.Lambda = s.Q.L_c * 1e3 / s.I
	s.LambdaLimit = 34 - 12*(s.M_1DividedByM_2)
	if s.Lambda > s.LambdaLimit {
		s.LambdaGreatThanLambdaLimit = true
	} else {
		s.LambdaGreatThanLambdaLimit = false
	}

	// 计算轴压比相关参数
	s.NDividedByFcA = s.Q.N * 1e3 / (s.Fc * s.Q.B * s.Q.H)
	if s.NDividedByFcA > 0.9 {
		s.NDividedByFcAGreatThanDotNine = true
	} else {
		s.NDividedByFcAGreatThanDotNine = false
	}

	// 附加偏心距
	s.E_a = math.Max(20, s.Q.H/30)

	// 二阶效应
	if !s.NDividedByFcAGreatThanDotNine && !s.M_1DividedByM_2GreatThanDotNine && !s.LambdaGreatThanLambdaLimit {
		s.Eta_ns = 1.0
		s.Cm = 1.0
	} else {
		s.Zeta_c = math.Min(1.0, 0.5*s.NDividedByFcA)
		s.Eta_ns = 1.0 + s.H_0*math.Pow(s.Q.L_c/s.Q.H*1e3, 2.0)*s.Zeta_c/(1300.0*(s.M_2*1e3/s.Q.N+s.E_a))
		s.Cm = math.Max(0.7, 0.7+0.3*s.M_1DividedByM_2)
	}
	s.M = math.Max(1.0, s.Cm*s.Eta_ns) * s.M_2

	// 偏心距及外力臂
	s.E_0 = s.M / s.Q.N * 1e3
	s.E_i = s.E_0 + s.E_a
	s.E = s.E_i + 0.5*s.Q.H - s.Q.D_s
	s.Ee = s.E_i - 0.5*s.Q.H + s.Q.D_ss

	// 受压区高度
	s.X = s.Q.N * 1e3 / (s.Q.C.Alpha1() * s.Fc * s.Q.B)
	if s.X > s.X_b {
		s.XiGreatThanXi_b = true
	} else {
		s.XiGreatThanXi_b = false
	}
	if s.X < s.DoubleD_ss {
		s.XLessThanDoubleD_ss = true
	} else {
		s.XLessThanDoubleD_ss = false
	}

	// 纵筋面积计算
	if s.XLessThanDoubleD_ss {
		s.Ass = s.Q.N * 1e3 * s.Ee / (s.Fy * (s.H_0 - s.Q.D_ss))
		s.As = s.Ass

	} else if s.XiGreatThanXi_b {
		s.Ee = -s.E_i + 0.5*s.Q.H - s.Q.D_ss
		fracOne := s.Q.N*1e3 - s.Q.C.Alpha1()*s.Fc*s.Q.B*s.X_b
		fracTwo := s.Q.N*1e3*s.E - 0.43*s.Q.C.Alpha1()*s.Fc*s.Q.B*math.Pow(s.H_0, 2)
		fracTwo /= (0.8 - s.Xi_b) * (s.H_0 - s.Q.D_ss)
		fracTwo += s.Q.C.Alpha1() * s.Fc * s.Q.B * s.H_0
		s.Xi = (fracOne)/(fracTwo) + s.Xi_b
		s.As = s.Q.N*1e3*s.E - s.Xi*(1-0.5*s.Xi)*s.Q.C.Alpha1()*s.Fc*s.Q.B*math.Pow(s.H_0, 2)
		s.As /= s.Fyy * (s.H_0 - s.Q.D_ss)
		s.Ass = s.As
	} else {
		s.Ass = (s.Q.N*1e3*s.E - s.Q.C.Alpha1()*s.Fc*s.Q.B*s.X*(s.H_0-0.5*s.X)) / (s.Fyy * (s.H_0 - s.Q.D_ss))
		s.As = s.Ass
	}

	// 最小配筋率验算
	s.OneSideRho_min = 0.2
	if s.Q.S == "HRB500" {
		s.TwoSidesRho_min = 0.5
	} else if s.Q.S == "HRB400" {
		s.TwoSidesRho_min = 0.55
	} else {
		s.TwoSidesRho_min = 0.6
	}
	s.As_min = s.OneSideRho_min / 100 * s.Q.B * s.Q.H
	s.AsPlusAss_min = s.TwoSidesRho_min / 100 * s.Q.B * s.Q.H
	if (s.As + s.Ass) > s.AsPlusAss_min {
		s.AsPlusAssGreatThanAsPlusAss_min = true
	} else {
		s.AsPlusAssGreatThanAsPlusAss_min = false
	}
	if s.As > s.As_min {
		s.AsGreatThanAs_min = true
	} else {
		s.AsGreatThanAs_min = false
	}
	return &s
}

func (s *RectShapeColumnSolve) Solve() {
}
