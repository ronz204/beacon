package discount

func Apply(cents int) int {
	if cents > 10000 {
		return cents - cents/10
	}

	return cents
}
