package proxy

// openAIModelIDs is what GET /v1/models advertises when
// config.openai_compatible_models is enabled. These ids are cosmetic: the
// proxy never routes to OpenAI, every one of them resolves to
// config.default_model (see Server.mapModel). They exist so OpenAI-compatible
// clients that pick a model from the list keep working.
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

var openAIModelSet = func() map[string]struct{} {
	m := make(map[string]struct{}, len(openAIModelIDs))
	for _, id := range openAIModelIDs {
		m[id] = struct{}{}
	}
	return m
}()

// isOpenAIModel reports whether id is one of the advertised OpenAI ids.
func isOpenAIModel(id string) bool {
	_, ok := openAIModelSet[id]
	return ok
}

// advertisedModels returns the ids and owner reported by GET /v1/models.
func (s *Server) advertisedModels() ([]string, string) {
	if s.cfg.OpenAICompatibleModels {
		return openAIModelIDs, "openai"
	}
	return s.cfg.Models, "quickwork"
}
