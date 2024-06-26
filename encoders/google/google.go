package encoders

import (
	"context"

	"github.com/google/generative-ai-go/genai"
)

// GoogleEncoder encodes a query string into a Google search URL.
type GoogleEncoder struct {
	Ctx context.Context

	client genai.Client
	name   string
}

// NewGoogleEncoder creates a new GoogleEncoder.
func NewGoogleEncoder(
	ctx context.Context,
	client genai.Client,
) *GoogleEncoder {
	return &GoogleEncoder{client: client}
}

// Encode encodes a query string into a Google search URL.
func (e *GoogleEncoder) Encode(query string) ([]float64, error) {
	model := e.client.EmbeddingModel(e.name)
	embedding, err := model.EmbedContent(e.Ctx)
	if err != nil {
		return nil, err
	}
	// type float32
	a := embedding.Embedding.Values
	// convert to []float64
	b := make([]float64, len(a))
	for i, v := range a {
		b[i] = float64(v)
	}
	return b, nil
}
