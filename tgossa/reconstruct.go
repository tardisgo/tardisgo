package tgossa

import "golang.org/x/tools/go/ssa"

func rPrintf(s string, args ...interface{}) {
	//fmt.Printf(s, args...)
}

// BlockAction keeps track of what action a block has when it is reconstructed
type BlockAction int

const (
	unset BlockAction = iota
	// IsElse is an else block to an if statement
	IsElse
	// NotElse is when the if statement does not have an else block
	NotElse
	// EndElseBracket shows where to put the end-bracket of an else
	EndElseBracket
	// EndWhile shows where to put the end-bracket of a while
	EndWhile
)

func (ba BlockAction) String() string {
	switch ba {
	default:
		return "???"
	case IsElse:
		return "else"
	case NotElse:
		return "if(){}"
	case EndElseBracket:
		return "if(){}else{...}"
	case EndWhile:
		return "while(){}"
	}
}

// BlockStackEntry holds the action and where it should be done
type BlockStackEntry struct {
	action     BlockAction
	seq, index int
}

// BlockStack keeps track of the block reconstructions for a function
type BlockStack struct {
	bse []BlockStackEntry
}

// Len returns the size of the stack
func (bs *BlockStack) Len() int {
	//rPrintf("DEBUG BlockStack.len=%d\n", len(bs.bse))
	return len(bs.bse)
}

// Push onto the BlockStack
func (bs *BlockStack) Push(action BlockAction, seq, index int) bool {
	item := BlockStackEntry{action, seq, index}
	for _, bb := range bs.bse { // TODO remove these belt-and-braces tests
		if bb.seq > seq {
			rPrintf("DEBUG BlockStack Push invalid of %#v onto %#v\n", item, bs)
			return false
		}
	}
	bs.bse = append(bs.bse, item)
	//rPrintf("DEBUG Push %#v\n", item)
	return true
}

// Pop a value from the BlockStack
func (bs *BlockStack) Pop() (action BlockAction, seq, index int, ok bool) {
	l := bs.Len()
	if l == 0 {
		return // ok==false
	}
	l--
	action = bs.bse[l].action
	seq = bs.bse[l].seq
	index = bs.bse[l].index
	bs.bse = bs.bse[:l]
	ok = true
	//rPrintf("DEBUG Pop %d %d %d %v\n", action, seq, index, ok)
	return
}

// BlockFormat holds how to reconstruct an individual block
type BlockFormat struct {
	Seq, Index        int
	Stack             *BlockStack
	IsWhileCandidate  bool
	WhileCandidateEnd int
	IfCandidate       BlockAction
	ReversePolarity   bool
}

// Reconstruct builds instructions for reconstructing SSA form into a High Level Language
func Reconstruct(blocksIn []*ssa.BasicBlock, usesGr bool) []BlockFormat {
	if usesGr || len(blocksIn) == 0 /*|| len(blocksIn) > 16*/ { // NOTE only very small functions for testing
		return nil // cannot reconstruct
	}
	if blocksIn[0].Index != 0 {
		//panic("tgossa.Reconstruct 0th block does not have Index==0")
		return nil
	}
	rPrintf("\nDEBUG reconstructing %s\n", blocksIn[0].Parent().String())
	r := make([]BlockFormat, len(blocksIn)+1)
	for b := range r {
		r[b].Stack = &BlockStack{}
	}

	mapIdxToSeq := make(map[int]int)
	for b := range blocksIn {
		mapIdxToSeq[blocksIn[b].Index] = b
	}
	return reconstruct(blocksIn, r, mapIdxToSeq)
}

