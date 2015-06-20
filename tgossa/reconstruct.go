package tgossa

import "golang.org/x/tools/go/ssa"

const (
	unset = iota
	IsElse
	NotElse
	ContinueWhile
)

type BlockFormat struct {
	Seq, Index         int
	IsElseStack        []int
	EndBracketCount    int
	IsWhileCandidate   bool
	WhileCandidateEnd  int
	IsWhleCandidateEnd bool
	//IsWhile           bool
	//EndWhileIndex     int
	ReversePolarity bool
}

func rPrintf(s string, args ...interface{}) {
	//fmt.Printf(s, args...)
}

// Reconstruct builds instructions for reconstructing SSA form into a High Level Language
func Reconstruct(blocksIn []*ssa.BasicBlock, usesGr bool) []BlockFormat {
	if usesGr || len(blocksIn) == 0 /*|| len(blocksIn) > 16*/ { // NOTE only very small functions work at present
		return nil // cannot reconstruct
	}
	if blocksIn[0].Index != 0 {
		//panic("tgossa.Reconstruct 0th block does not have Index==0")
		return nil
	}
	rPrintf("\nDEBUG reconstructing %s\n", blocksIn[0].Parent().String())
	r := make([]BlockFormat, len(blocksIn)+1)
	for b := range r {
		r[b].IsElseStack = make([]int, 0, 1)
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
				if targetSeq > 0 {
					if containsIf(blocks[targetSeq-1]) > 0 {
						rPrintf("DEBUG while-candidate %s ID %d seq %d has jump target %d with prior block containing an if\n",
							blocks[0].Parent().Name(), bb.Index, b, targetSeq)
						return nil
					}
				}
				formats[targetSeq].IsWhileCandidate = true
				formats[targetSeq].WhileCandidateEnd = lastDomineeSeq(blocks, b, mapIdxToSeq) // if more than 1 jump-back, the latest/largest will be used
			}
		}
	}
	whileStack := make([]*BlockFormat, 0, 6)
	whileStack = append(whileStack, &BlockFormat{Index: 0, Seq: 0, IsWhileCandidate: true, WhileCandidateEnd: len(blocks)})
	for b, bb := range blocks {
		if formats[b].WhileCandidateEnd > 0 {
			formats[formats[b].WhileCandidateEnd].IsWhleCandidateEnd = true
		}
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
		switch len(blocks[lower].Dominees()) {
		case 0:
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
				} else {
					if targetSeq == lower+1 {
						if len(formats[targetSeq].IsElseStack) == 0 {
							rPrintf("DEBUG jump forward by 1 to non-else block\n")
							return nil // not an else point = help!
						}
					} else {
						rPrintf("DEBUG the next block does not follow this\n")
						return nil // the next block does not follow this
					}
				}
			}
			target = containsIf(blocks[lower])
			if target >= 0 {
				rPrintf("DEBUG if statements without dominees are not supported\n")
				return nil // if statements without dominees are not supported
			}
			switch len(blocks[lower].Succs) {
			case 0, 1: //NoOp
			default:
				rPrintf("DEBUG more than one successor in non-dominating block\n")
				return nil
			}
		case 1:
			if blocks[lower].Dominees()[0].Index == blocks[lower+1].Index {
				// the next block flows from this, and this alone
				if len(blocks[lower].Succs) == 2 && containsIf(blocks[lower]) > 0 {
					targetNextEndSeq := lastDomineeSeq(blocks, lower+1, mapIdxToSeq)
					targetElseStartSeq := targetNextEndSeq + 1
					if targetElseStartSeq < lower {
						rPrintf("DEBUG (1 dom) targetElseStartSeq %d < lower %d\n", targetElseStartSeq, lower)
						return nil
					}
					if len(blocks[targetNextEndSeq].Succs) == 1 &&
						(mapIdxToSeq[blocks[targetNextEndSeq].Succs[0].Index] == targetElseStartSeq) ||
						dominatorHasSuccessor(blocks, targetNextEndSeq, targetElseStartSeq, mapIdxToSeq) {
						formats[targetElseStartSeq].IsElseStack = append(formats[targetElseStartSeq].IsElseStack, NotElse)
						rPrintf("DEBUG Found (1-dom) not-else %s at sequence %d blockId %d\n",
							blocks[0].Parent().Name(), targetElseStartSeq, formats[targetElseStartSeq].Index)
					} else {
						//	formats[targetElseEndSeq].EndBracketCount++
						//	formats[targetElseStartSeq].IsElseStack = append(formats[targetElseStartSeq].IsElseStack, IsElse)
						rPrintf("DEBUG TODO Found else %s at sequence %d blockId %d\n",
							blocks[0].Parent().Name(), targetElseStartSeq, formats[targetElseStartSeq].Index)
						return nil
					}

					/*

						s, ok := mapIdxToSeq[ifNextID]
						if !ok {
							return nil
						}
						if s <= lower {
							if formats[s].IsWhileCandidate {
								end := formats[s].WhileCandidateEnd
								if lower == end {
									rPrintf("DEBUG final jump back to While - follow-on block id %d jumps back from block id %d func %s\n",
										ifNextID, blocks[lower].Index, blocks[0].Parent().String())
									return nil
									//formats[end+1].IsElseStack = append(formats[end+1].IsElseStack, ContinueWhile)
								} else {
									rPrintf("DEBUG non-final jump back to While - follow-on block id %d jumps back from block id %d func %s (lower=%d end=%d)\n",
										ifNextID, blocks[lower].Index, blocks[0].Parent().String(), lower, end)
									//return nil
									formats[end].IsElseStack = append(formats[end].IsElseStack, ContinueWhile)
								}
							} else {
								rPrintf("DEBUG non-While follow-on block id %d jumps back from block id %d func %s\n",
									ifNextID, blocks[lower].Index, blocks[0].Parent().String())
								return nil
							}
						} else {
							formats[s].IsElseStack = append(formats[s].IsElseStack, NotElse)
						}

					*/
				}
			} else {
				rPrintf("DEBUG dominated block not the next one %s %d\n", blocks[0].Parent().Name(), lower)
				return nil // only dominates one block, but that's not the next one!
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
				targetElseEndSeq := lastDomineeSeq(blocks, targetElseStartSeq, mapIdxToSeq)
				if targetNextEndSeq < lower+1 ||
					targetElseStartSeq < targetNextEndSeq ||
					targetElseEndSeq < targetElseStartSeq ||
					targetNextEndSeq+1 != targetElseStartSeq {
					rPrintf("DEBUG if statement elements in wrong sequence")
					return nil
				}
				if targetElseStartSeq < lower {
					rPrintf("DEBUG targetElseStartSeq %d < lower %d\n", targetElseStartSeq, lower)
					return nil
				}
				if len(blocks[targetNextEndSeq].Succs) == 1 &&
					(mapIdxToSeq[blocks[targetNextEndSeq].Succs[0].Index] == targetElseStartSeq) ||
					dominatorHasSuccessor(blocks, targetNextEndSeq, targetElseStartSeq, mapIdxToSeq) {
					formats[targetElseStartSeq].IsElseStack = append(formats[targetElseStartSeq].IsElseStack, NotElse)
					rPrintf("DEBUG Found not-else %s at sequence %d blockId %d\n",
						blocks[0].Parent().Name(), targetElseStartSeq, blocks[targetElseStartSeq].Index)
				} else {
					formats[targetElseEndSeq].EndBracketCount++
					formats[targetElseStartSeq].IsElseStack = append(formats[targetElseStartSeq].IsElseStack, IsElse)
					rPrintf("DEBUG Found else %s at sequence %d blockId %d\n",
						blocks[0].Parent().Name(), targetElseStartSeq, blocks[targetElseStartSeq].Index)
				}
			}
		default:
			rPrintf("DEBUG more than two dominees %s %d\n", blocks[0].Parent().Name(), len(blocks[lower].Dominees()))
			return nil // more than 2 dominees!
		}
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

	for _, s := range blks[targetSeq].Succs {
		if mapIdxToSeq[s.Index] < targetSeq {
			return true // the successors are before this in the sequence
		}
	}

	for _, instr := range blks[targetSeq].Instrs {
		/* backward jumps */
		jmp, isJump := instr.(*ssa.Jump)
		if isJump {
			for b := 0; b < targetSeq && b < len(blks); b++ {
				if jmp.Block().Succs[0].Index == blks[b].Index {
					return true
				}
			}
		}
		/**/
		_, isCall := instr.(*ssa.Call)
		if isCall {
			return true
		}
	}
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
