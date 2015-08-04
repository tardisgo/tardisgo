package haxe

import (
	"github.com/tardisgo/tardisgo/pogo"
	"github.com/tardisgo/tardisgo/tgossa"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/types"
	"golang.org/x/tools/go/types/typeutil"
)

type haxeContext struct {
	pogoComp *pogo.Compilation // the host compilation context

	langListEntry    int  // which entry is this language in pogo.LanguageList
	useRegisterArray bool // should we use an array rather than individual register vars

	nextReturnAddress       int           // what number is the next pseudo block return address?
	hadReturn               bool          // has there been a return statement in this function?
	hadBlockReturn          bool          // has there been a return in this block?
	pseudoNextReturnAddress int           // what is the next pseudo block to emit/or limit of what's been emitted
	pseudoBlockNext         int           // what is the next pseudo block we should have emitted?
	currentfn               *ssa.Function // what we are currently working on
	currentfnName           string        // the Haxe name of what we are currently working on
	fnUsesGr                bool          // does the current function use Goroutines?
	fnTracksPhi             bool          // does the current function track Phi?

	funcNamesUsed     map[string]bool
	fnCanOptMap       map[string]bool
	reconstructInstrs []tgossa.BlockFormat
	elseStack         []string

	map1usePtr map[ssa.Value]oneUsePtr

	localFunctionMap map[int]string
	thisBlock        int

	rangeChecks map[string]struct{}

	inMustSplitSubFn bool
	deDupRHS         map[string]string
	subFnInstrs      []ssa.Instruction

	tempVarList []regToFree

	typesByID []types.Type
	pte       typeutil.Map
	pteKeys   []types.Type
}

type langType struct {
	hc *haxeContext
}

func (l langType) InitLang(i int, comp *pogo.Compilation) pogo.Language {
	ret := langType{hc: &haxeContext{
		langListEntry: i,
		pogoComp:      comp,
	}}
	ret.hc.funcNamesUsed = make(map[string]bool)
	return ret
}
func (l langType) PogoComp() *pogo.Compilation {
	return l.hc.pogoComp
}