func reconstruct(blocks []*ssa.BasicBlock, formats []BlockFormat, mapIdxToSeq map[int]int) []BlockFormat {
	for b, bb := range blocks {
		formats[b].Index = bb.Index
		formats[b].Seq = b
		for _, ss := range bb.Succs {
			targetSeq, ok := mapIdxToSeq[ss.Index]
			if !ok {
				return nil
			}
			if targetSeq < b { // while candidate
				formats[targetSeq].IsWhileCandidate = true
				formats[targetSeq].WhileCandidateEnd = lastDomineeSeq(blocks, b, mapIdxToSeq) // if more than 1 jump-back, the latest/largest will be used
			}
		}
	}
	whileStack := make([]*BlockFormat, 0, 6)
	whileStack = append(whileStack, &BlockFormat{Index: 0, Seq: 0, IsWhileCandidate: true, WhileCandidateEnd: len(blocks)})
	for b, bb := range blocks {
		targetSeq, found := mapIdxToSeq[containsJump(bb)]
		if found {
			for whileStack[len(whileStack)-1].WhileCandidateEnd <= b {
				rPrintf("DEBUG pop while range from %d to %d (jump from %d to %d)\n",
					whileStack[len(whileStack)-1].Seq,
					whileStack[len(whileStack)-1].WhileCandidateEnd,
					b, targetSeq)
				whileStack = whileStack[:len(whileStack)-1] // Pop stack
			}
			if targetSeq < whileStack[len(whileStack)-1].Seq || targetSeq > whileStack[len(whileStack)-1].WhileCandidateEnd {
				rPrintf("DEBUG jump at seq %d to %d outside current range from %d to %d\n",
					b, targetSeq,
					whileStack[len(whileStack)-1].Seq,
					whileStack[len(whileStack)-1].WhileCandidateEnd)
				return nil
			}
		}
		if formats[b].IsWhileCandidate {
			rPrintf("DEBUG while-candidate %s ID %d from seq %d to seq %d\n",
				blocks[0].Parent().Name(), bb.Index, b, formats[b].WhileCandidateEnd)
			for whileStack[len(whileStack)-1].WhileCandidateEnd < b {
				rPrintf("DEBUG pop while range from %d to %d (new while-candidate)\n",
					whileStack[len(whileStack)-1].Seq,
					whileStack[len(whileStack)-1].WhileCandidateEnd)
				whileStack = whileStack[:len(whileStack)-1] // Pop stack
			}
			if whileStack[len(whileStack)-1].Seq >= b ||
				whileStack[len(whileStack)-1].WhileCandidateEnd < formats[b].WhileCandidateEnd {
				rPrintf("DEBUG while-candidate out of range of previous\n")
				return nil
			}
			whileStack = append(whileStack, &formats[b])
		}
	}
	//TODO
	//if len(whileStack) == 1 {
	rPrintf("DEBUG whileStack len = %d\n", len(whileStack))
	//	return nil
	//}
	for lower := range blocks {
		rPrintf("DEBUG considering %s seq %d ID %d #dominees %d\n",
			blocks[0].Parent().Name(), lower, blocks[lower].Index, len(blocks[lower].Dominees()))
		if containsForbidden(blocks, lower, mapIdxToSeq) {
			rPrintf("DEBUG block contains forbidden fruit %s seq %d\n", blocks[0].Parent().Name(), lower)
			return nil // TODO 	remove once all working
		}
		if formats[lower].WhileCandidateEnd > 0 {
			afterAWhile := formats[lower].WhileCandidateEnd + 1
			if !(formats[afterAWhile].Stack.Push(EndWhile, lower, blocks[lower].Index)) {
				return nil // push fail has rPrintf()
			}
		}
		switch len(blocks[lower].Succs) {
		case 0, 1:
			target := containsJump(blocks[lower])
			if target >= 0 {
				targetSeq, ok := mapIdxToSeq[target]
				if !ok {
					rPrintf("DEBUG targetSeq not found\n")
					return nil
				}
				if targetSeq <= lower { // only valid if while
					if !formats[targetSeq].IsWhileCandidate {
						rPrintf("DEBUG jump back to non-while block\n")
						return nil
					}
					for target := lower; target > targetSeq; target-- {
						if formats[target].IsWhileCandidate &&
							formats[target].WhileCandidateEnd >= lower {
							rPrintf("DEBUG jump back has while before target while\n")
							return nil
						}
					}
				} else {
					if targetSeq != lower+1 {
						rPrintf("DEBUG the next block does not follow this\n")
						return nil // the next block does not follow this
					}
				}
			}
			target = containsIf(blocks[lower])
			if target >= 0 {
				rPrintf("DEBUG if statements without successors are not supported\n")
				return nil
			}
		case 2:
			if containsIf(blocks[lower]) > 0 {
				targetElse := 0
				if blocks[lower].Succs[0].Index == blocks[lower+1].Index {
					targetElse = blocks[lower].Succs[1].Index
				} else {
					formats[lower].ReversePolarity = true
					if blocks[lower].Succs[1].Index == blocks[lower+1].Index {
						targetElse = blocks[lower].Succs[0].Index
					} else {
						rPrintf("DEBUG next block not the expected one %s %d\n", blocks[0].Parent().Name(), lower)
						return nil // next block is not what we are expecting, so bale out
					}
				}
				targetNextEndSeq := lastDomineeSeq(blocks, lower+1, mapIdxToSeq)
				targetElseStartSeq, blockFound := mapIdxToSeq[targetElse]
				if !blockFound {
					rPrintf("DEBUG targetElse %d not found in %s, map=%v\n",
						targetElse, blocks[0].Parent().Name(), mapIdxToSeq)
					return nil // did not find the else block
				}
				if targetElseStartSeq < lower && len(blocks[lower].Dominees()) == 1 {
					for target := lower; target > targetElseStartSeq; target-- {
						if formats[target].IsWhileCandidate &&
							formats[target].WhileCandidateEnd >= lower {
							rPrintf("DEBUG (doms==1) targetElseStartSeq < lower - while before target while\n")
							return nil
						}
					}
					rPrintf("DEBUG (doms==1) targetElseStartSeq < lower - reverse polarity\n")
					formats[lower].ReversePolarity = !(formats[lower].ReversePolarity)
					targetElseStartSeq = lower + 1
				}
				targetElseEndSeq := lastDomineeSeq(blocks, targetElseStartSeq, mapIdxToSeq)
				if targetNextEndSeq < lower+1 {
					rPrintf("DEBUG if statement elements in wrong sequence: targetNextEndSeq < lower+1\n")
					return nil
				}
				if targetElseStartSeq < targetNextEndSeq && len(blocks[lower].Dominees()) > 1 {
					rPrintf("DEBUG (doms>1) if statement elements in wrong sequence: targetElseStartSeq < targetNextEndSeq\n")
					return nil
				}
				if targetElseEndSeq < targetElseStartSeq {
					rPrintf("DEBUG if statement elements in wrong sequence: targetElseEndSeq < targetElseStartSeq\n")
					return nil
				}
				if targetNextEndSeq+1 != targetElseStartSeq && len(blocks[lower].Dominees()) > 1 {
					rPrintf("DEBUG (doms>1) if statement elements in wrong sequence: targetNextEndSeq+1 != targetElseStartSeq\n")
					return nil
				}
				if targetElseStartSeq < lower {
					rPrintf("DEBUG targetElseStartSeq %d < lower %d\n", targetElseStartSeq, lower)
					return nil
				}
				for t := lower; t <= targetElseEndSeq; t++ {
					if formats[t].IsWhileCandidate && formats[t].WhileCandidateEnd > targetElseEndSeq {
						rPrintf("DEBUG if range overlaps while loop\n")
						return nil
					}
				}
				if len(blocks[lower].Dominees()) == 1 {
					formats[lower].IfCandidate = NotElse
					if !(formats[targetElseStartSeq].Stack.Push(NotElse, lower, blocks[lower].Index)) {
						return nil
					}
					rPrintf("DEBUG Found not-else %s at sequence %d blockId %d\n",
						blocks[0].Parent().Name(), targetElseStartSeq, formats[targetElseStartSeq].Index)
				} else {
					targetElseEndSeq++
					formats[lower].IfCandidate = EndElseBracket
					if !(formats[targetElseStartSeq].Stack.Push(EndElseBracket, lower, blocks[lower].Index)) {
						return nil
					}
					rPrintf("DEBUG EndBracket set for %s at sequence %d blockId %d\n",
						blocks[0].Parent().Name(), targetElseEndSeq, formats[targetElseEndSeq].Index)
					if !(formats[targetElseStartSeq].Stack.Push(IsElse, lower, blocks[lower].Index)) {
						return nil
					}
					rPrintf("DEBUG Found else %s at sequence %d blockId %d\n",
						blocks[0].Parent().Name(), targetElseStartSeq, blocks[targetElseStartSeq].Index)
				}
			} else {
				rPrintf("DEBUG two Succs %s but no if\n", blocks[0].Parent().Name())
				return nil
			}
		default:
			rPrintf("DEBUG more than two Succs %s %d\n", blocks[0].Parent().Name(), len(blocks[lower].Succs))
			return nil // more than 2 successors!
		}
	}
	// Now check that everthing is actually in the correct allignment
	indent := []int{}
	for f, ff := range formats {
		rPrintf("DEBUG before block at seq %d id %d stack %v\n",
			f, ff.Index, indent)
		for s := len(ff.Stack.bse) - 1; s >= 0; s-- {
			ss := ff.Stack.bse[s]
			rPrintf("DEBUG stack level %d action %s \n",
				s, ss.action)
			if len(indent) == 0 {
				rPrintf("DEBUG empty block stack!\n")
				return nil
			}
			tos := indent[len(indent)-1]
			off := tos
			if off < 0 {
				off = -off
			}
			switch ss.action {
			case NotElse, EndElseBracket:
				if -ss.seq != tos || ss.action != formats[off].IfCandidate {
					rPrintf("DEBUG (end if) non-matching current block %d stack entry %d", -ss.seq, tos)
					rPrintf(" start action %s end action %s\n", formats[off].IfCandidate, ss.action)
					return nil
				}
				indent = indent[:len(indent)-1]
			case EndWhile:
				if ss.seq != tos {
					rPrintf("DEBUG (end while) non-matching current block %d stack entry \n%d", ss.seq, tos)
					return nil
				}
				indent = indent[:len(indent)-1]
			case IsElse:
				if -ss.seq != tos ||
					(NotElse != formats[off].IfCandidate && EndElseBracket != formats[off].IfCandidate) {
					rPrintf("DEBUG (else) non-matching current block %d stack entry %d", ss.seq, tos)
					rPrintf(" start action %s end action %s\n", formats[off].IfCandidate, ss.action)
					return nil
				}
				// NOTE: No Stack Pop
			default:
				rPrintf("DEBUG unhandled action!\n")
				return nil
			}
		}
		// these are processed after the above in the code
		if ff.IsWhileCandidate {
			indent = append(indent, f)
		}
		if ff.IfCandidate > 0 {
			indent = append(indent, -f) // use -ve numbers for if blocks
		}
	}
	if len(indent) != 0 {
		rPrintf("DEBUG not all blocks have been closed, stack: %v\n", indent)
		return nil
	}
	return formats
}

