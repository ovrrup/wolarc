package views

import (
	"encoding/json"
	"testing"
)

func TestVectorSliceUnmarshal(t *testing.T) {
	// 1D vector JSON
	raw1D := []byte(`{"chunk_id": 1, "text": "hello", "hash": "abc", "vector": [0.1, 0.2, 0.3]}`)
	var c1 ChunkRecord
	if err := json.Unmarshal(raw1D, &c1); err != nil {
		t.Fatalf("Failed to unmarshal 1D vector: %v", err)
	}
	if len(c1.Vector) != 3 {
		t.Fatalf("Expected 3 items in vector, got %d", len(c1.Vector))
	}

	// 2D vector JSON
	raw2D := []byte(`{"chunk_id": 2, "text": "world", "hash": "def", "vector": [[0.1, 0.2, 0.3]]}`)
	var c2 ChunkRecord
	if err := json.Unmarshal(raw2D, &c2); err != nil {
		t.Fatalf("Failed to unmarshal 2D vector: %v", err)
	}
	if len(c2.Vector) != 3 {
		t.Fatalf("Expected 3 items in unrolled 2D vector, got %d", len(c2.Vector))
	}
}

func TestExtractVectorNested(t *testing.T) {
	var input interface{}
	raw := []byte(`[{"index": 0, "embedding": [[0.5, 0.6, 0.7]]}]`)
	if err := json.Unmarshal(raw, &input); err != nil {
		t.Fatalf("Failed to parse json: %v", err)
	}
	vec := extractVector(input)
	if len(vec) != 3 || vec[0] != 0.5 {
		t.Fatalf("Unexpected extracted vector: %v", vec)
	}
}
