package search

import (
	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/search/query"
)

// Match 创建 Match 查询（分词匹配，默认搜索所有 text 字段）
func Match(text string) query.Query {
	return bleve.NewMatchQuery(text)
}

// MatchIn 在指定字段中 Match 查询
// 例: MatchIn("Go", "title") 仅在 title 字段搜索 "Go"
func MatchIn(text, field string) query.Query {
	q := bleve.NewMatchQuery(text)
	if field != "" {
		q.SetField(field)
	}
	return q
}

// MatchPhrase 创建短语查询（精确短语）
func MatchPhrase(phrase string) query.Query {
	return bleve.NewMatchPhraseQuery(phrase)
}

// MatchPhraseIn 在指定字段中短语查询
func MatchPhraseIn(phrase, field string) query.Query {
	q := bleve.NewMatchPhraseQuery(phrase)
	if field != "" {
		q.SetField(field)
	}
	return q
}

// QueryString 创建查询字符串（支持 field:value 语法）
// 例: "title:golang +content:search" 或 "title:Go content:语言"
func QueryString(q string) query.Query {
	return bleve.NewQueryStringQuery(q)
}

// Term 创建精确词项查询
func Term(term string) query.Query {
	return bleve.NewTermQuery(term)
}

// TermIn 在指定字段中精确词项查询
func TermIn(term, field string) query.Query {
	q := bleve.NewTermQuery(term)
	if field != "" {
		q.SetField(field)
	}
	return q
}

// Prefix 创建前缀查询
func Prefix(prefix string) query.Query {
	return bleve.NewPrefixQuery(prefix)
}

// PrefixIn 在指定字段中前缀查询
func PrefixIn(prefix, field string) query.Query {
	q := bleve.NewPrefixQuery(prefix)
	if field != "" {
		q.SetField(field)
	}
	return q
}

// Fuzzy 创建模糊查询
func Fuzzy(term string) query.Query {
	return bleve.NewFuzzyQuery(term)
}

// FuzzyIn 在指定字段中模糊查询
func FuzzyIn(term, field string) query.Query {
	q := bleve.NewFuzzyQuery(term)
	if field != "" {
		q.SetField(field)
	}
	return q
}

// Conjunction 创建 AND 组合查询
func Conjunction(queries ...query.Query) query.Query {
	return bleve.NewConjunctionQuery(queries...)
}

// Disjunction 创建 OR 组合查询
func Disjunction(queries ...query.Query) query.Query {
	return bleve.NewDisjunctionQuery(queries...)
}

// Bool 创建布尔查询
// must: 必须匹配, should: 至少匹配一个, mustNot: 必须不匹配
func Bool(must, should, mustNot []query.Query) query.Query {
	q := bleve.NewBooleanQuery()
	for _, m := range must {
		q.AddMust(m)
	}
	for _, s := range should {
		q.AddShould(s)
	}
	for _, n := range mustNot {
		q.AddMustNot(n)
	}
	return q
}
