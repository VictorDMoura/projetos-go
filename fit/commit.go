package main

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"time"
)

// Commit representa um objeto de commit no Fit.
//
// Requisitos atendidos:
// - Data de criação (Timestamp)
// - Nome do autor (AuthorName)
// - Email do autor (AuthorEmail)
// - Mensagem explicativa (Message)
// - Código hash único calculado via SHA-1 (Hash)
// - ParentHash: aponta para o commit anterior (permite rastrear o histórico para o log)
// - Files: mapa com snapshot dos arquivos no momento do commit (caminho -> hash do conteúdo)
type Commit struct {
	Hash        string            `json:"hash"`
	ParentHash  string            `json:"parent_hash,omitempty"`
	AuthorName  string            `json:"author_name"`
	AuthorEmail string            `json:"author_email"`
	Timestamp   time.Time         `json:"timestamp"`
	Message     string            `json:"message"`
	Files       map[string]string `json:"files"`
}

// CalculateHash gera o hash SHA-1 determinístico do commit com base em seus atributos.
// Garante que cada commit tenha uma identidade criptográfica única.
func (c *Commit) CalculateHash() string {
	h := sha1.New()

	// Ordenamos as chaves dos arquivos para garantir hash determinístico
	sortedKeys := make([]string, 0, len(c.Files))
	for k := range c.Files {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)

	var filesBuilder strings.Builder
	for _, k := range sortedKeys {
		filesBuilder.WriteString(fmt.Sprintf("%s:%s;", k, c.Files[k]))
	}

	payload := fmt.Sprintf(
		"parent:%s\nauthor:%s <%s>\ntimestamp:%s\nmessage:%s\nfiles:%s",
		c.ParentHash,
		c.AuthorName,
		c.AuthorEmail,
		c.Timestamp.UTC().Format(time.RFC3339Nano),
		c.Message,
		filesBuilder.String(),
	)

	h.Write([]byte(payload))
	return hex.EncodeToString(h.Sum(nil))
}
