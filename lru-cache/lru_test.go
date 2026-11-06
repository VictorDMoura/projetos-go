package main

import (
	"reflect"
	"testing"
)

func TestLRUCache_KeyNotFound(t *testing.T) {
	cache := NewLRUCache(2)
	val := cache.Get("inexistente")
	if val != -1 {
		t.Fatalf("Esperava -1 para chave inexistente, obteve %v", val)
	}
}

func TestLRUCache_BasicGetSet(t *testing.T) {
	cache := NewLRUCache(2)

	cache.Set("chave1", "valor1")
	val := cache.Get("chave1")
	if val != "valor1" {
		t.Fatalf("Esperava 'valor1', obteve %v", val)
	}
}

func TestLRUCache_Eviction(t *testing.T) {
	cache := NewLRUCache(2)

	cache.Set("A", 1)
	cache.Set("B", 2)
	// Ordem esperada: B (recente), A (antigo)

	// Inserir C deve remover o mais antigo (A)
	cache.Set("C", 3)

	if val := cache.Get("A"); val != -1 {
		t.Fatalf("Esperava -1 para chave 'A' descartada, mas obteve %v", val)
	}

	if val := cache.Get("B"); val != 2 {
		t.Fatalf("Esperava 2 para chave 'B', mas obteve %v", val)
	}

	if val := cache.Get("C"); val != 3 {
		t.Fatalf("Esperava 3 para chave 'C', mas obteve %v", val)
	}
}

func TestLRUCache_GetUpdatesOrder(t *testing.T) {
	cache := NewLRUCache(2)

	cache.Set("A", 1)
	cache.Set("B", 2)
	// Acesso a 'A': agora 'A' é o mais recente e 'B' é o mais antigo
	valA := cache.Get("A")
	if valA != 1 {
		t.Fatalf("Esperava 1 para 'A', obteve %v", valA)
	}

	// Inserir 'C': 'B' deve ser descartado, pois 'A' foi acessado recentemente
	cache.Set("C", 3)

	if val := cache.Get("B"); val != -1 {
		t.Fatalf("Esperava que 'B' tivesse sido descartado (-1), mas obteve %v", val)
	}

	if val := cache.Get("A"); val != 1 {
		t.Fatalf("Esperava que 'A' permanecesse no cache (1), mas obteve %v", val)
	}

	if val := cache.Get("C"); val != 3 {
		t.Fatalf("Esperava que 'C' existisse no cache (3), mas obteve %v", val)
	}

	// Ordem esperada agora: C (mais recente), A (antigo)
	expectedKeys := []string{"C", "A"}
	if !reflect.DeepEqual(cache.Keys(), expectedKeys) {
		t.Fatalf("Esperava chaves %v, obteve %v", expectedKeys, cache.Keys())
	}
}

func TestLRUCache_UpdateExistingKey(t *testing.T) {
	cache := NewLRUCache(2)

	cache.Set("A", 1)
	cache.Set("B", 2)

	// Atualizar 'A' não deve aumentar a quantidade de itens
	cache.Set("A", 100)

	if val := cache.Get("A"); val != 100 {
		t.Fatalf("Esperava valor atualizado 100 para 'A', obteve %v", val)
	}

	// Agora a ordem deve ser A (mais recente), B (antigo)
	// Inserir 'C' deve descartar 'B', não 'A'
	cache.Set("C", 300)

	if val := cache.Get("B"); val != -1 {
		t.Fatalf("Esperava -1 para 'B', mas obteve %v", val)
	}

	if val := cache.Get("A"); val != 100 {
		t.Fatalf("Esperava 100 para 'A', mas obteve %v", val)
	}
}

func TestLRUCache_DifferentTypes(t *testing.T) {
	cache := NewLRUCache(3)

	type CustomData struct {
		ID  int
		Tag string
	}

	cache.Set("int", 42)
	cache.Set("slice", []int{1, 2, 3})
	cache.Set("struct", CustomData{ID: 7, Tag: "teste"})

	if val := cache.Get("int"); val != 42 {
		t.Fatalf("Esperava 42, obteve %v", val)
	}

	sliceVal, ok := cache.Get("slice").([]int)
	if !ok || len(sliceVal) != 3 || sliceVal[0] != 1 {
		t.Fatalf("Falha ao recuperar slice: %v", cache.Get("slice"))
	}

	structVal, ok := cache.Get("struct").(CustomData)
	if !ok || structVal.ID != 7 || structVal.Tag != "teste" {
		t.Fatalf("Falha ao recuperar struct: %v", cache.Get("struct"))
	}
}

func TestLRUCache_CapacityEdgeCases(t *testing.T) {
	// Capacidade 1
	cache1 := NewLRUCache(1)
	cache1.Set("X", 10)
	cache1.Set("Y", 20)

	if val := cache1.Get("X"); val != -1 {
		t.Fatalf("Esperava que 'X' fosse descartado com capacidade 1, obteve %v", val)
	}
	if val := cache1.Get("Y"); val != 20 {
		t.Fatalf("Esperava 20 para 'Y', obteve %v", val)
	}

	// Capacidade 0 ou negativa não deve causar pânico
	cache0 := NewLRUCache(0)
	cache0.Set("A", 1)
	if val := cache0.Get("A"); val != -1 {
		t.Fatalf("Esperava -1 para cache com capacidade 0, obteve %v", val)
	}
}
