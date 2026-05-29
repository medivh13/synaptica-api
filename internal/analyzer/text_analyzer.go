package analyzer

import (
	"math"
	"regexp"
	"strings"
)

type TextAnalyzer interface {
	Analyze(text string) AnalysisOutput
}

type AnalysisOutput struct {
	SentimentLabel     string
	SentimentScore     float64
	EmotionLabel       string
	EmotionScore       float64
	ToxicityScore      float64
	CognitiveLoadScore float64
	Language           string
}

type RuleBasedTextAnalyzer struct{}

var nonWordPattern = regexp.MustCompile(`[^a-z0-9]+`)

func NewRuleBasedTextAnalyzer() TextAnalyzer {
	return RuleBasedTextAnalyzer{}
}

func (RuleBasedTextAnalyzer) Analyze(text string) AnalysisOutput {
	normalized := normalizeText(text)
	words := tokenize(normalized)

	sentimentLabel, sentimentScore := analyzeSentiment(words)
	emotionLabel, emotionScore := analyzeEmotion(words)

	return AnalysisOutput{
		SentimentLabel:     sentimentLabel,
		SentimentScore:     sentimentScore,
		EmotionLabel:       emotionLabel,
		EmotionScore:       emotionScore,
		ToxicityScore:      analyzeToxicity(words),
		CognitiveLoadScore: analyzeCognitiveLoad(normalized, words),
		Language:           "en",
	}
}

func normalizeText(text string) string {
	return strings.Join(strings.Fields(strings.ToLower(text)), " ")
}

func tokenize(text string) []string {
	cleaned := strings.TrimSpace(nonWordPattern.ReplaceAllString(text, " "))
	if cleaned == "" {
		return nil
	}
	return strings.Fields(cleaned)
}

func analyzeSentiment(words []string) (string, float64) {
	positiveWords := map[string]struct{}{
		"good": {}, "great": {}, "excellent": {}, "amazing": {}, "love": {},
		"useful": {}, "promising": {}, "successful": {}, "improve": {},
		"breakthrough": {}, "efficient": {},
	}
	negativeWords := map[string]struct{}{
		"bad": {}, "worse": {}, "terrible": {}, "fail": {}, "failure": {},
		"broken": {}, "problem": {}, "risk": {}, "danger": {}, "scam": {},
		"collapse": {}, "crisis": {}, "threat": {}, "loss": {},
	}

	positiveCount := countWords(words, positiveWords)
	negativeCount := countWords(words, negativeWords)
	totalMatched := max(positiveCount+negativeCount, 1)
	score := float64(positiveCount-negativeCount) / float64(totalMatched)

	switch {
	case score > 0.15:
		return "positive", score
	case score < -0.15:
		return "negative", score
	default:
		return "neutral", score
	}
}

func analyzeEmotion(words []string) (string, float64) {
	groups := map[string]map[string]struct{}{
		"fear": {
			"fear": {}, "afraid": {}, "scared": {}, "worry": {}, "worried": {},
			"anxiety": {}, "anxious": {}, "risk": {}, "danger": {}, "threat": {},
			"unsafe": {}, "uncertain": {},
		},
		"anger": {
			"angry": {}, "anger": {}, "hate": {}, "stupid": {}, "ridiculous": {},
			"broken": {}, "corrupt": {}, "exploit": {}, "exploited": {},
			"exploitative": {}, "unfair": {},
		},
		"joy": {
			"love": {}, "great": {}, "excellent": {}, "amazing": {}, "happy": {},
			"excited": {}, "promising": {}, "useful": {},
		},
		"sadness": {
			"sad": {}, "loss": {}, "lost": {}, "lonely": {}, "depressed": {},
			"depression": {}, "hopeless": {}, "grief": {},
		},
		"disgust": {
			"disgust": {}, "disgusting": {}, "gross": {}, "nasty": {}, "shameful": {},
		},
		"surprise": {
			"surprise": {}, "surprising": {}, "unexpected": {}, "shocking": {}, "wow": {},
		},
	}

	winningLabel := "neutral"
	winningCount := 0
	totalMatches := 0
	for label, group := range groups {
		count := countWords(words, group)
		totalMatches += count
		if count > winningCount {
			winningLabel = label
			winningCount = count
		}
	}

	if totalMatches == 0 {
		return "neutral", 0
	}

	return winningLabel, float64(winningCount) / float64(totalMatches)
}

func analyzeToxicity(words []string) float64 {
	toxicWords := map[string]struct{}{
		"idiot": {}, "stupid": {}, "dumb": {}, "trash": {}, "garbage": {},
		"hate": {}, "useless": {}, "moron": {}, "scam": {},
	}
	return math.Min(1, float64(countWords(words, toxicWords))/float64(max(len(words), 1)))
}

func analyzeCognitiveLoad(text string, words []string) float64 {
	if len(words) == 0 {
		return 0
	}

	technicalMarkers := map[string]struct{}{
		"architecture": {}, "distributed": {}, "database": {}, "concurrency": {},
		"algorithm": {}, "protocol": {}, "compiler": {}, "inference": {},
		"neural": {}, "model": {}, "optimization": {}, "scalability": {},
		"latency": {}, "memory": {}, "postgres": {}, "kubernetes": {},
		"docker": {}, "api": {}, "benchmark": {},
	}

	totalWordLength := 0
	for _, word := range words {
		totalWordLength += len(word)
	}

	wordCount := len(words)
	averageWordLength := float64(totalWordLength) / float64(wordCount)
	sentenceCount := max(countSentences(text), 1)
	averageSentenceLength := float64(wordCount) / float64(sentenceCount)
	technicalMarkerCount := countWords(words, technicalMarkers)

	score := 0.0
	if wordCount > 30 {
		score += 0.25
	}
	if averageWordLength > 6 {
		score += 0.25
	}
	if technicalMarkerCount >= 2 {
		score += 0.25
	}
	if averageSentenceLength > 18 {
		score += 0.25
	}

	return math.Min(1, score)
}

func countWords(words []string, dictionary map[string]struct{}) int {
	count := 0
	for _, word := range words {
		if _, ok := dictionary[word]; ok {
			count++
		}
	}
	return count
}

func countSentences(text string) int {
	count := 0
	for _, char := range text {
		if char == '.' || char == '!' || char == '?' {
			count++
		}
	}
	return count
}
