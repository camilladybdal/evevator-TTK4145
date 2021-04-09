package costfunc 

/* Based on minimal movement. Should it be beased on minimal time instead? */

//need to know: direction of elevator, what floor the elevator is in
func costfunction(neworder <-chan Order, localelevstate){
	cost := localelevstate.floor - order.floor
	if cost == 0 {
		return cost 
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
	return cost 
}



