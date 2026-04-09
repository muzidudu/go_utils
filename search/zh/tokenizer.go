package zh

import (
	"strings"

	"github.com/blevesearch/bleve/v2/analysis"
	"github.com/go-ego/gse"
)

// gseTokenizer 将 gse 分词结果转为 Bleve TokenStream（含字节偏移，供高亮与向量使用）。
type gseTokenizer struct {
	seg        *gse.Segmenter
	searchMode bool
}

func (t *gseTokenizer) Tokenize(input []byte) analysis.TokenStream {
	if len(input) == 0 {
		return nil
	}
	text := string(input)

	if t.searchMode {
		segs := t.seg.ModeSegment(input, true)
		if len(segs) > 0 {
			return segmentsToStream(segs)
		}
	}

	words := t.seg.Cut(text, true)
	return cutWordsToStream(words, text, input)
}

func segmentsToStream(segs []gse.Segment) analysis.TokenStream {
	out := make(analysis.TokenStream, 0, len(segs))
	pos := 1
	for _, s := range segs {
		tok := s.Token()
		if tok == nil {
			continue
		}
		term := tok.Text()
		if term == "" {
			continue
		}
		out = append(out, &analysis.Token{
			Term:     []byte(term),
			Start:    s.Start(),
			End:      s.End(),
			Position: pos,
			Type:     analysis.Ideographic,
		})
		pos++
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// cutWordsToStream 在 ModeSegment 无结果时回退；按词在原文中顺序匹配以恢复字节位置。
func cutWordsToStream(words []string, text string, input []byte) analysis.TokenStream {
	var out analysis.TokenStream
	pos := 1
	bytePos := 0
	for _, w := range words {
		if w == "" {
			continue
		}
		idx := strings.Index(text[bytePos:], w)
		if idx < 0 {
			continue
		}
		start := bytePos + idx
		end := start + len([]byte(w))
		out = append(out, &analysis.Token{
			Term:     input[start:end],
			Start:    start,
			End:      end,
			Position: pos,
			Type:     analysis.Ideographic,
		})
		pos++
		bytePos = end
	}
	return out
}
