package engine

func UpdateHistory(history *HistoryTable, move MoveNG, color Color, delta Value) {
	from := FromSQ(move)
	to := ToSQ(move)
	entry := history[color][from][to]

	// Ensure the update value is within [-400, 400]
	delta = max(-400, min(400, delta))

	// Ensure the new value is within [-16384, 16384]
	history[color][from][to] += 32*delta - entry*(Value(abs(int(delta))))/512
}

func GetHistoryScore(history *HistoryTable, move MoveNG, color Color) Value {
	return history[color][FromSQ(move)][ToSQ(move)]
}
