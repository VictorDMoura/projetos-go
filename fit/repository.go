package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	FitDirName   = ".fit"
	ObjectsDir   = "objects"
	IndexFile    = "index.json"
	HeadFile     = "HEAD"
	ConfigFile   = "config.json"
	DefaultLimit = 5
)

// Config armazena configurações locais do repositório Fit (autor padrão, etc.)
type Config struct {
	AuthorName  string `json:"author_name"`
	AuthorEmail string `json:"author_email"`
}

// StatusResult consolida o estado do repositório para exibição no comando 'status'.
type StatusResult struct {
	StagedFiles    []string // Arquivos prontos para commit (em stage)
	ModifiedFiles  []string // Arquivos modificados na working tree mas não adicionados
	UntrackedFiles []string // Arquivos novos não rastreados
	Clean          bool     // Indica se a working tree está totalmente limpa
}

// Repository gerencia as operações no repositório .fit.
type Repository struct {
	RootDir string // Diretório raiz do projeto trabalhado
	FitDir  string // Caminho para a pasta .fit
}

// NewRepository instancia um repositório apontando para o diretório informado.
func NewRepository(rootDir string) *Repository {
	return &Repository{
		RootDir: rootDir,
		FitDir:  filepath.Join(rootDir, FitDirName),
	}
}

// FindRepository procura a pasta .fit a partir do diretório atual subindo na árvore.
func FindRepository() (*Repository, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	dir := currentDir
	for {
		fitPath := filepath.Join(dir, FitDirName)
		if fi, err := os.Stat(fitPath); err == nil && fi.IsDir() {
			return NewRepository(dir), nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Chegou na raiz do sistema operacional sem encontrar .fit
			break
		}
		dir = parent
	}

	return nil, errors.New("erro: nenhum repositório .fit encontrado (execute 'fit init' primeiro)")
}

// ============================================================================
// COMANDO: fit init
// ============================================================================

// Init cria a pasta .fit e todos os arquivos e subpastas necessárias:
// - .fit/objects: armazena cópias dos arquivos (blobs) e objetos de commit indexados por hash
// - .fit/HEAD: armazena a referência para o commit atual mais recente
// - .fit/index.json: armazena o estado dos arquivos adicionados (staging area)
// - .fit/config.json: armazena configurações do repositório (ex: nome e email do autor)
func (r *Repository) Init() error {
	if _, err := os.Stat(r.FitDir); err == nil {
		return fmt.Errorf("o repositório .fit já existe em %s", r.FitDir)
	}

	// 1. Cria a pasta raiz .fit
	if err := os.MkdirAll(r.FitDir, 0755); err != nil {
		return fmt.Errorf("falha ao criar pasta %s: %w", r.FitDir, err)
	}

	// 2. Cria a subpasta .fit/objects
	objectsPath := filepath.Join(r.FitDir, ObjectsDir)
	if err := os.MkdirAll(objectsPath, 0755); err != nil {
		return fmt.Errorf("falha ao criar pasta de objetos: %w", err)
	}

	// 3. Cria o arquivo .fit/HEAD inicialmente vazio
	headPath := filepath.Join(r.FitDir, HeadFile)
	if err := os.WriteFile(headPath, []byte(""), 0644); err != nil {
		return fmt.Errorf("falha ao criar arquivo HEAD: %w", err)
	}

	// 4. Cria o arquivo .fit/index.json inicial
	indexPath := filepath.Join(r.FitDir, IndexFile)
	idx := NewIndex()
	if err := idx.Save(indexPath); err != nil {
		return fmt.Errorf("falha ao criar índice inicial: %w", err)
	}

	// 5. Cria configuração inicial padrão
	cfgPath := filepath.Join(r.FitDir, ConfigFile)
	defaultCfg := Config{
		AuthorName:  "Desenvolvedor Fit",
		AuthorEmail: "dev@fit.local",
	}
	cfgData, _ := json.MarshalIndent(defaultCfg, "", "  ")
	if err := os.WriteFile(cfgPath, cfgData, 0644); err != nil {
		return fmt.Errorf("falha ao criar arquivo de configuração: %w", err)
	}

	return nil
}

