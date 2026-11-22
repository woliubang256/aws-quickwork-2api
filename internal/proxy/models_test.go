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
	if !isOpenAIModel("gpt-6.2-astra") || !isOpenAIModel("o3") {
		t.Fatalf("expected common OpenAI ids in the advertised list")
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
