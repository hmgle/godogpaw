package engine

func GetHistoryScore(history *HistoryTable, move MoveNG, color Color) Value {
	return history[color][FromSQ(move)][ToSQ(move)]
}
