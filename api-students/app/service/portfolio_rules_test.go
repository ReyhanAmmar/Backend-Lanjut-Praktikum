package service

import (
	"reflect"
	"testing"
)

func TestNormalizeTechs(t *testing.T) {
	input := []string{"  Go ", "go", "PostgreSQL", " POSTGRESQL ", "MongoDB", "", "   "}
	expected := []string{"go", "postgresql", "mongodb"}

	result := NormalizeTechs(input)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("NormalizeTechs() = %v, want %v", result, expected)
	}
}

