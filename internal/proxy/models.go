package proxy

import "strings"

// openAIModelIDs is the built-in list advertised by GET /v1/models when
// config.openai_compatible_models is enabled and config.openai_models is
// empty. These ids are cosmetic: the proxy never routes to OpenAI, every one
// of them resolves to config.default_model (see Server.mapModel). They exist
// so OpenAI-compatible clients that pick a model from the list keep working.
var openAIModelIDs = []string{
	// GPT-6.2 / 6.1 / 6 family (current lineup at the time of writing).
	"gpt-6.2-astra",
	"gpt-6.2-sol",
	"gpt-6.1-terra",
	"gpt-6.1-sol",
	"gpt-6-astra",
	"gpt-6-sol",
	"gpt-6-luna",
	// Previous generation, kept so older clients keep working.
	"gpt-5.5",
	"gpt-5.4",
	"gpt-5.4-mini",
	"o3",
	"o4-mini",
}

// normalizedModelIDs trims entries, drops empty ones and removes
// case-insensitive duplicates while keeping the first spelling.
func normalizedModelIDs(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	out := make([]string, 0, len(in))
	seen := make(map[string]struct{}, len(in))
	for _, raw := range in {
		id := strings.TrimSpace(raw)
		if id == "" {
			continue
		}
		key := strings.ToLower(id)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, id)
	}
	return out
}

// effectiveOpenAIModels returns the ids advertised in compatible mode.
// config.openai_models overrides the built-in list when it has entries.
func (s *Server) effectiveOpenAIModels() []string {
	if ids := normalizedModelIDs(s.cfg.OpenAIModels); len(ids) > 0 {
		return ids
	}
	return openAIModelIDs
}

// advertisedModels returns the ids and owner reported by GET /v1/models.
func (s *Server) advertisedModels() ([]string, string) {
	if s.cfg.OpenAICompatibleModels {
		return s.effectiveOpenAIModels(), "openai"
	}
	return s.cfg.Models, "quickwork"
}

// isAdvertisedOpenAIModel reports whether id appears in the effective list.
func (s *Server) isAdvertisedOpenAIModel(id string) bool {
	for _, m := range s.effectiveOpenAIModels() {
		if strings.EqualFold(m, id) {
			return true
		}
	}
	return false
}
