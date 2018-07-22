package main

type Conc string

func (c *Conc) Fc() (val float64) {
	switch *c {
	case "C20":
		val = 9.6
	case "C25":
		val = 11.9
	case "C30":
		val = 14.3
	case "C35":
		val = 16.7
	case "C40":
		val = 19.1
	case "C45":
		val = 21.1
	case "C50":
		val = 23.1
	}
	return val
}

func (c *Conc) Ft() (ft float64) {
	switch *c {
	case "C20":
		ft = 1.1
	case "C25":
		ft = 1.27
	case "C30":
		ft = 1.43
	case "C35":
		ft = 1.57
	case "C40":
		ft = 1.71
	case "C45":
		ft = 1.80
	case "C50":
		ft = 1.89
	}
	return ft
}

func (c *Conc) Alpha1() (val float64) {
	switch *c {
	case "C80":
		val = 0.94
	case "C75":
		val = 0.95
	case "C70":
		val = 0.96
	case "C65":
		val = 0.97
	case "C60":
		val = 0.98
	case "C55":
		val = 0.99
	default:
		val = 1.0
	}
	return val
}

func (c *Conc) Xi_b(s Steel) float64 {
	return (&s).Xi_b(*c)
}

type Steel string

func (s *Steel) Fy() (val float64) {
	switch *s {
	case "HRB500":
		val = 435
	case "HRB400":
		val = 360
	case "HRB335":
		val = 300
	case "HPB300":
		val = 270

	}
	return val
}

func (s *Steel) Fyy() (val float64) {
	switch *s {
	case "HRB500":
		val = 410
	case "HRB400":
		val = 360
	case "HRB335":
		val = 300
	case "HPB300":
		val = 270

	}
	return val
}

func (s *Steel) Xi_b(c Conc) (val float64) {
	switch *s {
	case "HRB500":
		val = 0.482
	case "HRB400":
		val = 0.518
	case "HRB335":
		val = 0.55
	case "HPB300":
		val = 0.576

	}
	return val
}
