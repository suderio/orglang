
# Filosofia de Design

Link da conversa original:
<https://gemini.google.com/share/776a9615fd64>

## Princípios Core

1. Ortogonalidade: Poucas primitivas que combinam perfeitamente.
2. Determinismo de Memória: Uso de Arenas com ciclo de vida atrelado ao
    escopo (Teardown).
3. Fluxo como Topologia: Sincronização via grafos (sintaxe DOT) em vez
    de mutexes manuais.
4. Tudo é Stream: De arquivos a relógios, tudo segue uma interface de
    canal/fluxo.

## Estrutura da Linguagem

### 1. Funções Puras (Lógica)

Não possuem efeitos colaterais. Utilizam slots posicionais `left` e
`right` para evitar a nomeação desnecessária de variáveis (Point-free
style).

### Unárias: Operam sobre `right`

### Binárias: Operam sobre `left` e `right`

### Nulária: Não possuem argumentos

### 2. Recursos

Gerenciam o ciclo de vida dos recursos através do operador `@`.

### Setup: Ocorre na abertura do script/expressão

### Step: Ocorre a cada execução do recurso

### Teardown: Limpeza garantida ao fim do escopo (fechar FDs, resetar Arenas)

### 3. Topologia de Fluxo (dot language)

Utiliza setas `->` / `-<>` para definir como os dados pulsam entre os
nós. O paralelismo é implícito e gerenciado por *Backpressure*.

### 4. Primitivas Atômicas e de Sistema

  ----------- ----------------- ------------------------------------------------------------
  Primitiva   Abstração         Papel no Runtime
  \@handle    Canal de I/O      Interface única para Arquivos e Sockets.
  \@mem       Canal de Espaço   Memória endereçável tratada como um stream seekable.
  \@signal    Canal de Evento   Gatilhos temporais (`@clock`) ou do sistema (`@metadata`).
  \@sys       Canal de Evento   Invocação direta de syscall (ex: read, write, open).
  ----------- ----------------- ------------------------------------------------------------

### 5. Decisões de Implementação

### Compilador (Frontend)

### Linguagem: Go (pela velocidade de desenvolvimento e facilidade com CLI/Grafos)

### Estratégia de Parsing: Pratt Parser (Top-Down Operator Precedence)

### Gestão de Grafos: Biblioteca `dominikbraun/graph` para validação de topologia e detecção de deadlocks

## Alvo (Backend/Runtime)

### Linguagem Alvo: C99 / Zig

### Modelo de I/O

### Linux: `io_uring` para I/O assíncrono de alto desempenho

### Windows: `IOCP` (I/O Completion Ports) via Overlapped I/O

## Memória: Mapeamento direto de Arenas para `mmap` (Unix) ou `MapViewOfFile` (Windows) para persistência zero-copy

### Próximos Passos

1. Implementar o Lexer básico em Go usando `text/scanner`.
2. Estruturar a tabela de precedência para o Pratt Parser (NUD/LED).
3. Definir o Header de Runtime em C para gestão das Arenas e Slots
    `left` / `right`.
