package main

import (
	"fmt"
	"strings"
)

// node representa um nó da nossa lista duplamente encadeada.
//
// Por que precisamos de uma lista duplamente encadeada?
//  1. Cada nó tem ponteiros para o elemento anterior (prev) e o próximo (next).
//  2. Isso permite remover ou mover qualquer nó para o início da lista em tempo O(1)
//     (tempo constante), sem precisar percorrer a lista.
//  3. Guardamos a "key" dentro do nó para que, ao remover o item mais antigo da cauda (tail),
//     possamos saber qual chave remover do mapa de busca (hash map) em O(1).
type node struct {
	key  string
	val  any
	prev *node
	next *node
}

// LRUCache representa a estrutura do nosso Cache LRU (Least Recently Used).
//
// Como funciona a arquitetura:
// - Combinamos uma Hash Map (map[string]*node) com uma Lista Duplamente Encadeada.
// - O Hash Map nos dá busca por chave em O(1).
// - A Lista Encadeada mantém a ordem de uso dos elementos em O(1).
// - Usamos "nós sentinelas" (head e tail) fictícios:
//   - O nó mais recentemente acessado/inserido fica logo após o "head" (head.next).
//   - O nó menos recentemente acessado (mais antigo) fica logo antes do "tail" (tail.prev).
//   - Os sentinelas eliminam a necessidade de verificações "if node == nil" nas bordas.
type LRUCache struct {
	capacity int              // Capacidade máxima de itens no cache
	items    map[string]*node // Hash map para busca O(1) de nós por chave
	head     *node            // Nó sentinela que marca o início (mais recente)
	tail     *node            // Nó sentinela que marca o fim (mais antigo)
}

// NewLRUCache é a função construtora do Cache LRU.
// Recebe a capacidade do cache (um número inteiro).
func NewLRUCache(capacity int) *LRUCache {
	cache := &LRUCache{
		capacity: capacity,
		items:    make(map[string]*node),
		head:     &node{}, // Sentinela de início
		tail:     &node{}, // Sentinela de fim
	}

	// Conecta os dois sentinelas inicialmente:
	// head <-> tail
	cache.head.next = cache.tail
	cache.tail.prev = cache.head

	return cache
}

// Get busca um valor no cache pela chave.
//
// Requisitos atendidos:
//  1. Recebe uma chave string e retorna qualquer tipo (any).
//  2. Retorna -1 se a chave não existir no cache.
//  3. Memoriza a ordem de acesso: ao acessar um item existente, move-o para a cabeça
//     da lista (marcando-o como o item mais recentemente usado).
func (c *LRUCache) Get(key string) any {
	// 1. Busca no mapa
	n, exists := c.items[key]
	if !exists {
		// Se não existir, retorna -1 conforme especificação
		return -1
	}

	// 2. Como o item acabou de ser lido, atualizamos sua ordem de acesso.
	// Removemos de sua posição atual e colocamos logo após a cabeça (head.next).
	c.moveToHead(n)

	// 3. Retorna o valor armazenado
	return n.val
}

// Set insere ou atualiza um par chave-valor no cache.
//
// Requisitos atendidos:
// 1. Recebe a chave string e o valor de qualquer tipo (any).
// 2. Se a chave já existir: atualiza o valor e move o item para a cabeça (mais recente).
// 3. Se a chave não existir:
//   - Se o número de itens atingiu a capacidade, descarta o item mais antigo
//     (o nó que fica antes do tail, ou seja, o menos recentemente usado - LRU).
//   - Insere o novo item na cabeça da lista e registra no mapa.
func (c *LRUCache) Set(key string, value any) {
	// Se a capacidade for menor ou igual a zero, o cache não comporta itens.
	if c.capacity <= 0 {
		return
	}

	// Caso 1: A chave já existe no cache
	if n, exists := c.items[key]; exists {
		// Atualiza o valor existente
		n.val = value
		// Move o nó para a posição de mais recente (início)
		c.moveToHead(n)
		return
	}

	// Caso 2: A chave é nova e o cache atingiu a capacidade máxima
	if len(c.items) >= c.capacity {
		// Remove o item menos recentemente usado (LRU), localizado em tail.prev
		lru := c.removeTail()
		// Deleta o registro do mapa para liberar a chave
		delete(c.items, lru.key)
	}

	// Cria o novo nó
	newNode := &node{
		key: key,
		val: value,
	}

	// Adiciona o novo nó no início da lista (logo após head)
	c.addToHead(newNode)

	// Salva a referência no mapa de busca
	c.items[key] = newNode
}

// ============================================================================
// Métodos auxiliares para manipulação da lista duplamente encadeada (O(1))
// ============================================================================

// addToHead insere um nó imediatamente após o nó sentinela 'head'.
func (c *LRUCache) addToHead(n *node) {
	n.prev = c.head
	n.next = c.head.next

	c.head.next.prev = n
	c.head.next = n
}

// removeNode desconecta um nó qualquer da lista.
// Graças aos ponteiros prev e next, isso é feito em O(1).
func (c *LRUCache) removeNode(n *node) {
	n.prev.next = n.next
	n.next.prev = n.prev
}

// moveToHead move um nó já existente na lista para a posição de mais recente (após head).
func (c *LRUCache) moveToHead(n *node) {
	c.removeNode(n)
	c.addToHead(n)
}

// removeTail remove e retorna o nó menos recentemente usado (aquele antes do tail).
func (c *LRUCache) removeTail() *node {
	lru := c.tail.prev
	c.removeNode(lru)
	return lru
}

// ============================================================================
// Métodos auxiliares para inspeção e depuração
// ============================================================================

// Keys retorna as chaves do cache em ordem de uso (do mais recente ao mais antigo).
func (c *LRUCache) Keys() []string {
	keys := make([]string, 0, len(c.items))
	current := c.head.next
	for current != c.tail {
		keys = append(keys, current.key)
		current = current.next
	}
	return keys
}

// String formata o estado do cache para visualização amigável.
func (c *LRUCache) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Cache (Tamanho: %d/%d) [Mais Recente -> Mais Antigo]: ", len(c.items), c.capacity))

	current := c.head.next
	first := true
	for current != c.tail {
		if !first {
			sb.WriteString(" <-> ")
		}
		sb.WriteString(fmt.Sprintf("%s:%v", current.key, current.val))
		first = false
		current = current.next
	}

	if len(c.items) == 0 {
		sb.WriteString("(vazio)")
	}

	return sb.String()
}