// ============================================================================
// COMANDO: fit add
// ============================================================================

// Add copia os arquivos informados para dentro da pasta .fit/objects/ e registra no índice.
// A cópia é identificada pelo seu hash SHA-1 para que qualquer comando possa localizá-la
// facilmente com busca O(1) através do hash contido no índice (.fit/index.json).
func (r *Repository) Add(targetPaths ...string) ([]string, error) {
	indexPath := filepath.Join(r.FitDir, IndexFile)
	idx, err := LoadIndex(indexPath)
	if err != nil {
		return nil, fmt.Errorf("falha ao carregar índice: %w", err)
	}

	var addedFiles []string

	for _, target := range targetPaths {
		absTarget := target
		if !filepath.IsAbs(target) {
			absTarget = filepath.Join(r.RootDir, target)
		}

		info, err := os.Stat(absTarget)
		if err != nil {
			return nil, fmt.Errorf("arquivo ou pasta não encontrada '%s': %w", target, err)
		}

		if info.IsDir() {
			// Se for um diretório, percorre recursivamente
			err = filepath.WalkDir(absTarget, func(path string, d fs.DirEntry, walkErr error) error {
				if walkErr != nil {
					return walkErr
				}
				// Ignora a própria pasta .fit e a pasta .git
				if d.IsDir() {
					if d.Name() == FitDirName || d.Name() == ".git" {
						return filepath.SkipDir
					}
					return nil
				}
				added, err := r.addSingleFile(path, idx)
				if err != nil {
					return err
				}
				if added != "" {
					addedFiles = append(addedFiles, added)
				}
				return nil
			})
			if err != nil {
				return nil, err
			}
		} else {
			// Adiciona arquivo único
			added, err := r.addSingleFile(absTarget, idx)
			if err != nil {
				return nil, err
			}
			if added != "" {
				addedFiles = append(addedFiles, added)
			}
		}
	}

	// Salva o índice atualizado em .fit/index.json
	if err := idx.Save(indexPath); err != nil {
		return nil, fmt.Errorf("falha ao salvar índice: %w", err)
	}

	return addedFiles, nil
}

// addSingleFile copia o arquivo para .fit/objects/<hash> e atualiza a entrada no índice.
func (r *Repository) addSingleFile(absPath string, idx *Index) (string, error) {
	relPath, err := filepath.Rel(r.RootDir, absPath)
	if err != nil {
		return "", err
	}
	relPath = filepath.ToSlash(relPath)

	// Não adiciona arquivos que estejam dentro de .fit
	if strings.HasPrefix(relPath, FitDirName) || strings.HasPrefix(relPath, ".git") {
		return "", nil
	}

	// 1. Lê o conteúdo do arquivo
	content, err := os.ReadFile(absPath)
	if err != nil {
		return "", fmt.Errorf("erro ao ler arquivo '%s': %w", relPath, err)
	}

	// 2. Calcula o hash SHA-1 do conteúdo
	hash := HashBytes(content)

	// 3. Copia o arquivo para a pasta .fit/objects/<hash>
	objectPath := filepath.Join(r.FitDir, ObjectsDir, hash)
	if err := os.WriteFile(objectPath, content, 0644); err != nil {
		return "", fmt.Errorf("falha ao copiar arquivo para objetos: %w", err)
	}

	// 4. Registra no índice (staging area)
	idx.Entries[relPath] = IndexEntry{
		Path: relPath,
		Hash: hash,
		Size: int64(len(content)),
	}

	return relPath, nil
}

// ============================================================================
// COMANDO: fit commit
// ============================================================================

