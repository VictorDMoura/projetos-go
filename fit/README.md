# Fit - Clone do Git em Go

Um clone simplificado e didático do **Git** escrito em Go, implementando o mecanismo de controle de versão através da pasta oculta `.fit`.

---

## 📌 Arquitetura Interna (`.fit`)

O programa cria e manipula a pasta `.fit`, que organiza os dados da seguinte forma:

```text
.fit/
├── config.json         # Configurações do repositório (autor padrão, email)
├── HEAD                # Referência contendo o hash do commit mais recente
├── index.json          # Staging Area: mapeia caminhos dos arquivos para seus hashes
└── objects/            # Armazenamento endereçado por conteúdo (Blobs e Commits)
    ├── <sha1-blob>     # Cópia do arquivo adicionado
    └── <sha1-commit>   # Objeto JSON do commit
```

---

## 🛠️ Comandos Implementados

### 1. `fit init`
- Cria a pasta `.fit` e todas as subpastas e arquivos necessários (`objects/`, `HEAD`, `index.json`, `config.json`).
- Previne reinicialização acidental se o repositório já existir.

### 2. `fit add <arquivo...>`
- Lê o arquivo no disco, calcula seu hash criptográfico **SHA-1** e **copia o arquivo** para `.fit/objects/<hash>`.
- Registra a entrada no índice `.fit/index.json` (caminho, hash e tamanho).
- Isso permite que qualquer outro comando encontre o arquivo instantaneamente em $O(1)$ através do hash.

### 3. `fit commit -m "<mensagem>"`
- Salva um objeto `Commit` serializado em `.fit/objects/<hash>`.
- O commit contém:
  - **Data/Hora**: Timestamp UTC com precisão de nanossegundos.
  - **Autor**: Nome e e-mail (obtidos via flag `--author`/`--email` ou `.fit/config.json`).
  - **Mensagem**: Descrição das mudanças fornecida via `-m`.
  - **Hash Único**: Calculado via SHA-1 combinando autor, data, mensagem, arquivos e o hash do commit anterior (`ParentHash`).
  - **Snapshot de Arquivos**: Mapeamento de cada arquivo para seu respectivo blob.
- Atualiza o ponteiro `HEAD` com o novo hash gerado.

### 4. `fit status`
- Compara três camadas de estado:
  1. **Working Tree**: Arquivos reais no sistema de arquivos.
  2. **Staging Area (`index.json`)**: Arquivos adicionados com `fit add`.
  3. **HEAD**: Snapshot do último commit gravado.
- Exibe de forma colorida e organizada:
  - Arquivos preparados para commit (Staged).
  - Arquivos modificados fora do stage.
  - Arquivos novos não rastreados (Untracked).

### 5. `fit log`
- Percorre a cadeia de commits a partir do `HEAD` seguindo o `ParentHash`.
- Retorna e imprime o resumo dos **5 commits mais recentes** em ordem cronológica reversa (do mais atual para o mais antigo).

---

## 🚀 Como Usar

### Executar a Demonstração Guiada
Para rodar uma demonstração completa automatizada:
```bash
go run . demo
```

### Usar como CLI
```bash
# 1. Inicializar
go run . init

# 2. Criar arquivos de teste
echo "Olá Fit" > arquivo.txt

# 3. Consultar status
go run . status

# 4. Adicionar arquivos (copia para .fit e prepara para o commit)
go run . add arquivo.txt

# 5. Criar commit
go run . commit -m "Meu primeiro commit no Fit"

# 6. Visualizar histórico (até 5 commits mais recentes)
go run . log
```

---

## 🧪 Testes Automatizados

Para executar toda a suíte de testes unitários:
```bash
go test -v ./...
```

Cenários cobertos pelos testes:
- [x] Criação correta de diretórios e arquivos em `fit init`.
- [x] Cópia fiel de arquivos para `.fit/objects/<hash>` e fácil recuperação em `fit add`.
- [x] Criação de objeto `Commit` com data, autor, mensagem, snapshot e hash único.
- [x] Limitação estrita aos 5 commits mais recentes em ordem cronológica decrescente em `fit log`.
- [x] Detecção de arquivos untracked, modificados e staged em `fit status`.
- [x] Bloqueio contra commits vazios ou sem mensagem.
