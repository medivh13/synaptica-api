package service

import (
	"context"
	"fmt"
	"strings"

	"synaptica-api/internal/analyzer"
	"synaptica-api/internal/model"
	"synaptica-api/internal/repository"
)

type CommentAnalysisService interface {
	RunCommentAnalysis(ctx context.Context, input RunCommentAnalysisInput) (*RunCommentAnalysisResult, error)
	FindAnalyzedComments(ctx context.Context, input FindAnalyzedCommentsInput) ([]model.AnalyzedComment, error)
}

type RunCommentAnalysisInput struct {
	PostID    string
	Source    string
	KeywordID string
	Limit     int
}

type RunCommentAnalysisResult struct {
	EntityType        string `json:"entity_type"`
	Analyzed          int    `json:"analyzed"`
	InsertedOrUpdated int    `json:"inserted_or_updated"`
}

type FindAnalyzedCommentsInput struct {
	PostID    string
	Source    string
	KeywordID string
	Limit     int
	Offset    int
}

type commentAnalysisService struct {
	analysisRepo repository.AnalysisRepository
	textAnalyzer analyzer.TextAnalyzer
}

func NewCommentAnalysisService(analysisRepo repository.AnalysisRepository, textAnalyzer analyzer.TextAnalyzer) CommentAnalysisService {
	return &commentAnalysisService{
		analysisRepo: analysisRepo,
		textAnalyzer: textAnalyzer,
	}
}

func (s *commentAnalysisService) RunCommentAnalysis(ctx context.Context, input RunCommentAnalysisInput) (*RunCommentAnalysisResult, error) {
	targets, err := s.analysisRepo.FindUnanalyzedComments(ctx, repository.AnalysisCommentFilter{
		PostID:    strings.TrimSpace(input.PostID),
		Source:    strings.TrimSpace(strings.ToLower(input.Source)),
		KeywordID: strings.TrimSpace(input.KeywordID),
		Limit:     input.Limit,
	})
	if err != nil {
		return nil, fmt.Errorf("find unanalyzed comments: %w", err)
	}

	results := make([]model.TextAnalysisResult, 0, len(targets))
	for _, target := range targets {
		output := s.textAnalyzer.Analyze(target.Body)
		results = append(results, model.TextAnalysisResult{
			EntityType:         "comment",
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

	upserted, err := s.analysisRepo.UpsertAnalysisResults(ctx, results)
	if err != nil {
		return nil, fmt.Errorf("upsert comment analysis results: %w", err)
	}

	return &RunCommentAnalysisResult{
		EntityType:        "comment",
		Analyzed:          len(results),
		InsertedOrUpdated: upserted,
	}, nil
}

func (s *commentAnalysisService) FindAnalyzedComments(ctx context.Context, input FindAnalyzedCommentsInput) ([]model.AnalyzedComment, error) {
	comments, err := s.analysisRepo.FindAnalyzedComments(ctx, repository.AnalysisCommentResultFilter{
		PostID:    strings.TrimSpace(input.PostID),
		Source:    strings.TrimSpace(strings.ToLower(input.Source)),
		KeywordID: strings.TrimSpace(input.KeywordID),
		Limit:     input.Limit,
		Offset:    input.Offset,
	})
	if err != nil {
		return nil, fmt.Errorf("find analyzed comments: %w", err)
	}

	return comments, nil
}
