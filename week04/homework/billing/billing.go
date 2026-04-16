package main

import "fmt"

type Valus struct {
	FirstLevel       float64
	FirstLevelPrice  float64
	SecondLevel      float64
	SecondLevelPrice float64
	ThirdLevel       float64
	ThirdLevelPrice  float64
}

const RealVal Valus = Valus{
	FirstLevel:       0,
	FirstLevelPrice:  0.5,
	SecondLevel:      200,
	SecondLevelPrice: 0.8,
	ThirdLevel:       400,
	ThirdLevelPrice:  1.2,
}

func init() {
	fmt.Println(RealVal)
}
func main() {
	fmt.Println("main")
}
