package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Protocol struct {
	Criteria          []string `json:"criteria" bson:"criteria"`
	InclusionCriteria string   `json:"inclusion_criteria" bson:"inclusion_criteria"`
	ExclusionCriteria string   `json:"exclusion_criteria" bson:"exclusion_criteria"`
}

type Project struct {
	ID           bson.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	ProjectId    string        `json:"projectId" bson:"projectId"`
	Name         string        `json:"name" bson:"name"`
	Description  string        `json:"description" bson:"description"`
	LLMProvider  string        `json:"llmProvider" bson:"llmProvider"` // "openai" | "groq" | "gemini" | "claude"
	LLMModel     string        `json:"llmModel,omitempty" bson:"llmModel,omitempty"`
	LLMApiKey    string        `json:"llmApiKey,omitempty" bson:"llmApiKey,omitempty"` // AES-256 Encrypted
	TotalStudies int           `json:"totalStudies" bson:"totalStudies"`
	Protocol     Protocol      `json:"protocol" bson:"protocol"`
	CreatedAt    time.Time     `json:"createdAt" bson:"createdAt"`
	UpdatedAt    time.Time     `json:"updatedAt" bson:"updatedAt"`

	// Real-time computed or cached stats for dashboard
	IncludedCount  int `json:"includedCount,omitempty" bson:"includedCount,omitempty"`
	ExcludedCount  int `json:"excludedCount,omitempty" bson:"excludedCount,omitempty"`
	UndecidedCount int `json:"undecidedCount,omitempty" bson:"undecidedCount,omitempty"`
	AIScreenedCount int `json:"aiScreenedCount,omitempty" bson:"aiScreenedCount,omitempty"`
}
