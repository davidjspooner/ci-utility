package textfmt

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type sgrState struct {
	attrs map[int]bool // active SGR codes
}

func newSGRState() *sgrState {
	return &sgrState{attrs: make(map[int]bool)}
}

func (s *sgrState) apply(code int) {
	if code == 0 {
		s.attrs = make(map[int]bool)
		return
	}
	s.attrs[code] = true
}

func (s *sgrState) reset() {
	s.attrs = make(map[int]bool)
}

func (s *sgrState) copy() *sgrState {
	newAttrs := make(map[int]bool)
	for k, v := range s.attrs {
		newAttrs[k] = v
	}
	return &sgrState{attrs: newAttrs}
}

func (s *sgrState) string() string {
	if len(s.attrs) == 0 {
		return ""
	}
	var codes []int
	for code := range s.attrs {
		codes = append(codes, code)
	}
	sort.Ints(codes)
	parts := make([]string, len(codes))
	for i, code := range codes {
		parts[i] = strconv.Itoa(code)
	}
	return "\x1b[" + strings.Join(parts, ";") + "m"
}

func parseSGRCodes(seq string) ([]int, error) {
	if !strings.HasPrefix(seq, "\x1b[") || !strings.HasSuffix(seq, "m") {
		return nil, fmt.Errorf("invalid SGR sequence format")
	}
	body := seq[2 : len(seq)-1] // strip \x1b[ and trailing 'm'
	parts := strings.Split(body, ";")
	var codes []int
	for _, p := range parts {
		if p == "" {
			continue
		}
		val, err := strconv.Atoi(p)
		if err != nil || !isValidSGRCode(val) {
			return nil, fmt.Errorf("invalid SGR code")
		}
		codes = append(codes, val)
	}
	return codes, nil
}
