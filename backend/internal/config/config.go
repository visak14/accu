package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port           string
	MongoURI       string
	DBName         string
	SecretKey      string // Used for AES-256 encryption of API keys
	DefaultLLMProv string
	OpenAIApiKey   string
	GroqApiKey     string
	GeminiApiKey   string
	ClaudeApiKey   string
}

func LoadConfig() *Config {
	// Try loading .env from current directory or parent directory
	_ = godotenv.Load(".env")
	_ = godotenv.Load("../.env")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "sympro_slr"
	}

	secretKey := os.Getenv("SECRET_KEY")
	if secretKey == "" {
		secretKey = "sympro-slr-secret-encryption-key-32b" // 32 bytes fallback
	}

	defaultLLMProv := os.Getenv("DEFAULT_LLM_PROVIDER")
	if defaultLLMProv == "" {
		defaultLLMProv = "groq"
	}

	return &Config{
		Port:           port,
		MongoURI:       mongoURI,
		DBName:         dbName,
		SecretKey:      secretKey,
		DefaultLLMProv: defaultLLMProv,
		OpenAIApiKey:   os.Getenv("OPENAI_API_KEY"),
		GroqApiKey:     os.Getenv("GROQ_API_KEY"),
		GeminiApiKey:   os.Getenv("GEMINI_API_KEY"),
		ClaudeApiKey:   os.Getenv("CLAUDE_API_KEY"),
	}
}
