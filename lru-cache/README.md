# Cache LRU (Least Recently Used) em Go

Implementação didática e de alta performance de um algoritmo de **Cache LRU** em Go.

---

## 📌 O que é um Cache LRU?

**LRU** significa *Least Recently Used* (Menos Recentemente Usado). É uma política de descarte de dados com capacidade limitada:
- Quando o cache atinge a capacidade máxima e um novo item precisa ser inserido, o cache descarta o item que não é acessado há mais tempo.
- Sempre que um item é acessado (`Get`) ou atualizado (`Set`), ele é considerado o mais recente (*Most Recently Used - MRU*).

---

## 🏗️ Como foi implementado? (Complexidade $O(1)$)

Para obter complexidade de tempo constante **$O(1)$** tanto na leitura (`Get`) quanto na escrita/descarte (`Set`), combinamos duas estruturas de dados:

1. **Hash Map (`map[string]*node`)**:
   - Permite encontrar qualquer elemento no cache em tempo constante $O(1)$ a partir da sua chave.

2. **Lista Duplamente Encadeada (`node` com `prev` e `next`)**:
   - Mantém a ordem cronológica de acesso.
   - Permite remover ou mover qualquer nó da lista para o início em tempo $O(1)$ (sem a necessidade de percorrer arrays ou fazer deslocamento de índices).
   - O nó armazena tanto a `key` quanto o `val`. Armazenar a `key` no próprio nó é crucial para que, ao descartar o nó da ponta mais antiga (`tail`), saibamos exatamente qual chave deletar do Hash Map em $O(1)$.

3. **Nós Sentinelas (`head` e `tail`)**:
   - São nós fictícios nas extremidades que simplificam a lógica de inserção e remoção, eliminando condições especiais de ponteiros nulos (`nil`).
   - `head.next`: Aponta sempre para o item mais recentemente usado (**MRU**).
   - `tail.prev`: Aponta sempre para o item menos recentemente usado (**LRU**).

```text
[head] <---> [Item Mais Recente] <---> ... <---> [Item Mais Antigo] <---> [tail]
  ^                                                      ^
  |                                                      |
Inserções / Leituras aqui                   Descarte por capacidade aqui
```

---

## 🛠️ Métodos Implementados

| Método | Assinatura | Descrição |
|---|---|---|
| **Construtor** | `NewLRUCache(capacity int) *LRUCache` | Inicializa o cache com a capacidade informada e conecta os nós sentinelas. |
| **Get** | `(c *LRUCache) Get(key string) any` | Retorna o valor associado à chave. Se não existir, retorna `-1`. Se existir, move o item para o início (mais recente). |
| **Set** | `(c *LRUCache) Set(key string, value any)` | Salva ou atualiza a chave. Se o cache estiver cheio, descarta o item mais antigo antes de inserir o novo. |

---

## 🚀 Como Executar

### 1. Executar a Demonstração
Dentro da pasta `lru-cache`:
```bash
go run .
```

### 2. Executar os Testes Unitários
```bash
go test -v ./...
```

---

## 🧪 Casos de Teste Cobertos

- [x] Retorno de `-1` para chaves não encontradas.
- [x] Inserção e recuperação básica de chave e valor.
- [x] Descarte automático do item menos usado ao estourar a capacidade.
- [x] Atualização de ordem de acesso após chamada de `Get(key)`.
- [x] Atualização de chave existente com novo valor sem alterar a contagem total de itens.
- [x] Suporte a tipos variados (`int`, `string`, `struct`, `slices`).
- [x] Tratamento de capacidades extremas (capacidade 1, 0 ou negativa).