// Commit cria e salva um objeto Commit na pasta .fit/objects/<commitHash> e atualiza o HEAD.
//
// Requisitos atendidos:
// - Possui data/timestamp
// - Nome e email do autor
// - Mensagem descritiva
// - Código hash único
func (r *Repository) Commit(message string, authorName, authorEmail string) (*Commit, error) {
	message = strings.TrimSpace(message)
	if message == "" {
		return nil, errors.New("a mensagem do commit não pode ser vazia (use -m \"mensagem\")")
	}

	// 1. Carrega o índice (arquivos que foram adicionados via 'fit add')
	indexPath := filepath.Join(r.FitDir, IndexFile)
	idx, err := LoadIndex(indexPath)
	if err != nil {
		return nil, fmt.Errorf("erro ao carregar staging area: %w", err)
	}

	if len(idx.Entries) == 0 {
		return nil, errors.New("nada para commitar (nenhum arquivo em stage, execute 'fit add' antes)")
	}

	// 2. Resolve o autor caso não informado
	cfg := r.GetConfig()
	if authorName == "" {
		authorName = cfg.AuthorName
	}
	if authorEmail == "" {
		authorEmail = cfg.AuthorEmail
	}

	// 3. Lê o commit pai (se houver) a partir do HEAD
	headCommitHash, err := r.GetHead()
	if err != nil {
		return nil, err
	}

	// Se houver commit anterior, verifica se o índice é idêntico ao último commit
	if headCommitHash != "" {
		parentCommit, err := r.GetCommit(headCommitHash)
		if err == nil && parentCommit != nil {
			if r.isIndexIdenticalToCommit(idx, parentCommit) {
				return nil, errors.New("nada para commitar (nenhuma alteração detectada em relação ao último commit)")
			}
		}
	}

	// 4. Constrói o mapa de arquivos do commit
	filesMap := make(map[string]string, len(idx.Entries))
	for path, entry := range idx.Entries {
		filesMap[path] = entry.Hash
	}

	// 5. Monta o objeto Commit
	commit := &Commit{
		ParentHash:  headCommitHash,
		AuthorName:  authorName,
		AuthorEmail: authorEmail,
		Timestamp:   time.Now(),
		Message:     message,
		Files:       filesMap,
	}

	// 6. Calcula o hash único
	commit.Hash = commit.CalculateHash()

	// 7. Salva o objeto Commit na pasta .fit/objects/<hash>
	commitData, err := json.MarshalIndent(commit, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("falha ao serializar commit: %w", err)
	}

	commitPath := filepath.Join(r.FitDir, ObjectsDir, commit.Hash)
	if err := os.WriteFile(commitPath, commitData, 0644); err != nil {
		return nil, fmt.Errorf("falha ao salvar objeto de commit: %w", err)
	}

	// 8. Atualiza o arquivo HEAD apontando para o novo commit
	headPath := filepath.Join(r.FitDir, HeadFile)
	if err := os.WriteFile(headPath, []byte(commit.Hash), 0644); err != nil {
		return nil, fmt.Errorf("falha ao atualizar HEAD: %w", err)
	}

	return commit, nil
}

// ============================================================================
// COMANDO: fit log
// ============================================================================

// Log percorre os commits em ordem cronológica reversa (do mais recente ao mais antigo)
// respeitando o limite especificado (por padrão os 5 últimos commits).
func (r *Repository) Log(limit int) ([]*Commit, error) {
	if limit <= 0 {
		limit = DefaultLimit
	}

	headHash, err := r.GetHead()
	if err != nil {
		return nil, err
	}

	if headHash == "" {
		return []*Commit{}, nil // Repositório recém-criado, sem commits
	}

	var commits []*Commit
	currentHash := headHash

	// Percorre a cadeia de commits pelo ParentHash
	for currentHash != "" && len(commits) < limit {
		commit, err := r.GetCommit(currentHash)
		if err != nil {
			return nil, fmt.Errorf("falha ao carregar commit %s: %w", currentHash, err)
		}

		commits = append(commits, commit)
		currentHash = commit.ParentHash
	}

	return commits, nil
}

// ============================================================================
// COMANDO: fit status
// ============================================================================

