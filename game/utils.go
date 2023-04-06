package game

import (
	"math"

	"github.com/chehsunliu/poker"
	"gonum.org/v1/gonum/stat/combin"
)

// getBestHand returns the best possible and and a score made from community cards and hole cards
func getBestHand(holeCards []poker.Card, communityCards []poker.Card) ([]poker.Card, int, string) {
	combinedCards := append(holeCards, communityCards...)
	bestHand := make([]poker.Card, 5)
	bestScore := int32(math.MaxInt32)
	var bestRank string
	permgen := combin.NewPermutationGenerator(7, 5)
	currentHand := make([]poker.Card, 5)

	for permgen.Next() {
		hand := permgen.Permutation(nil)
		for i := 0; i < 5; i++ {
			currentHand[i] = combinedCards[hand[i]]
		}

		score := poker.Evaluate(currentHand)
		if score < bestScore {
			bestScore = score
			copy(bestHand, currentHand)
			bestRank = poker.RankString(score)
		}
	}

	return bestHand, int(bestScore), bestRank
}
