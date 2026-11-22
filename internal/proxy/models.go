package proxy

// openAIModelIDs is what GET /v1/models advertises when
// config.openai_compatiable_models is enabled. These ids are cosmetic: the
// proxy never routes to OpenAI, every one of them resolves to
// config.default_model (see Server.mapModel). They exist so OpenAI-compatible
// clients that pick a model from the list keep working.
var openAIModelIDs = []string{
	"gpt-5.5",
	"gpt-5.4",
	"gpt-5.4-mini",
	"gpt-5.4-nano",
	"gpt-5.2",
	"gpt-5.1",
	"gpt-5",
	"gpt-5-mini",
	"gpt-5-nano",
	"o3",
	"o3-mini",
	"o4-mini",
	"gpt-4.1",
	"gpt-4.1-mini",
	"gpt-4o",
	"gpt-4o-mini",
	"gpt-3.5-turbo",
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
