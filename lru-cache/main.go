package main

import (
	"fmt"
)

type Usuario struct {
	Nome  string
	Email string
}

func main() {
	fmt.Println("==================================================")
	fmt.Println("          DEMONSTRAÇÃO DO CACHE LRU EM GO         ")
	fmt.Println("==================================================")

	// 1. Criamos um cache com capacidade para 3 itens
	capacidade := 3
	fmt.Printf("\n[1] Criando Cache LRU com capacidade = %d\n", capacidade)
	cache := NewLRUCache(capacidade)
	fmt.Println(cache)

	// 2. Inserindo os primeiros itens
	fmt.Println("\n[2] Inserindo 'A': 100, 'B': 200, 'C': 300...")
	cache.Set("A", 100)
	fmt.Println("Após Set('A', 100):", cache)

	cache.Set("B", 200)
	fmt.Println("Após Set('B', 200):", cache)

	cache.Set("C", 300)
	fmt.Println("Após Set('C', 300):", cache)

	// 3. Acessando 'A' via Get
	// Como 'A' é acessado, ele deve ir para a frente da lista (mais recentemente usado)
	fmt.Println("\n[3] Executando Get('A')...")
	valA := cache.Get("A")
	fmt.Printf("Resultado Get('A'): %v\n", valA)
	fmt.Println("Estado do Cache:", cache)
	fmt.Println("-> Note que 'A' agora é o mais recente, e 'B' passou a ser o mais antigo!")

	// 4. Testando Get em chave inexistente (deve retornar -1)
	fmt.Println("\n[4] Executando Get('X') (chave inexistente)...")
	valX := cache.Get("X")
	fmt.Printf("Resultado Get('X'): %v (esperado -1)\n", valX)

	// 5. Inserindo novo item 'D', ultrapassando a capacidade (3)
	// Como 'B' é o item mais antigo (LRU), 'B' deve ser descartado!
	fmt.Println("\n[5] Inserindo 'D': 400 (atingirá a capacidade máxima)...")
	cache.Set("D", 400)
	fmt.Println("Após Set('D', 400):", cache)
	fmt.Println("-> Note que 'B' foi removido automaticamente pois era o item menos usado!")

	// Verificando que 'B' não existe mais:
	valB := cache.Get("B")
	fmt.Printf("Get('B') após ser descartado: %v (retornou -1)\n", valB)

	// 6. Atualizando uma chave existente
	fmt.Println("\n[6] Atualizando o valor de 'C' para 'Modificado'...")
	cache.Set("C", "Modificado")
	fmt.Println("Após Set('C', 'Modificado'):", cache)
	fmt.Println("-> O valor foi atualizado e 'C' foi para a frente da lista!")

	// 7. Testando valores de diferentes tipos (structs, strings, etc.)
	fmt.Println("\n[7] Armazenando tipos complexos (Struct)...")
	cache.Set("user_1", Usuario{Nome: "Ana Silva", Email: "ana@email.com"})
	fmt.Println("Após adicionar Usuario:", cache)
	user := cache.Get("user_1")
	fmt.Printf("Dados recuperados de user_1: %+v (Tipo: %T)\n", user, user)

	fmt.Println("\n==================================================")
	fmt.Println("     FIM DA DEMONSTRAÇÃO - CACHE FUNCIONANDO!     ")
	fmt.Println("==================================================")
}
