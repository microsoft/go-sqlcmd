// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package console

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type completer func(line string) []string

func TestKeywordComplete(t *testing.T) {
	testKeywordComplete(t, CompleteLine)
}

func BenchmarkAutoComplete(b *testing.B) {
	for i := 0; i < b.N; i++ {
		testKeywordComplete(b, CompleteLine)
	}
}

func BenchmarkPrefixTreeAutoComplete(b *testing.B) {
	for i := 0; i < b.N; i++ {
		testKeywordComplete(b, prefixTreeCompleteLine)
	}
}

// This method is generic so we could plug in different implementations of completer for comparison
func testKeywordComplete(t testing.TB, c completer) {
	t.Helper()
	keywords := c("BR")
	assert.ElementsMatch(t, []string{"BREAK", "BROWSE"}, keywords, "CompleteLine(ER)")
	keywords = c("SELECT name fr")
	assert.ElementsMatch(t, []string{"SELECT name freetext", "SELECT name freetexttable", "SELECT name from"}, keywords, "CompleteLine(SELECT name fr)")
	keywords = c("my word is SEMANTIC")
	assert.ElementsMatch(t, []string{"my word is SEMANTICKEYPHRASETABLE", "my word is SEMANTICSIMILARITYDETAILSTABLE", "my word is SEMANTICSIMILARITYTABLE"}, keywords, "CompleteLine(SEMANTIC)")
	keywords = c("BREAD")
	assert.Empty(t, keywords, "CompleteLine(BREAD)")
	keywords = c("Z")
	assert.Empty(t, keywords, "CompleteLine(BREAD)")
}

// This code provides an alternative implementation of console.CompleteLine
// It's here for benchmarking purposes, so if one outperforms the other we can easily swap.
// Binary search outperforms it, likely due to the relatively small search space.
// With 179 keywords:
/*goos: windows
goarch: amd64
pkg: github.com/microsoft/go-sqlcmd/pkg/console
cpu: AMD Ryzen 9 5950X 16-Core Processor
BenchmarkAutoComplete
BenchmarkAutoComplete-32                  491948              2481 ns/op             586 B/op         31 allocs/op
BenchmarkPrefixTreeAutoComplete
BenchmarkPrefixTreeAutoComplete-32        383217              3124 ns/op            1450 B/op         37 allocs/op
PASS
ok      github.com/microsoft/go-sqlcmd/pkg/console      2.746s
*/
func prefixTreeCompleteLine(line string) []string {
	idx := strings.LastIndexAny(line, " ;") + 1
	// we don't try to complete without a starting letter
	if idx == len(line) {
		return []string{}
	}
	prefix := strings.ToUpper(string(line[idx:]))
	words := keywordList.GetKeywords(prefix)
	lines := make([]string, len(words))
	for i, w := range words {
		lines[i] = mergeLine(line, w, idx)
	}
	return lines
}

type prefixTree struct {
	children      [27]*prefixTree
	isLeaf        bool
	maxlen        int
	word          string
	validChildren []int
}

func newPrefixTree() *prefixTree {
	var tree = &prefixTree{
		isLeaf:        false,
		validChildren: make([]int, 0, 10),
	}
	return tree
}

func runeIndex(ch rune) rune {
	if ch == '_' {
		return 26
	} else {
		return ch - 'A'
	}
}

// Insert relies on the incoming list being sorted
func (p *prefixTree) Insert(word string) {
	cur := p
	for _, ch := range word {
		idx := runeIndex(ch)
		if cur.children[idx] == nil {
			cur.children[idx] = newPrefixTree()
			cur.validChildren = append(cur.validChildren, int(idx))
		}
		cur = cur.children[idx]
	}
	cur.isLeaf = true
	cur.word = word
	if len(word) > p.maxlen {
		p.maxlen = len(word)
	}
}

func (p *prefixTree) GetKeywords(prefix string) []string {
	cur := p
	for _, ch := range prefix {
		idx := runeIndex(ch)
		if idx < 0 || idx > 26 {
			return []string{}
		}
		if cur.children[idx] == nil {
			return []string{}
		} else {
			cur = cur.children[idx]
		}
	}
	length := len(prefix)
	words := make([]string, 0, 10)
	word := make([]rune, length, p.maxlen)
	copy(word, []rune(prefix))
	cur.appendLevel(&word, length, &words)
	return words
}

func (p *prefixTree) appendLevel(word *[]rune, length int, words *[]string) {
	if p.isLeaf {
		*words = append(*words, p.word)
	}
	for _, i := range p.validChildren {
		ch := '_'
		if i < 26 {
			ch = 'A' + rune(i)
		}
		if len(*word) == length {
			*word = append(*word, ch)
		} else {
			(*word)[length] = ch
		}
		p.children[i].appendLevel(word, length+1, words)
	}
}

var keywordList *prefixTree

// use capital letters for all keywords
func init() {
	keywordList = newPrefixTree()
	for i := range keywords {
		keywordList.Insert(keywords[i])
	}
}
