package keybinds

import (
	"strconv"
	"strings"

	"github.com/samuelstranges/chronos/types"
)

// FindExact looks for exact key sequence match in the current mode and returns VimActionMsg
func (r *CommandRegistry) FindExact(keySeq string, mode types.VimMode, count int) types.VimActionMsg {
	// Select the right binding map based on mode
	var bindings map[string]string
	switch mode {
	case types.ModeNormal:
		bindings = r.normalBindings
	case types.ModeVisual:
		bindings = r.visualBindings
	case types.ModeSearch:
		bindings = r.searchBindings
	default:
		return types.VimActionMsg{}
	}

	// Check for pattern matches first (commands with # placeholders)
	if patternMsg := r.matchCommandPattern(keySeq, count); patternMsg.Action != "" {
		return patternMsg
	}

	// Try without numeric prefix
	nonDigitStart := 0
	for i, c := range keySeq {
		if c < '0' || c > '9' {
			nonDigitStart = i
			break
		}
	}

	if nonDigitStart < len(keySeq) {
		cmdPart := keySeq[nonDigitStart:]
		if action, ok := bindings[cmdPart]; ok {
			return types.VimActionMsg{Action: action, Argument: count}
		}
	}

	// Try full sequence
	if action, ok := bindings[keySeq]; ok {
		return types.VimActionMsg{Action: action, Argument: count}
	}

	return types.VimActionMsg{}
}

// matchCommandPattern checks if keySeq matches any registered pattern and returns VimActionMsg
func (r *CommandRegistry) matchCommandPattern(keySeq string, count int) types.VimActionMsg {
	// Check each registered pattern entry (already in deterministic order)
	for _, entry := range r.patternEntries {
		if match, args := r.matchPattern(keySeq, entry.Pattern); match {
			// Extract the numeric argument from the pattern and use it as count
			patternCount := count
			if len(args) > 0 {
				if num, err := strconv.Atoi(args[0]); err == nil {
					patternCount = num
				}
			}
			return types.VimActionMsg{Action: entry.Action, Argument: patternCount}
		}
	}

	return types.VimActionMsg{}
}

// matchPattern checks if keySeq matches the pattern and extracts numeric arguments
func (r *CommandRegistry) matchPattern(keySeq, pattern string) (bool, []string) {
	if len(keySeq) != len(pattern) {
		return false, nil
	}

	var args []string
	var currentArg strings.Builder

	for i, pc := range pattern {
		sc := rune(keySeq[i])

		if pc == '#' {
			// Pattern expects a digit
			if sc >= '0' && sc <= '9' {
				currentArg.WriteRune(sc)
			} else {
				return false, nil
			}
		} else {
			// Pattern expects exact character match
			if pc != sc {
				return false, nil
			}
			// If we were building an argument, save it
			if currentArg.Len() > 0 {
				args = append(args, currentArg.String())
				currentArg.Reset()
			}
		}
	}

	// Save any remaining argument
	if currentArg.Len() > 0 {
		args = append(args, currentArg.String())
	}

	return true, args
}

// IsPrefix checks if sequence is prefix of longer command in the current mode
func (r *CommandRegistry) IsPrefix(keySeq string, mode types.VimMode) bool {
	// Check for pattern prefixes first
	if r.isPatternPrefix(keySeq) {
		return true
	}

	if prefixes, ok := r.prefixBindings[mode]; ok {
		return prefixes[keySeq]
	}
	return false
}

// isPatternPrefix checks if a sequence could be the start of a valid pattern command
func (r *CommandRegistry) isPatternPrefix(seq string) bool {
	if len(seq) == 0 {
		return false
	}

	// Check if sequence could be a prefix of any registered pattern
	for _, entry := range r.patternEntries {
		if r.couldMatchPattern(seq, entry.Pattern) {
			return true
		}
	}

	return false
}

// couldMatchPattern checks if sequence could be a partial match for pattern
func (r *CommandRegistry) couldMatchPattern(seq, pattern string) bool {
	if len(seq) >= len(pattern) {
		return false // Sequence too long to be a prefix
	}

	// Check if sequence matches pattern up to its length
	for i, sc := range seq {
		if i >= len(pattern) {
			return false
		}

		pc := rune(pattern[i])
		if pc == '#' {
			// Pattern expects a digit
			if sc < '0' || sc > '9' {
				return false
			}
		} else {
			// Pattern expects exact character match
			if pc != sc {
				return false
			}
		}
	}

	return true
}