func dominatorHasSuccessor(blocks []*ssa.BasicBlock, domSeq, targetSeq int, mapIdxToSeq map[int]int) bool {
	if blocks[domSeq].Idom() == nil {
		return false
	}
	sucks := blocks[domSeq].Idom().Succs
	if len(sucks) == 0 {
		return false
	}
	for _, s := range sucks {
		sSeq, ok := mapIdxToSeq[s.Index]
		if ok {
			if sSeq == targetSeq {
				return true
			}
		}
	}
	idomSeq, ok := mapIdxToSeq[blocks[domSeq].Idom().Index]
	if ok {
		if dominatorHasSuccessor(blocks, idomSeq, targetSeq, mapIdxToSeq) {
			return true
		}
	}
	return false
}

func containsJump(bl *ssa.BasicBlock) int {
	for _, instr := range bl.Instrs {
		jmp, isJump := instr.(*ssa.Jump)
		if isJump {
			return jmp.Block().Succs[0].Index
		}
	}
	return -1
}

func containsForbidden(blks []*ssa.BasicBlock, targetSeq int, mapIdxToSeq map[int]int) bool {
	/* allow everything for testing...
	for _, s := range blks[targetSeq].Succs {
		if mapIdxToSeq[s.Index] < targetSeq {
			return true // the successors are before this in the sequence
		}
	}

	for _, instr := range blks[targetSeq].Instrs {
		// backward jumps
		jmp, isJump := instr.(*ssa.Jump)
		if isJump {
			for b := 0; b < targetSeq && b < len(blks); b++ {
				if jmp.Block().Succs[0].Index == blks[b].Index {
					return true
				}
			}
		}

		// calls
		_, isCall := instr.(*ssa.Call)
		if isCall {
			return true
		}
	}
	*/
	return false
}

func containsIf(bl *ssa.BasicBlock) int {
	for _, instr := range bl.Instrs {
		_if, isIf := instr.(*ssa.If)
		if isIf {
			return _if.Block().Succs[1].Index // the end of the if expression
		}
	}
	return -1
}

func lastDomineeSeq(blks []*ssa.BasicBlock, startSeq int, mapIdxToSeq map[int]int) int {
	if startSeq >= len(blks) {
		return startSeq
	}
	doms := blks[startSeq].Dominees()
	if len(doms) == 0 {
		return startSeq
	}
	ret, ok := mapIdxToSeq[doms[len(doms)-1].Index]
	if !ok {
		return startSeq
	}
	return lastDomineeSeq(blks, ret, mapIdxToSeq)
}
