package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// setupTestRepo cria um diretório temporário e inicializa um repositório .fit nele
func setupTestRepo(t *testing.T) (*Repository, string) {
	t.Helper()
	tempDir, err := os.MkdirTemp("", "fit-test-*")
	if err != nil {
		t.Fatalf("falha ao criar diretório temporário: %v", err)
	}

	repo := NewRepository(tempDir)
	if err := repo.Init(); err != nil {
		t.Fatalf("falha ao inicializar repo: %v", err)
	}

	return repo, tempDir
}

// TestFit_Init valida a criação da pasta .fit e suas estruturas internas
func TestFit_Init(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "fit-init-test-*")
	if err != nil {
		t.Fatalf("erro ao criar tempdir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	repo := NewRepository(tempDir)
	if err := repo.Init(); err != nil {
		t.Fatalf("Init falhou: %v", err)
	}

	// 1. Verifica pasta .fit
	if fi, err := os.Stat(repo.FitDir); err != nil || !fi.IsDir() {
		t.Fatalf("esperava pasta .fit existir, erro: %v", err)
	}

	// 2. Verifica subpasta objects
	objectsPath := filepath.Join(repo.FitDir, ObjectsDir)
	if fi, err := os.Stat(objectsPath); err != nil || !fi.IsDir() {
		t.Fatalf("esperava pasta objects existir, erro: %v", err)
	}

	// 3. Verifica arquivo HEAD
	headPath := filepath.Join(repo.FitDir, HeadFile)
	if _, err := os.Stat(headPath); err != nil {
		t.Fatalf("esperava arquivo HEAD existir, erro: %v", err)
	}

	// 4. Verifica arquivo index.json
	indexPath := filepath.Join(repo.FitDir, IndexFile)
	if _, err := os.Stat(indexPath); err != nil {
		t.Fatalf("esperava arquivo index.json existir, erro: %v", err)
	}

	// 5. Testar que não permite inicializar duas vezes
	if err := repo.Init(); err == nil {
		t.Fatalf("esperava erro ao tentar inicializar repositório já existente")
	}
}

// TestFit_Add valida a cópia de arquivos para .fit/objects/<hash> e a fácil recuperação
func TestFit_Add(t *testing.T) {
	repo, tempDir := setupTestRepo(t)
	defer os.RemoveAll(tempDir)

	testFileName := "arquivo_teste.txt"
	testContent := "Conteudo para teste do comando add"
	filePath := filepath.Join(tempDir, testFileName)
	if err := os.WriteFile(filePath, []byte(testContent), 0644); err != nil {
		t.Fatalf("erro ao criar arquivo: %v", err)
	}

	// Executa o add
	added, err := repo.Add(testFileName)
	if err != nil {
		t.Fatalf("falha no Add: %v", err)
	}
	if len(added) != 1 || added[0] != testFileName {
		t.Fatalf("esperava adicionar %s, obteve %v", testFileName, added)
	}

	// Verifica se foi registrado no index.json
	idx, err := LoadIndex(filepath.Join(repo.FitDir, IndexFile))
	if err != nil {
		t.Fatalf("falha ao carregar índice: %v", err)
	}

	entry, exists := idx.Entries[testFileName]
	if !exists {
		t.Fatalf("arquivo %s não encontrado no índice", testFileName)
	}

	// Verifica se a cópia existe fisicamente em .fit/objects/<hash>
	objectPath := repo.GetObjectFilePath(entry.Hash)
	copiedContent, err := os.ReadFile(objectPath)
	if err != nil {
		t.Fatalf("cópia do arquivo não encontrada em .fit/objects/%s: %v", entry.Hash, err)
	}

	// Garante que o conteúdo copiado é idêntico ao original
	if string(copiedContent) != testContent {
		t.Fatalf("conteúdo copiado difere do original. Esperado: '%s', obtido: '%s'", testContent, string(copiedContent))
	}
}

// TestFit_Commit valida criação do objeto de commit com todos os campos obrigatórios
func TestFit_Commit(t *testing.T) {
	repo, tempDir := setupTestRepo(t)
	defer os.RemoveAll(tempDir)

	// Cria arquivo e adiciona ao stage
	os.WriteFile(filepath.Join(tempDir, "codigo.go"), []byte("package main"), 0644)
	repo.Add("codigo.go")

	authorName := "Carlos Silva"
	authorEmail := "carlos@empresa.com"
	commitMsg := "Primeira versão do código"

	commit, err := repo.Commit(commitMsg, authorName, authorEmail)
	if err != nil {
		t.Fatalf("falha ao criar commit: %v", err)
	}

	// Valida campos do commit
	if commit.Hash == "" {
		t.Fatalf("esperava hash gerado para o commit")
	}
	if commit.AuthorName != authorName {
		t.Fatalf("esperava authorName '%s', obteve '%s'", authorName, commit.AuthorName)
	}
	if commit.AuthorEmail != authorEmail {
		t.Fatalf("esperava authorEmail '%s', obteve '%s'", authorEmail, commit.AuthorEmail)
	}
	if commit.Message != commitMsg {
		t.Fatalf("esperava mensagem '%s', obteve '%s'", commitMsg, commit.Message)
	}
	if commit.Timestamp.IsZero() {
		t.Fatalf("esperava timestamp preenchido no commit")
	}

	// Valida se o objeto foi gravado no disco em .fit/objects/<hash>
	loadedCommit, err := repo.GetCommit(commit.Hash)
	if err != nil {
		t.Fatalf("objeto de commit não encontrado no disco: %v", err)
	}
	if loadedCommit.Hash != commit.Hash {
		t.Fatalf("hash divergente do commit lido do disco")
	}

	// Valida se o HEAD aponta para este commit
	headHash, err := repo.GetHead()
	if err != nil {
		t.Fatalf("falha ao ler HEAD: %v", err)
	}
	if headHash != commit.Hash {
		t.Fatalf("HEAD não foi atualizado. Esperado: %s, obtido: %s", commit.Hash, headHash)
	}
}

// TestFit_Log_FiveCommitsLimit valida que o log percorre os 5 commits mais recentes em ordem
func TestFit_Log_FiveCommitsLimit(t *testing.T) {
	repo, tempDir := setupTestRepo(t)
	defer os.RemoveAll(tempDir)

	// Cria 7 commits sequenciais
	totalCommits := 7
	var createdHashes []string

	for i := 1; i <= totalCommits; i++ {
		time.Sleep(5 * time.Millisecond) // Garante timestamp crescente
		filename := fmt.Sprintf("arquivo_%d.txt", i)
		os.WriteFile(filepath.Join(tempDir, filename), []byte(fmt.Sprintf("conteudo %d", i)), 0644)
		repo.Add(filename)

		c, err := repo.Commit(fmt.Sprintf("Commit número %d", i), "Dev", "dev@fit.local")
		if err != nil {
			t.Fatalf("erro ao criar commit %d: %v", i, err)
		}
		createdHashes = append(createdHashes, c.Hash)
	}

	// Executa o log (deve trazer no máximo 5)
	commits, err := repo.Log(5)
	if err != nil {
		t.Fatalf("falha ao executar Log: %v", err)
	}

	if len(commits) != 5 {
		t.Fatalf("esperava exatamente 5 commits retornados, obteve %d", len(commits))
	}

	// O primeiro do log deve ser o mais recente (o 7º criado)
	if commits[0].Hash != createdHashes[6] {
		t.Fatalf("o commit mais recente no log deveria ser %s, obteve %s", createdHashes[6], commits[0].Hash)
	}

	// O último do log (5º) deve ser o 3º criado (pois 7, 6, 5, 4, 3)
	if commits[4].Hash != createdHashes[2] {
		t.Fatalf("o 5º commit no log deveria ser %s, obteve %s", createdHashes[2], commits[4].Hash)
	}

	// Valida ordenação estrita do mais recente para o mais antigo
	for i := 0; i < len(commits)-1; i++ {
		if commits[i].Timestamp.Before(commits[i+1].Timestamp) {
			t.Fatalf("ordem cronológica incorreta: commit[%d] é mais antigo que commit[%d]", i, i+1)
		}
	}
}

// TestFit_Status valida o status com arquivos em stage, modificados e não rastreados
func TestFit_Status(t *testing.T) {
	repo, tempDir := setupTestRepo(t)
	defer os.RemoveAll(tempDir)

	// 1. Arquivo não rastreado
	os.WriteFile(filepath.Join(tempDir, "untracked.txt"), []byte("nao rastreado"), 0644)
	st, err := repo.Status()
	if err != nil {
		t.Fatalf("erro ao consultar status: %v", err)
	}
	if len(st.UntrackedFiles) != 1 || st.UntrackedFiles[0] != "untracked.txt" {
		t.Fatalf("esperava 'untracked.txt' em UntrackedFiles, obteve %v", st.UntrackedFiles)
	}

	// 2. Adiciona ao stage
	repo.Add("untracked.txt")
	st, _ = repo.Status()
	if len(st.StagedFiles) != 1 {
		t.Fatalf("esperava 1 arquivo em StagedFiles, obteve %v", st.StagedFiles)
	}
	if len(st.UntrackedFiles) != 0 {
		t.Fatalf("não esperava arquivos não rastreados após add")
	}

	// 3. Comita o arquivo
	repo.Commit("adiciona untracked.txt", "Dev", "dev@fit.local")
	st, _ = repo.Status()
	if !st.Clean {
		t.Fatalf("esperava working tree limpa após commit")
	}

	// 4. Modifica o arquivo no disco sem adicionar ao stage
	os.WriteFile(filepath.Join(tempDir, "untracked.txt"), []byte("conteudo alterado"), 0644)
	st, _ = repo.Status()
	if len(st.ModifiedFiles) != 1 {
		t.Fatalf("esperava 1 arquivo em ModifiedFiles após alteração em disco, obteve %v", st.ModifiedFiles)
	}
}

// TestFit_ValidationErrors valida bloqueios para commit vazio
func TestFit_ValidationErrors(t *testing.T) {
	repo, tempDir := setupTestRepo(t)
	defer os.RemoveAll(tempDir)

	// Tentar commit sem nenhum arquivo em stage
	_, err := repo.Commit("mensagem qualquer", "Dev", "dev@fit.local")
	if err == nil {
		t.Fatalf("esperava erro ao commitar sem arquivos em stage")
	}

	// Tentar commit com mensagem vazia
	os.WriteFile(filepath.Join(tempDir, "teste.txt"), []byte("teste"), 0644)
	repo.Add("teste.txt")
	_, err = repo.Commit("   ", "Dev", "dev@fit.local")
	if err == nil {
		t.Fatalf("esperava erro ao commitar com mensagem vazia")
	}
}
