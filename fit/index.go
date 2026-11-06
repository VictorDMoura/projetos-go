package main

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
)

// IndexEntry representa uma entrada na área de preparação (staging area).
// Contém o caminho relativo do arquivo, o hash SHA-1 do conteúdo e o tamanho em bytes.
type IndexEntry struct {
	Path string `json:"path"`
	Hash string `json:"hash"`
	Size int64  `json:"size"`
}

// Index representa o índice da staging area do repositório (.fit/index.json).
type Index struct {
	Entries map[string]IndexEntry `json:"entries"`
}

// NewIndex cria uma nova instância vazia do índice.
func NewIndex() *Index {
	return &Index{
		Entries: make(map[string]IndexEntry),
	}
}

// LoadIndex lê e desserializa o arquivo de índice (.fit/index.json).
func LoadIndex(indexPath string) (*Index, error) {
	data, err := os.ReadFile(indexPath)
	if err != nil {
		if os.IsNotExist(err) {
			return NewIndex(), nil
		}
		return nil, err
	}

	idx := NewIndex()
	if len(data) == 0 {
		return idx, nil
	}

	if err := json.Unmarshal(data, idx); err != nil {
		return nil, err
	}
	if idx.Entries == nil {
		idx.Entries = make(map[string]IndexEntry)
	}
	return idx, nil
}

// SaveIndex grava o índice atualizado no disco em .fit/index.json.
func (idx *Index) Save(indexPath string) error {
	data, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(indexPath, data, 0644)
}

// HashFile calcula o hash SHA-1 do conteúdo de um arquivo no disco.
func HashFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	h := sha1.New()
	if _, err := io.Copy(h, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// HashBytes calcula o hash SHA-1 de uma fatia de bytes em memória.
func HashBytes(content []byte) string {
	h := sha1.New()
	h.Write(content)
	return hex.EncodeToString(h.Sum(nil))
}
