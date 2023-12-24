package controllers

type Gears struct {
}

func (c *Controller) ApplyGearTransform() {
	//determine current gear
	//support 6 gears + Reverse (rawInput 4-9 with 10 as reverse)
	// currentGear := 0 //-1 reverse, 0 is neutral, 1-6 is gear
	// for i := 4; i <= 10; i++ {
	// 	if c.rawInputs[i] > sbus.MidValue {
	// 		if i == 10 {
	// 			currentGear = -1
	// 		} else {
	// 			currentGear = i - 3
	// 		}
	// 	}
	// }
}
