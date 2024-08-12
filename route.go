package semanticrouter

import (
	"context"
	"fmt"

	"github.com/conneroisu/go-semantic-router/domain"
	"golang.org/x/sync/errgroup"
	"gonum.org/v1/gonum/mat"
)

// Router represents a semantic router.
//
// Router is a struct that contains a slice of Routes and an Encoder.
//
// Match can be called on a Router to find the best route for a given utterance.
type Router struct {
	Routes             []Route             `json:"routes"  yaml:"routes"  toml:"routes"`                                   // Routes is a slice of Routes.
	Encoder            Encoder             `json:"encoder" yaml:"encoder" toml:"encoder"`                                  // Encoder is an Encoder that encodes utterances into vectors.
	Storage            Store               `json:"storage" yaml:"storage" toml:"storage"`                                  // Storage is a Store that stores the utterances.
	biFuncCoefficients []biFuncCoefficient `json:"biFuncCoefficients" yaml:"biFuncCoefficients" toml:"biFuncCoefficients"` // biFuncCoefficients is a slice of biFuncCoefficients that represent the bi-function coefficients.
}

// Route represents a route in the semantic router.
//
// It is a struct that contains a name and a slice of Utterances.
type Route struct {
	Name       string             `json:"name"       yaml:"name"       toml:"name"`       // Name is the name of the route.
	Utterances []domain.Utterance `json:"utterances" yaml:"utterances" toml:"utterances"` // Utterances is a slice of Utterances.
}

// Encoder represents a encoding driver in the semantic router.
//
// It is an interface that defines a single method, Encode, which takes a string
// and returns a []float64 representing the embedding of the string.
type Encoder interface {
	Encode(ctx context.Context, utterance string) ([]float64, error)
}

// Store is an interface that defines a method, Store, which takes a []float64
// and stores it in a some sort of data store, and a method, Get, which takes a
// string and returns a []float64 from the data store.
type Store interface {
	Store(ctx context.Context, keyValPair domain.Utterance) error
	Get(ctx context.Context, key string) ([]float64, error)
}

// Option is a function that configures a Router.
type Option func(*Router)

// biFuncCoefficient is an struct that represents a function and it's coefficient.
type biFuncCoefficient struct {
	Func        func(queryVec *mat.VecDense, indexVec *mat.VecDense) float64
	Coefficient float64
}

// WithSimilarityDotMatrix sets the similarity function to use with a coefficient.
func WithSimilarityDotMatrix(coefficient float64) Option {
	return func(r *Router) {
		r.biFuncCoefficients = append(r.biFuncCoefficients, biFuncCoefficient{
			Func:        SimilarityDotMatrix,
			Coefficient: coefficient,
		})
	}
}

// WithEuclideanDistance sets the EuclideanDistance function with a coefficient.
func WithEuclideanDistance(coefficient float64) Option {
	return func(r *Router) {
		r.biFuncCoefficients = append(r.biFuncCoefficients, biFuncCoefficient{
			Func:        EuclideanDistance,
			Coefficient: coefficient,
		})
	}
}

// WithManhattanDistance sets the ManhattanDistance function with a coefficient.
func WithManhattanDistance(coefficient float64) Option {
	return func(r *Router) {
		r.biFuncCoefficients = append(r.biFuncCoefficients, biFuncCoefficient{
			Func:        ManhattanDistance,
			Coefficient: coefficient,
		})
	}
}

// WithJaccardSimilarity sets the JaccardSimilarity function with a coefficient.
func WithJaccardSimilarity(coefficient float64) Option {
	return func(r *Router) {
		r.biFuncCoefficients = append(r.biFuncCoefficients, biFuncCoefficient{
			Func:        JaccardSimilarity,
			Coefficient: coefficient,
		})
	}
}

// WithPearsonCorrelation sets the PearsonCorrelation function with a coefficient.
func WithPearsonCorrelation(coefficient float64) Option {
	return func(r *Router) {
		r.biFuncCoefficients = append(r.biFuncCoefficients, biFuncCoefficient{
			Func:        PearsonCorrelation,
			Coefficient: coefficient,
		})
	}
}

// NewRouter creates a new semantic router.
func NewRouter(
	routes []Route,
	encoder Encoder,
	store Store,
	opts ...Option,
) (router *Router, err error) {
	routesLen := len(routes)
	ctx := context.Background()
	for i := 0; i < routesLen; i++ {
		route := routes[i]
		utters := route.Utterances
		for _, utter := range utters {
			en, err := encoder.Encode(ctx, utter.Utterance)
			if err != nil {
				return nil, fmt.Errorf("error encoding utterance: %w", err)
			}
			err = utter.SetEmbedding(en)
			if err != nil {
				return nil, fmt.Errorf("error encoding utterance: %w", err)
			}
			err = store.Store(ctx, utter)
			if err != nil {
				return nil,
					fmt.Errorf(
						"error storing utterance: %s: %w",
						utter.Utterance,
						err,
					)
			}
		}
	}
	return &Router{
		Routes:  routes,
		Encoder: encoder,
		Storage: store,
	}, nil
}

// Match returns the route that matches the given utterance.
//
// The score is the similarity score between the query vector and the index vector.
//
// If the given context is canceled, the context's error is returned if it is non-nil.
func (r *Router) Match(
	ctx context.Context,
	utterance string,
) (bestRouteName string, bestScore float64, err error) {
	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		encoding, err := r.Encoder.Encode(ctx, utterance)
		if err != nil {
			return ErrEncoding{
				Message: fmt.Sprintf(
					"error encoding utterance: %s",
					utterance,
				),
			}
		}
		queryVec := mat.NewVecDense(len(encoding), encoding)
		for _, route := range r.Routes {
			for _, ut := range route.Utterances {
				em, err := r.Storage.Get(ctx, ut.Utterance)
				if err != nil {
					return ErrGetEmbedding{
						Message: fmt.Sprintf(
							"error getting embedding: %s",
							ut.Utterance,
						),
					}
				}
				emLen := len(em)
				if emLen != queryVec.Len() {
					continue
				}
				indexVec := mat.NewVecDense(emLen, em)
				simScore := r.computeScore(queryVec, indexVec)
				if simScore > bestScore {
					bestScore = simScore
					bestRouteName = route.Name
				}
			}
		}
		if bestRouteName == "" {
			return ErrNoRouteFound{
				Message:   "no route found",
				Utterance: utterance,
			}
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		return "", 0.0, fmt.Errorf("no route found: %w", err)
	}
	return bestRouteName, bestScore, nil
}

// computeScore computes the score for a given utterance and route.
//
// It takes a query vector and an index vector as input and returns a score.
//
// Additionally, it leverages the router's biFuncCoefficients to apply different
// weighting factors to functions to get the similarity score.
func (r *Router) computeScore(queryVec *mat.VecDense, indexVec *mat.VecDense) float64 {
	score := 0.0
	for _, fn := range r.biFuncCoefficients {
		score += fn.Coefficient * fn.Func(queryVec, indexVec)
	}
	return score
}
