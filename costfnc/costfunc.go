package costfunc 

//need to know: direction of elevator, what floor the elevator is in
func costfunction(neworder <-chan Order, localelevstate){
	cost := localelevstate.floor - order.floor
	if cost == 0 {
		return cost //(broadcast dette)
	}
	if cost > 0 {
		if localelevstate.motordir != -1 {
			cost = cost + numFloors
		}
	}
	if cost < 0 {
		cost = abs(cost)
		if localelevstate.motordir != 1 {
			cost = cost + numFloors
		}
	}
	return cost //(broadcast dette)

}

func minimumcost(order <-chan Order, costotherelev){
	if mycost < cost1 && mycost < cost2 {
		return 1
	}
	return 0
}
