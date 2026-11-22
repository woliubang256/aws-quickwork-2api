package config

import (
	"encoding/json"
	"strings"
	"testing"
)

// The JSON key is part of the public interface; keep it spelled correctly.
func TestOpenAICompatibleModelsJSONKey(t *testing.T) {
	var c Config
	if err := json.Unmarshal([]byte(`{"openai_compatible_models":true}`), &c); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !c.OpenAICompatibleModels {
		t.Fatal("openai_compatible_models must enable the flag")
	}

	body, err := json.Marshal(Config{OpenAICompatibleModels: true})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(body), `"openai_compatible_models":true`) {
		t.Fatalf("serialized config must use openai_compatible_models:\n%s", body)
	}
}

func TestOpenAIModelsJSONKey(t *testing.T) {
	var c Config
	if err := json.Unmarshal([]byte(`{"openai_models":["gpt-6.2-astra","gpt-6.2-sol"]}`), &c); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(c.OpenAIModels) != 2 || c.OpenAIModels[0] != "gpt-6.2-astra" {
		t.Fatalf("openai_models = %v", c.OpenAIModels)
	}
}
