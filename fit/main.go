package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		printHelp()
		return
	}

	command := os.Args[1]

	switch command {
	case "init":
		cmdInit()
	case "add":
		cmdAdd(os.Args[2:])
	case "commit":
		cmdCommit(os.Args[2:])
	case "status":
		cmdStatus()
	case "log":
		cmdLog()
	case "demo":
		runDemonstration()
	case "help", "--help", "-h":
		printHelp()
	default:
		fmt.Printf("Comando desconhecido: '%s'\n", command)
		printHelp()
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Println("==================================================")
	fmt.Println("              FIT - CLONE DO GIT EM GO            ")
	fmt.Println("==================================================")
	fmt.Println("Uso: fit <comando> [argumentos...]")
	fmt.Println("\nComandos disponíveis:")
	fmt.Println("  init                  Inicializa um novo repositório .fit")
	fmt.Println("  add <arquivo...>      Copia o arquivo para .fit e adiciona ao stage")
	fmt.Println("  commit -m \"mensagem\"  Salva um commit com os arquivos em stage")
	fmt.Println("  status                Exibe o status dos arquivos (stage, modificados, untracked)")
	fmt.Println("  log                   Exibe o resumo dos 5 commits mais recentes")
	fmt.Println("  demo                  Executa uma demonstração guiada completa")
	fmt.Println("==================================================")
}

func cmdInit() {
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("Erro ao obter diretório atual: %v\n", err)
		os.Exit(1)
	}

	repo := NewRepository(currentDir)
	if err := repo.Init(); err != nil {
		fmt.Printf("Erro ao inicializar repositório: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Repositório Fit inicializado com sucesso em: %s\n", filepath.Join(currentDir, FitDirName))
}

func cmdAdd(args []string) {
	if len(args) == 0 {
		fmt.Println("Uso: fit add <arquivo1> [arquivo2... ou .]")
		os.Exit(1)
	}

	repo, err := FindRepository()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	added, err := repo.Add(args...)
	if err != nil {
		fmt.Printf("Erro no comando add: %v\n", err)
		os.Exit(1)
	}

	if len(added) == 0 {
		fmt.Println("Nenhum arquivo elegível para adicionar.")
		return
	}

	for _, file := range added {
		fmt.Printf("✓ Adicionado ao stage: %s\n", file)
	}
}

func cmdCommit(args []string) {
	commitFlags := flag.NewFlagSet("commit", flag.ExitOnError)
	msgFlag := commitFlags.String("m", "", "Mensagem do commit")
	authorFlag := commitFlags.String("author", "", "Nome do autor (opcional)")
	emailFlag := commitFlags.String("email", "", "Email do autor (opcional)")

	if err := commitFlags.Parse(args); err != nil {
		os.Exit(1)
	}

	message := strings.TrimSpace(*msgFlag)
	if message == "" {
		fmt.Println("Erro: informe uma mensagem para o commit usando -m \"sua mensagem\"")
		os.Exit(1)
	}

	repo, err := FindRepository()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	commit, err := repo.Commit(message, *authorFlag, *emailFlag)
	if err != nil {
		fmt.Printf("Erro ao criar commit: %v\n", err)
		os.Exit(1)
	}

	shortHash := commit.Hash
	if len(shortHash) > 7 {
		shortHash = shortHash[:7]
	}

	fmt.Printf("[commit %s] %s\n", shortHash, commit.Message)
	fmt.Printf("  Autor: %s <%s>\n", commit.AuthorName, commit.AuthorEmail)
	fmt.Printf("  Data:  %s\n", commit.Timestamp.Format("02/01/2006 15:04:05"))
	fmt.Printf("  Arquivos inclusos: %d\n", len(commit.Files))
}

func cmdStatus() {
	repo, err := FindRepository()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	status, err := repo.Status()
	if err != nil {
		fmt.Printf("Erro ao consultar status: %v\n", err)
		os.Exit(1)
	}

	headHash, _ := repo.GetHead()
	if headHash == "" {
		fmt.Println("No branch principal (nenhum commit realizado ainda)")
		fmt.Println()
	} else {
		short := headHash
		if len(short) > 7 {
			short = short[:7]
		}
		fmt.Printf("No branch principal (último commit: %s)\n\n", short)
	}

	if status.Clean {
		fmt.Println("Nada para commitar, working tree limpa.")
		return
	}

	// Mudanças em Stage
	if len(status.StagedFiles) > 0 {
		fmt.Println("Mudanças prontas para serem commitadas:")
		fmt.Println("  (use \"fit commit -m <mensagem>\" para gravar)")
		for _, f := range status.StagedFiles {
			fmt.Printf("    \033[32m%s\033[0m\n", f)
		}
		fmt.Println()
	}

	// Arquivos modificados fora de Stage
	if len(status.ModifiedFiles) > 0 {
		fmt.Println("Mudanças não preparadas para commit:")
		fmt.Println("  (use \"fit add <arquivo>...\" para atualizar o stage)")
		for _, f := range status.ModifiedFiles {
			fmt.Printf("    \033[31m%s\033[0m\n", f)
		}
		fmt.Println()
	}

	// Arquivos não rastreados
	if len(status.UntrackedFiles) > 0 {
		fmt.Println("Arquivos não rastreados:")
		fmt.Println("  (use \"fit add <arquivo>...\" para incluir no commit)")
		for _, f := range status.UntrackedFiles {
			fmt.Printf("    \033[31m%s\033[0m\n", f)
		}
		fmt.Println()
	}
}

