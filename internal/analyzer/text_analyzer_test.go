package analyzer

import "testing"

func TestRuleBasedTextAnalyzerSentimentAndEmotion(t *testing.T) {
	analyzer := NewRuleBasedTextAnalyzer()

	output := analyzer.Analyze("Amazing breakthrough and useful API. I love this promising model.")

	if output.SentimentLabel != "positive" {
		t.Fatalf("expected positive sentiment, got %q", output.SentimentLabel)
	}
	if output.EmotionLabel != "joy" {
		t.Fatalf("expected joy emotion, got %q", output.EmotionLabel)
	}
	if output.Language != "en" {
		t.Fatalf("expected language en, got %q", output.Language)
	}
}

func TestRuleBasedTextAnalyzerToxicity(t *testing.T) {
	analyzer := NewRuleBasedTextAnalyzer()

	output := analyzer.Analyze("This is stupid garbage and a scam")

	if output.SentimentLabel != "negative" {
		t.Fatalf("expected negative sentiment, got %q", output.SentimentLabel)
	}
	if output.ToxicityScore <= 0 {
		t.Fatalf("expected toxicity score above zero, got %f", output.ToxicityScore)
	}
}
