package proxy

import (
	"testing"

	"quickwork-openai/internal/config"
)

func TestAdvertisedModelsHonoursFlag(t *testing.T) {
	off := &Server{cfg: &config.Config{
		Models:       []string{"smart", "fast"},
		DefaultModel: "smart",
	}}
	ids, owner := off.advertisedModels()
	if owner != "quickwork" {
		t.Fatalf("owner = %q, want quickwork", owner)
	}
	if len(ids) != 2 || ids[0] != "smart" {
		t.Fatalf("ids = %v, want the configured aliases", ids)
	}

	on := &Server{cfg: &config.Config{
		Models:                 []string{"smart"},
		DefaultModel:           "smart",
		OpenAICompatibleModels: true,
	}}
	ids, owner = on.advertisedModels()
	if owner != "openai" {
		t.Fatalf("owner = %q, want openai", owner)
	}
	if len(ids) != len(openAIModelIDs) {
		t.Fatalf("got %d ids, want %d", len(ids), len(openAIModelIDs))
	}
	if !on.isAdvertisedOpenAIModel("gpt-6.2-astra") || !on.isAdvertisedOpenAIModel("o3") {
		t.Fatalf("expected common OpenAI ids in the advertised list")
	}
}

func TestOpenAIModelsOverrideBuiltinList(t *testing.T) {
	s := &Server{cfg: &config.Config{
		Models:                 []string{"smart"},
		DefaultModel:           "smart",
		OpenAICompatibleModels: true,
		OpenAIModels: []string{
			" gpt-6.2-astra ",
			"",
			"GPT-6.2-ASTRA", // duplicate, different case
			"gpt-6.2-sol",
		},
	}}

	ids, owner := s.advertisedModels()
	if owner != "openai" {
		t.Fatalf("owner = %q, want openai", owner)
	}
	if len(ids) != 2 || ids[0] != "gpt-6.2-astra" || ids[1] != "gpt-6.2-sol" {
		t.Fatalf("ids = %v, want the trimmed/deduped override", ids)
	}
	if !s.isAdvertisedOpenAIModel("GPT-6.2-SOL") {
		t.Fatal("override ids must match case-insensitively")
	}
	if s.isAdvertisedOpenAIModel("gpt-6-luna") {
		t.Fatal("built-in ids must not leak into an explicit override")
	}
	if got := s.mapModel("gpt-6.2-sol"); got != "smart" {
		t.Fatalf("mapModel(gpt-6.2-sol) = %q, want smart", got)
	}
}

func TestOpenAIModelsIgnoredWhenFlagOff(t *testing.T) {
	s := &Server{cfg: &config.Config{
		Models:       []string{"smart"},
		DefaultModel: "smart",
		OpenAIModels: []string{"my-model"},
	}}

	ids, owner := s.advertisedModels()
	if owner != "quickwork" || len(ids) != 1 || ids[0] != "smart" {
		t.Fatalf("flag off must keep local aliases, got %v (%s)", ids, owner)
	}
}

func TestMapModelRedirectsAdvertisedIds(t *testing.T) {
	s := &Server{cfg: &config.Config{
		Models:                 []string{"smart"},
		DefaultModel:           "smart",
		OpenAICompatibleModels: true,
	}}

	for _, model := range []string{"gpt-6.2-astra", "GPT-6.1-TERRA", "openai/gpt-6.2-sol", "o3"} {
		if got := s.mapModel(model); got != "smart" {
			t.Fatalf("mapModel(%q) = %q, want smart", model, got)
		}
	}

	// Local aliases must keep working while compatibility mode is on.
	if got := s.mapModel("fast"); got != "fast" {
		t.Fatalf("mapModel(fast) = %q, want fast", got)
	}
}

func TestMapModelKeepsRedirectWhenFlagOff(t *testing.T) {
	s := &Server{cfg: &config.Config{
		Models:       []string{"smart"},
		DefaultModel: "smart",
	}}

	// Unknown models already fall back to the default model; the flag only
	// changes what /v1/models advertises.
	if got := s.mapModel("gpt-4o"); got != "smart" {
		t.Fatalf("mapModel(gpt-4o) = %q, want smart", got)
	}
	if got := s.mapModel("openai/gpt-4o"); got != "smart" {
		t.Fatalf("mapModel(openai/gpt-4o) = %q, want smart", got)
	}
}
