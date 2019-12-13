package game

const (
	OptionRed   = "red"
	OptionBlack = "black"
)

func Credit(a, b string, round int) (x int, y int) {
	if a == OptionRed {
		if b == OptionRed {
			x, y = -3, -3
		} else if b == OptionBlack {
			x, y = 5, -5
		}
	} else if a == OptionBlack {
		if b == OptionRed {
			x, y = -5, 5
		} else if b == OptionBlack {
			x, y = 3, 3
		}
	}
	if round == 2 {
		x, y = 2*x, 2*y
	} else if round == 4 {
		x, y = 3*x, 3*y
	}
	return
}
