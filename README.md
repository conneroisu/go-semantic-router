# go-semantic-router

<p align="center">
    <a href="https://pkg.go.dev/github.com/conneroisu/go-semantic-router?tab=doc"><img src="https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white" alt="go.dev"></a>
    <a href="https://github.com/conneroisu/go-semantic-router/actions/workflows/test.yaml"><img src="https://github.com/conneroisu/go-semantic-router/actions/workflows/test.yaml/badge.svg" alt="Build Status"></a>
    <a href="https://codecov.io/gh/conneroisu/go-semantic-router" > <img src="https://codecov.io/gh/conneroisu/go-semantic-router/graph/badge.svg?token=JAGYI2V82D"/> </a>
    <a href="https://goreportcard.com/report/github.com/conneroisu/go-semantic-router"><img src="https://goreportcard.com/badge/github.com/conneroisu/go-semantic-router" alt="Go Report Card"></a>
    <a href="https://www.phorm.ai/query?projectId=fd665f24-5c41-42ed-907b-f322457a562d"><img src="https://img.shields.io/badge/Phorm-Ask_AI-%23F2777A.svg?&logo=data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNSIgaGVpZ2h0PSI0IiBmaWxsPSJub25lIiB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciPgogIDxwYXRoIGQ9Ik00LjQzIDEuODgyYTEuNDQgMS40NCAwIDAgMS0uMDk4LjQyNmMtLjA1LjEyMy0uMTE1LjIzLS4xOTI"
</p>

Go Semantic Router is a superfast decision-making layer for your LLMs and agents written in pure [ Go ](https://go.dev/).

Rather than waiting for slow LLM generations to make tool-use decisions, use the magic of semantic vector space to make those decisions — routing requests using configurable semantic meaning.

A pure-go package for abstractly computing similarity scores between a query vector embedding and a set of vector embeddings.

## Installation

```bash
go get github.com/conneroisu/go-semantic-router
```

## Usage

### OpenAI Encoder

```go
import "github.com/conneroisu/go-semantic-router/encoders/openai"
```

### Google Encoder

```go
import "github.com/conneroisu/go-semantic-router/encoders/google"
```

### Ollama Encoder

```go
import "github.com/conneroisu/go-semantic-router/encoders/ollama"
```


### Conversational Agents Example

```go
// Package main shows how to use the semantic router to find the best route for a given utterance
// in the context of a chat bot or other conversational application.
package main

import (
	"context"
	"fmt"
	"os"

	semantic_router "github.com/conneroisu/go-semantic-router"
	"github.com/conneroisu/go-semantic-router/domain"
	encoders "github.com/conneroisu/go-semantic-router/encoders/openai"
	"github.com/conneroisu/go-semantic-router/stores/memory"
	"github.com/sashabaranov/go-openai"
)

// PoliticsRoutes represents a set of routes that are noteworthy.
var PoliticsRoutes = semantic_router.Route{
	Name: "politics",
	Utterances: []domain.Utterance{
		{Utterance: "isn't politics the best thing ever"},
		{Utterance: "why don't you tell me about your political opinions"},
		{Utterance: "don't you just love the president"},
		{Utterance: "they're going to destroy this country!"},
		{Utterance: "they will save the country!"},
	},
}

// ChitchatRoutes represents a set of routes that are noteworthy.
var ChitchatRoutes = semantic_router.Route{
	Name: "chitchat",
	Utterances: []domain.Utterance{
		{Utterance: "how's the weather today?"},
		{Utterance: "how are things going?"},
		{Utterance: "lovely weather today"},
		{Utterance: "the weather is horrendous"},
		{Utterance: "let's go to the chippy"},
	},
}

// main runs the example.
func main() {
	if err := run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// run runs the example.
func run() error {
	ctx := context.Background()
	store := memory.NewStore()
	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))
	router, err := semantic_router.NewRouter(
		[]semantic_router.Route{
			PoliticsRoutes,
			ChitchatRoutes,
		},
		encoders.Encoder{
			Client: client,
		},
		store,
	)
	if err != nil {
		return fmt.Errorf("error creating router: %w", err)
	}

	finding, p, err := router.Match(ctx, "how's the weather today?")
	if err != nil {
		fmt.Println("Error:", err)
	}

	fmt.Println("p:", p)
	fmt.Println("Found:", finding)
	return nil
}
```

Output:

```
Found: chitchat
```

### Veterinarian Example

The following example shows how to use the semantic router to find the best route for a given utterance in the context of a veterinarian appointment.

The goal of the example is to decide whether spoken utterances are relevant to a noteworthy conversation or a chitchat conversation.


#### The Code Example:
```go
// Package main shows how to use the semantic router to find the best route for a given utterance
// in the context of a veterinarian appointment.
package main

import (
	"context"
	"fmt"
	"os"

	semantic_router "github.com/conneroisu/go-semantic-router"
	"github.com/conneroisu/go-semantic-router/domain"
	"github.com/conneroisu/go-semantic-router/encoders/ollama"
	"github.com/conneroisu/go-semantic-router/stores/memory"
	"github.com/ollama/ollama/api"
)

// NoteworthyRoutes represents a set of routes that are noteworthy.
// noteworthy here means that the routes are likely to be relevant to a noteworthy conversation in a veterinarian appointment.
var NoteworthyRoutes = semantic_router.Route{
	Name: "noteworthy",
	Utterances: []domain.Utterance{
		{Utterance: "what is the best way to treat a dog with a cold?"},
		{Utterance: "my cat has been limping, what should I do?"},
	},
}

// ChitchatRoutes represents a set of routes that are chitchat.
// chitchat here means that the routes are likely to be relevant to a chitchat conversation in a veterinarian appointment.
var ChitchatRoutes = semantic_router.Route{
	Name: "chitchat",
	Utterances: []domain.Utterance{
		{Utterance: "what is your favorite color?"},
		{Utterance: "what is your favorite animal?"},
	},
}

// main runs the example.
func main() {
	if err := run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// run runs the example.
func run() error {
	ctx := context.Background()
	cli, err := api.ClientFromEnvironment()
	if err != nil {
		return fmt.Errorf("error creating client: %w", err)
	}
	router, err := semantic_router.NewRouter(
		[]semantic_router.Route{NoteworthyRoutes, ChitchatRoutes},
		&ollama.Encoder{
			Client: cli,
			Model:  "mxbai-embed-large",
		},
		memory.NewStore(),
	)
	if err != nil {
		return fmt.Errorf("error creating router: %w", err)
	}
	finding, p, err := router.Match(ctx, "how's the weather today?")
	if err != nil {
		fmt.Println("Error:", err)
	}
	fmt.Println("Found:", finding)
	fmt.Println("p:", p)
	return nil
}
```

#### The Output

The output of the veterinarian example is:
```bash
Found: chitchat
p: 0.4656368810166642
```
