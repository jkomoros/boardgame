func checkRunWon(runState []string) string {

	if len(runState) != dim {
		return Empty
	}

	targetToken := runState[0]

	if targetToken == Empty {
		return Empty
	}

	for _, token := range runState {
		if token != targetToken {
			return Empty
		}
	}

	return targetToken
}