func cmdLog() {
	repo, err := FindRepository()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Requisito: percorre os 5 commits em ordem do mais atual pro mais antigo
	commits, err := repo.Log(5)
	if err != nil {
		fmt.Printf("Erro ao consultar log: %v\n", err)
		os.Exit(1)
	}

	if len(commits) == 0 {
		fmt.Println("Nenhum commit encontrado no repositório.")
		return
	}

	fmt.Printf("Exibindo os %d commit(s) mais recentes (do mais novo ao mais antigo):\n\n", len(commits))

	for i, c := range commits {
		fmt.Printf("\033[33mcommit %s\033[0m (índice: %d)\n", c.Hash, i+1)
		if c.ParentHash != "" {
			fmt.Printf("Parent: %s\n", c.ParentHash)
		}
		fmt.Printf("Author: %s <%s>\n", c.AuthorName, c.AuthorEmail)
		fmt.Printf("Date:   %s\n\n", c.Timestamp.Format("Mon Jan 02 15:04:05 2006 -0700"))
		fmt.Printf("    %s\n\n", c.Message)
		fmt.Printf("    Arquivos neste commit (%d):\n", len(c.Files))
		for file, hash := range c.Files {
			shortH := hash
			if len(shortH) > 10 {
				shortH = shortH[:10]
			}
			fmt.Printf("      - %s (hash: %s)\n", file, shortH)
		}
		fmt.Println(strings.Repeat("-", 60))
	}
}

// runDemonstration executa um fluxo completo passo a passo em um diretório temporário
func runDemonstration() {
	fmt.Println("==================================================")
	fmt.Println("        DEMONSTRAÇÃO DO CLONE DO GIT (FIT)       ")
	fmt.Println("==================================================")

	tempDir, err := os.MkdirTemp("", "fit-demo-*")
	if err != nil {
		fmt.Println("Erro ao criar diretório temporário:", err)
		return
	}
	defer os.RemoveAll(tempDir)

	fmt.Printf("\n[1] Inicializando repositório Fit em %s...\n", tempDir)
	repo := NewRepository(tempDir)
	if err := repo.Init(); err != nil {
		fmt.Println("Falha no init:", err)
		return
	}
	fmt.Println("✓ Repositório inicializado com pasta .fit criada!")

	// 2. Criando arquivos de teste
	fmt.Println("\n[2] Criando arquivos 'main.go' e 'README.md'...")
	os.WriteFile(filepath.Join(tempDir, "main.go"), []byte("package main\n\nfunc main() {}\n"), 0644)
	os.WriteFile(filepath.Join(tempDir, "README.md"), []byte("# Meu Projeto\n"), 0644)

	// Status antes do add
	status, _ := repo.Status()
	fmt.Printf("Arquivos não rastreados detectados pelo status: %v\n", status.UntrackedFiles)

	// 3. Executando fit add
	fmt.Println("\n[3] Executando 'fit add main.go README.md'...")
	added, _ := repo.Add("main.go", "README.md")
	fmt.Printf("✓ Arquivos adicionados à staging area: %v\n", added)

	// 4. Executando fit commit
	fmt.Println("\n[4] Executando 'fit commit -m \"commit inicial\"'...")
	c1, err := repo.Commit("commit inicial", "Alice Silva", "alice@email.com")
	if err != nil {
		fmt.Println("Erro no commit:", err)
		return
	}
	fmt.Printf("✓ Commit gravado com sucesso! Hash: %s\n", c1.Hash)

	// 5. Criando múltiplos commits para testar o limite de 5 commits no log
	fmt.Println("\n[5] Criando mais 5 commits sequenciais para validar o histórico...")
	for i := 2; i <= 6; i++ {
		time.Sleep(10 * time.Millisecond)
		fileName := fmt.Sprintf("modulo_%d.go", i)
		content := fmt.Sprintf("package modulo%d\n// versao %d\n", i, i)
		os.WriteFile(filepath.Join(tempDir, fileName), []byte(content), 0644)
		repo.Add(fileName)
		msg := fmt.Sprintf("Adiciona módulo %d", i)
		c, err := repo.Commit(msg, "Alice Silva", "alice@email.com")
		if err != nil {
			fmt.Printf("Erro no commit %d: %v\n", i, err)
			return
		}
		fmt.Printf("  -> Commit %d criado: [%s] %s\n", i, c.Hash[:7], c.Message)
	}

	// 6. Testando fit log (deve listar no máximo os 5 mais recentes)
	fmt.Println("\n[6] Executando 'fit log' (percorre os 5 commits mais recentes)...")
	commits, err := repo.Log(5)
	if err != nil {
		fmt.Println("Erro ao ler log:", err)
		return
	}
	fmt.Printf("Total de commits retornados pelo log: %d (o repositório tem 6 no total)\n\n", len(commits))
	for i, c := range commits {
		fmt.Printf("  [%d] %s (%s) - \"%s\"\n", i+1, c.Hash[:8], c.Timestamp.Format("15:04:05"), c.Message)
	}

	fmt.Println("\n==================================================")
	fmt.Println("      DEMONSTRAÇÃO CONCLUÍDA COM SUCESSO!         ")
	fmt.Println("==================================================")
}
