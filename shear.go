package main

import (
	"fmt"
	"os"
	"text/template"
)

type ShearQuestion struct {
	Title              string
	B, H               float64 // b\times h
	C                  Conc
	S                  Steel
	SV                 Steel
	CaculateSelfWeight bool
	Lgk, Lqk           float64 // Load gk, qk
	L0, Ln             float64 // 计算跨度、净跨
	A                  float64 // a_s, h_0 = h - a_s
}

type ShearSolve struct {
	Q                       ShearQuestion
	H_0                     float64 // h_0, h_0 = h - a_s
	F_t, F_c, F_yv          float64 // t_t, f_c, f_{yv}
	Alpha_1                 float64
	Alpha_c                 float64
	Lgk_self                float64 // 梁自重
	Lp                      float64 // p = 1.2g_k + 1.4q_k
	Rho_sv, Rho_svmin       float64 // 配箍率、最小配箍率：\rho_v, \rho_{v,min} 精度，x.xxx%
	Rho_svLessThanRho_svmin bool
	V                       float64 // 设计剪力
	SV_d, SV_area1, SV_n    float64 // 箍筋直径、单肢面积，肢数
	SV_s, SV_smax           float64 // 箍筋间距, 规范规定箍筋最大间距
	SV_sGreatThanSV_smax    bool
	VuLimit                 float64 // 0.25\beta_c f_c b h_0
	VGreatThanVuLimit       bool    // V > 0.25\beta_c f_c b h_0
}

func NewShearSolve(q *ShearQuestion) *ShearSolve {
	s := ShearSolve{
		Q: *q,
	}

	s.F_t = q.C.Ft()
	s.F_c = q.C.Fc()
	s.F_yv = q.SV.Fy()
	s.H_0 = q.H - q.A
	s.Alpha_1 = q.C.Alpha1()
	s.Alpha_c = 1.0
	if q.CaculateSelfWeight == true {
		s.Lgk_self = 25.0 * q.B * q.H / (1000.0 * 1000.0)
	} else {
		s.Lgk_self = 0.0
	}
	s.Lp = 1.2*(q.Lgk+s.Lgk_self) + 1.4*q.Lqk

	return &s
}

func (s *ShearSolve) Solve() {
	s.V = 0.5 * s.Lp * s.Q.Ln
	s.Rho_sv = (s.V*1e3 - 0.7*s.F_t*s.Q.B*s.H_0) / (s.F_yv * s.Q.B * s.H_0) * 100
	s.Rho_svmin = 0.24 * s.F_t / s.F_yv * 100
	s.SV_n = 2
	s.SV_d = 8
	s.SV_area1 = 0.25 * 3.14159 * s.SV_d * s.SV_d

	s.VuLimit = 0.25 * s.F_c * s.Q.B * s.H_0 / 1e3
	s.VGreatThanVuLimit = false
	if s.V > s.VuLimit {
		s.VGreatThanVuLimit = true
		return
	}

	rho := s.Rho_sv
	s.Rho_svLessThanRho_svmin = false
	if rho < s.Rho_svmin {
		s.Rho_svLessThanRho_svmin = true
		rho = s.Rho_svmin
	}
	s.SV_s = (s.SV_area1 * s.SV_n * 100.0) / (s.Q.B * rho)

	if s.Rho_sv > 0 {
		switch {
		case s.Q.H < 150:
			s.SV_smax = s.Q.H
		case s.Q.H > 150 && !(s.Q.H > 300):
			s.SV_smax = 150
		case s.Q.H > 300 && !(s.Q.H > 500):
			s.SV_smax = 200
		case s.Q.H > 500 && !(s.Q.H > 800):
			s.SV_smax = 250
		case s.Q.H > 800:
			s.SV_smax = 350
		}
	} else {
		switch {
		case s.Q.H < 150:
			s.SV_smax = s.Q.H
		case s.Q.H > 150 && !(s.Q.H > 300):
			s.SV_smax = 200
		case s.Q.H > 300 && !(s.Q.H > 500):
			s.SV_smax = 300
		case s.Q.H > 500 && !(s.Q.H > 800):
			s.SV_smax = 350
		case s.Q.H > 800:
			s.SV_smax = 400
		}
	}
	if s.SV_s > s.SV_smax {
		s.SV_sGreatThanSV_smax = true
	} else {
		s.SV_sGreatThanSV_smax = false
	}
}

func MakeShearTex(texfile string) {
	file, err := os.Create(texfile)
	if err != nil {
		fmt.Printf("Can't open file: %s\n", texfile)
		return
	}
	defer file.Close()

	Ques := ShearQuestion{
		B: 200, H: 450,
		A: 35, C: "C40", S: "HRB400", SV: "HPB300",
	}
	gks := []float64{15.6, 9.5}
	qks := []float64{10.7, 8}
	l0s := []float64{6.0, 5.0}
	lns := []float64{5.76, 4.76}
	titles := []string{"受剪承载力题1", "受剪承载力题2"}
	cakWeights := []bool{false, true}
	for i := range gks {
		Ques.Lgk = gks[i]
		Ques.Lqk = qks[i]
		Ques.L0 = l0s[i]
		Ques.Ln = lns[i]
		Ques.Title = titles[i]
		Ques.CaculateSelfWeight = cakWeights[i]
		Solv := NewShearSolve(&Ques)
		if Ques.CaculateSelfWeight {
			Solv.Lp = 1.2*(Ques.Lgk+Solv.Lgk_self) + 1.4*Ques.Lqk
		} else {
			Solv.Lp = 1.2*Ques.Lgk + 1.4*Ques.Lqk
		}
		Solv.Solve()
		tmpl := template.Must(template.ParseFiles("Shear.template.tex"))
		tmpl.Execute(file, Solv)
	}
}
