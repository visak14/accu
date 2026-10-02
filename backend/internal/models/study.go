package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type AISuggestionResult struct {
	Suggestion string   `json:"suggestion"` // "include" | "exclude"
	Confidence float64  `json:"confidence"` // 0.0 - 1.0
	Reasoning  string   `json:"reasoning"`
	Matches    []string `json:"matches"`
}

type Study struct {
	ID             bson.ObjectID       `json:"_id,omitempty" bson:"_id,omitempty"`
	StudyId        string              `json:"studyId" bson:"studyId"`
	ProjectId      string              `json:"projectId" bson:"projectId"`
	RawID          string              `json:"ID" bson:"ID"`
	Title          string              `json:"title" bson:"title"`
	Year           int                 `json:"year" bson:"year"`
	Author         string              `json:"author" bson:"author"`
	Abstract       string              `json:"abstract" bson:"abstract"`
	ArticleType    string              `json:"article_type" bson:"article_type"`
	Decision       string              `json:"decision" bson:"decision"` // "undecided" | "included" | "excluded"
	AISuggestion   *string             `json:"aiSuggestion" bson:"aiSuggestion"` // null | "include" | "exclude"
	AIConfidence   *float64            `json:"aiConfidence" bson:"aiConfidence"`
	AIReason       *string             `json:"aiReason" bson:"aiReason"`
	AIMatches      []string            `json:"aiMatches,omitempty" bson:"aiMatches,omitempty"`
	AIJsonResponse *AISuggestionResult `json:"aiJsonResponse" bson:"aiJsonResponse"`
	DecidedAt      *time.Time          `json:"decidedAt" bson:"decidedAt"`
	AIScreenedAt   *time.Time          `json:"aiScreenedAt,omitempty" bson:"aiScreenedAt,omitempty"`
}

type DecisionUpdateRequest struct {
	Decision string `json:"decision"` // "included" | "excluded" | "undecided"
}

type AISuggestRequest struct {
	Provider string `json:"provider,omitempty"`
	Model    string `json:"model,omitempty"`
	ApiKey   string `json:"apiKey,omitempty"`
}
