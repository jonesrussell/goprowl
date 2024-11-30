package engine

import (
	"math"
)

// Ranker handles document scoring and ranking
type Ranker struct {
	k1 float64 // BM25 parameter
	b  float64 // BM25 parameter
}

// New creates a new Ranker with default BM25 parameters
func New() *Ranker {
	return &Ranker{
		k1: 1.2,
		b:  0.75,
	}
}

// Score calculates BM25 score for a document
func (r *Ranker) Score(tf int, docLength, avgDocLength float64,
	docsWithTerm, totalDocs int64) float64 {
	// BM25 scoring formula
	idf := math.Log(1 + (float64(totalDocs)-float64(docsWithTerm)+0.5)/
		(float64(docsWithTerm)+0.5))

	numerator := float64(tf) * (r.k1 + 1)
	denominator := float64(tf) + r.k1*(1-r.b+r.b*docLength/avgDocLength)

	return idf * numerator / denominator
}

// BoostScore applies custom boosting factors
func (r *Ranker) BoostScore(score float64, boostFactors map[string]float64) float64 {
	finalScore := score
	for _, boost := range boostFactors {
		finalScore *= boost
	}
	return finalScore
}
