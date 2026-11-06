# Diretrizes para Agentes de IA (AGENTS.md)

Este repositório é um monorepo para práticas, desafios e pequenos projetos em **Go (Golang)**.

---

## 🏛️ Estrutura e Organização

- **Isolamento de Projetos**: Cada novo desafio ou projeto deve ficar em sua própria pasta independente (ex: `lru-cache/`, `rate-limiter/`, etc.).
- **Módulo Independente**: Cada subprojeto possui seu próprio `go.mod`, código-fonte, testes e seu próprio `README.md` explicativo.
- **Índice Global**: Sempre que um novo projeto for criado, adicione uma entrada descritiva no [`README.md`](./README.md) da raiz.
- **Não Interferência**: Não modifique projetos já concluídos sem que o usuário peça explicitamente.

---

## 💻 Padrões de Código e Desenvolvimento

- **Go Idiomático**:
  - Prefira a biblioteca padrão (`standard library`) do Go sempre que possível antes de introduzir dependências externas.
  - Siga convenções de nomenclatura idiomáticas do Go (nomes de variáveis curtos e expressivos, `camelCase`, `PascalCase` para exportados).
  - Sempre formate o código com `gofmt -w -s .`.
- **Didática e Documentação**:
  - Mantenha comentários explicativos em português explicando o porquê das decisões arquiteturais e estruturas de dados adotadas.
  - Forneça uma demonstração visual clara no arquivo `main.go`.

---

## 🧪 Testes e Qualidade

- **Testes Obrigatórios**: Todo projeto deve incluir testes unitários abrangentes (`*_test.go`).
- **Verificação Contínua**: Antes de concluir qualquer tarefa, execute os testes e garanta que todos passem:
  ```bash
  cd <pasta-do-projeto>
  go test -v ./...
  ```

---

## 🛠️ Comandos Frequentes

| Ação | Comando |
|---|---|
| Executar demonstração | `go run .` |
| Executar testes unitários | `go test -v ./...` |
| Formatar código | `gofmt -w -s .` |
| Criar novo módulo | `go mod init <nome-do-modulo>` |