// Status compara o estado do repositório em três níveis:
// 1. HEAD (último commit)
// 2. Index (staging area)
// 3. Working Tree (arquivos no disco)
func (r *Repository) Status() (*StatusResult, error) {
	res := &StatusResult{}

	// Carrega índice
	indexPath := filepath.Join(r.FitDir, IndexFile)
	idx, err := LoadIndex(indexPath)
	if err != nil {
		return nil, err
	}

	// Carrega commit do HEAD
	var headFiles map[string]string
	headHash, _ := r.GetHead()
	if headHash != "" {
		headCommit, err := r.GetCommit(headHash)
		if err == nil && headCommit != nil {
			headFiles = headCommit.Files
		}
	}
	if headFiles == nil {
		headFiles = make(map[string]string)
	}

	// 1. Arquivos em Stage (Index vs HEAD)
	for path, entry := range idx.Entries {
		headHashVal, exists := headFiles[path]
		if !exists {
			res.StagedFiles = append(res.StagedFiles, fmt.Sprintf("novo arquivo:   %s", path))
		} else if headHashVal != entry.Hash {
			res.StagedFiles = append(res.StagedFiles, fmt.Sprintf("modificado:     %s", path))
		}
	}

	// 2. Arquivos da Working Tree vs Index (Modificados / Não rastreados)
	workingFiles := make(map[string]string)
	err = filepath.WalkDir(r.RootDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			if d.Name() == FitDirName || d.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}

		relPath, err := filepath.Rel(r.RootDir, path)
		if err != nil {
			return err
		}
		relPath = filepath.ToSlash(relPath)

		h, err := HashFile(path)
		if err != nil {
			return err
		}
		workingFiles[relPath] = h
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Checa arquivos modificados na working tree
	for path, entry := range idx.Entries {
		workHash, exists := workingFiles[path]
		if !exists {
			res.ModifiedFiles = append(res.ModifiedFiles, fmt.Sprintf("deletado:       %s", path))
		} else if workHash != entry.Hash {
			res.ModifiedFiles = append(res.ModifiedFiles, fmt.Sprintf("modificado:     %s", path))
		}
	}

	// Checa arquivos não rastreados (Untracked)
	for path := range workingFiles {
		if _, inIndex := idx.Entries[path]; !inIndex {
			if _, inHead := headFiles[path]; !inHead {
				res.UntrackedFiles = append(res.UntrackedFiles, path)
			}
		}
	}

	if len(res.StagedFiles) == 0 && len(res.ModifiedFiles) == 0 && len(res.UntrackedFiles) == 0 {
		res.Clean = true
	}

	return res, nil
}

// ============================================================================
// Métodos Auxiliares
// ============================================================================

// GetHead retorna o hash do commit apontado por HEAD.
func (r *Repository) GetHead() (string, error) {
	headPath := filepath.Join(r.FitDir, HeadFile)
	data, err := os.ReadFile(headPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

// GetCommit lê e desserializa um commit a partir de .fit/objects/<hash>.
func (r *Repository) GetCommit(hash string) (*Commit, error) {
	commitPath := filepath.Join(r.FitDir, ObjectsDir, hash)
	data, err := os.ReadFile(commitPath)
	if err != nil {
		return nil, err
	}

	var commit Commit
	if err := json.Unmarshal(data, &commit); err != nil {
		return nil, err
	}
	return &commit, nil
}

// GetObjectFilePath retorna o caminho físico de um blob copiado para dentro de .fit.
func (r *Repository) GetObjectFilePath(hash string) string {
	return filepath.Join(r.FitDir, ObjectsDir, hash)
}

// ReadObjectContent lê os bytes de um objeto copiado para .fit/objects/<hash>.
func (r *Repository) ReadObjectContent(hash string) ([]byte, error) {
	return os.ReadFile(r.GetObjectFilePath(hash))
}

// GetConfig obtém as configurações do repositório em .fit/config.json.
func (r *Repository) GetConfig() Config {
	cfgPath := filepath.Join(r.FitDir, ConfigFile)
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		return Config{AuthorName: "Desenvolvedor Fit", AuthorEmail: "dev@fit.local"}
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{AuthorName: "Desenvolvedor Fit", AuthorEmail: "dev@fit.local"}
	}
	return cfg
}

func (r *Repository) isIndexIdenticalToCommit(idx *Index, c *Commit) bool {
	if len(idx.Entries) != len(c.Files) {
		return false
	}
	for path, entry := range idx.Entries {
		commitHash, exists := c.Files[path]
		if !exists || commitHash != entry.Hash {
			return false
		}
	}
	return true
}

// CopyFile helper para cópia de arquivos direta
func CopyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err = io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}
