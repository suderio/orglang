package parser

// BindingEntry represents the parsing rules for a specific identifier or operator.
type BindingEntry struct {
	LBP      int  // Left Binding Power (how tightly it binds to the left in LED)
	RBP      int  // Right Binding Power (associativity in LED)
	PrefixBP int  // Binding Power for NUD consumption (0 if not prefix)
	IsPrefix bool // True if it can appear in prefix position (NUD) as an operator
	IsInfix  bool // True if it can appear in infix position (LED)
}

// BindingTable manages dynamic operator bindings.
type BindingTable struct {
	entries map[string]BindingEntry
	parent  *BindingTable // For scoped lookups (if we support scopes in the future)
}

func NewBindingTable() *BindingTable {
	bt := &BindingTable{
		entries: make(map[string]BindingEntry),
	}
	bt.initDefaults()
	return bt
}

func (bt *BindingTable) Lookup(name string) (BindingEntry, bool) {
	entry, ok := bt.entries[name]
	if ok {
		return entry, true
	}
	if bt.parent != nil {
		return bt.parent.Lookup(name)
	}
	return BindingEntry{}, false
}

func (bt *BindingTable) RegisterPrefix(name string, bp int) {
	bt.entries[name] = BindingEntry{
		LBP:      0,
		RBP:      0,
		PrefixBP: bp,
		IsPrefix: true,
		IsInfix:  false,
	}
}

func (bt *BindingTable) RegisterInfix(name string, lbp int) {
	// Default left-associative: RBP = LBP
	bt.entries[name] = BindingEntry{
		LBP:      lbp,
		RBP:      lbp,
		PrefixBP: 0,
		IsPrefix: false,
		IsInfix:  true,
	}
}

func (bt *BindingTable) RegisterInfixRightAssoc(name string, lbp int) {
	// Right-associative: RBP = LBP - 1
	bt.entries[name] = BindingEntry{
		LBP:      lbp,
		RBP:      lbp - 1,
		PrefixBP: 0,
		IsPrefix: false,
		IsInfix:  true,
	}
}

func (bt *BindingTable) RegisterValue(name string) {
	// Just a value, NUD returns Name(name)
	bt.entries[name] = BindingEntry{
		LBP:      0,
		RBP:      0,
		PrefixBP: 0,
		IsPrefix: false,
		IsInfix:  false,
	}
}

func (bt *BindingTable) RegisterDual(name string, prefixBP, infixLBP int) {
	bt.entries[name] = BindingEntry{
		LBP:      infixLBP,
		RBP:      infixLBP, // Default left-assoc for infix
		PrefixBP: prefixBP,
		IsPrefix: true,
		IsInfix:  true,
	}
}

// RegisterCustomInfix registers an operator with explicit LBP and RBP
func (bt *BindingTable) RegisterCustomInfix(name string, lbp, rbp int) {
	bt.entries[name] = BindingEntry{
		LBP:      lbp,
		RBP:      rbp,
		PrefixBP: 0,
		IsPrefix: false,
		IsInfix:  true,
	}
}

func (bt *BindingTable) initDefaults() {
	// Infix Operators
	bt.RegisterInfixRightAssoc("**", 500)
	bt.RegisterInfix("*", 300)
	bt.RegisterInfix("/", 300)
	bt.RegisterInfix("%", 300)
	bt.RegisterInfix("&", 300)
	bt.RegisterInfix("+", 200)

	// - is dual: Prefix (negation, 900), Infix (sub, 200)
	bt.RegisterDual("-", 900, 200)

	bt.RegisterInfix("|", 200)
	bt.RegisterInfix("^", 200)
	bt.RegisterInfix("<<", 200)
	bt.RegisterInfix(">>", 200)
	bt.RegisterInfix("=", 150)
	bt.RegisterInfix("<>", 150)
	bt.RegisterInfix("~=", 150)
	bt.RegisterInfix("<", 150)
	bt.RegisterInfix(">", 150)
	bt.RegisterInfix("<=", 150)
	bt.RegisterInfix(">=", 150)
	bt.RegisterInfix("&&", 140)
	bt.RegisterInfix("||", 130)

	// Special Tokens
	bt.RegisterInfixRightAssoc(":", 80)
	bt.RegisterInfixRightAssoc("@:", 80)
	bt.RegisterInfix(",", 60)
	bt.RegisterInfix("->", 50)
	bt.RegisterInfix("-<", 50)
	bt.RegisterInfix("-<>", 50)

	// Note: |> and o logic is special (parseAtom), but we register their precedence here
	bt.RegisterInfix("|>", 400)
	bt.RegisterInfix("o", 400)
	// Elvis
	// Elvis - typically lower than arithmetic but higher than assignment
	bt.RegisterInfix("?:", 125)
	// Error Coalescing
	bt.RegisterInfix("??", 125)
	// Conditional (cond ? table)
	bt.RegisterInfix("?", 100)
	// Dot
	bt.RegisterInfix(".", 800)

	// Prefix-only operators
	bt.RegisterPrefix("!", 900)
	bt.RegisterPrefix("~", 900)
	bt.RegisterPrefix("++", 900)
	bt.RegisterPrefix("--", 900)
	bt.RegisterPrefix("@", 900)

	// @ is both prefix (resource instantiation) and infix (module loading).
	// Both use BP 900 as per the parser plan.
	bt.RegisterDual("@", 900, 900)
}
