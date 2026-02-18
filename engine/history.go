package engine

const maxHistory Value = 8192

func GetHistoryScore(history *HistoryTable, move MoveNG, color Color) Value {
	return history[color][FromSQ(move)][ToSQ(move)]
}

// UpdateHistory adds a bonus to the history table with saturation to prevent overflow.
// Uses the formula: bonus - history * |bonus| / maxHistory
func UpdateHistory(history *HistoryTable, move MoveNG, color Color, bonus int32) {
	from := FromSQ(move)
	to := ToSQ(move)
	entry := &history[color][from][to]
	// Gravity formula to prevent unbounded growth
	*entry += bonus - *entry*absVal(bonus)/maxHistory
}

func absVal(x Value) Value {
	if x < 0 {
		return -x
	}
	return x
}
