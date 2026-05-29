package service

import (
	"context"
	"fmt"
	"strings"

	"synaptica-api/internal/analyzer"
	"synaptica-api/internal/model"
	"synaptica-api/internal/repository"
)

type PostAnalysisService interface {
	RunPostAnalysis(ctx context.Context, input RunPostAnalysisInput) (*RunPostAnalysisResult, error)
	FindAnalyzedPosts(ctx context.Context, input FindAnalyzedPostsInput) ([]model.AnalyzedPost, error)
}

type RunPostAnalysisInput struct {
	Source    string
	KeywordID string
	Limit     int
}

type RunPostAnalysisResult struct {
	EntityType        string `json:"entity_type"`
	Analyzed          int    `json:"analyzed"`
	InsertedOrUpdated int    `json:"inserted_or_updated"`
}

type FindAnalyzedPostsInput struct {
	Source    string
	KeywordID string
	Limit     int
	Offset    int
}

type postAnalysisService struct {
	analysisRepo repository.AnalysisRepository
	textAnalyzer analyzer.TextAnalyzer
}

func NewPostAnalysisService(analysisRepo repository.AnalysisRepository, textAnalyzer analyzer.TextAnalyzer) PostAnalysisService {
	return &postAnalysisService{
		analysisRepo: analysisRepo,
		textAnalyzer: textAnalyzer,
	}
}

func (s *postAnalysisService) RunPostAnalysis(ctx context.Context, input RunPostAnalysisInput) (*RunPostAnalysisResult, error) {
	source := strings.TrimSpace(strings.ToLower(input.Source))
	keywordID := strings.TrimSpace(input.KeywordID)

	targets, err := s.analysisRepo.FindUnanalyzedPosts(ctx, repository.AnalysisPostFilter{
		Source:    source,
		KeywordID: keywordID,
		Limit:     input.Limit,
	})
	if err != nil {
		return nil, fmt.Errorf("find unanalyzed posts: %w", err)
	}

	results := make([]model.TextAnalysisResult, 0, len(targets))
	for _, target := range targets {
		output := s.textAnalyzer.Analyze(combinePostText(target))
		results = append(results, model.TextAnalysisResult{
			EntityType:         "post",
			EntityID:           target.ID,
			SentimentLabel:     output.SentimentLabel,
			SentimentScore:     output.SentimentScore,
			EmotionLabel:       output.EmotionLabel,
			EmotionScore:       output.EmotionScore,
			ToxicityScore:      output.ToxicityScore,
			CognitiveLoadScore: output.CognitiveLoadScore,
			Language:           output.Language,
		})
	}

	upserted, err := s.analysisRepo.UpsertPostAnalysisResults(ctx, results)
	if err != nil {
		return nil, fmt.Errorf("upsert post analysis results: %w", err)
	}

	return &RunPostAnalysisResult{
		EntityType:        "post",
		Analyzed:          len(results),
		InsertedOrUpdated: upserted,
	}, nil
}

func (s *postAnalysisService) FindAnalyzedPosts(ctx context.Context, input FindAnalyzedPostsInput) ([]model.AnalyzedPost, error) {
	posts, err := s.analysisRepo.FindAnalyzedPosts(ctx, repository.AnalysisResultFilter{
		Source:    strings.TrimSpace(strings.ToLower(input.Source)),
		KeywordID: strings.TrimSpace(input.KeywordID),
		Limit:     input.Limit,
		Offset:    input.Offset,
	})
	if err != nil {
		return nil, fmt.Errorf("find analyzed posts: %w", err)
	}

	return posts, nil
}

func combinePostText(post model.PostAnalysisTarget) string {
	if post.Body == nil {
		return post.Title
	}
	return post.Title + " " + *post.Body
}
