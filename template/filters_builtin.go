package template

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/flosch/pongo2/v6"
)

// 本包内置 filter，在 New 时加入待注册列表，Load 时写入 pongo2。
var packageBuiltinFilters = []struct {
	name string
	fn   pongo2.FilterFunction
}{
	{"contains", filterContains},
	{"trim", filterTrim},
	{"trim_left", filterTrimLeft},
	{"trim_right", filterTrimRight},
	{"list", filterList},
	{"fields", filterFields},
	{"count", filterCount},
	{"index", filterIndex},
	{"repeat", filterRepeat},
	{"dump", filterDump},
	{"split", filterSplit},
	{"json", filterJson},
	{"wordwrap", filterWordwrap},
	{"suffix", filterSuffix},
	{"default_empty", filterDefaultEmpty},
	{"truncate_chars", filterTruncateChars},
	{"replace", filterReplace},
}

func (e *SitesEngine) initBuiltinFilters() {
	for _, b := range packageBuiltinFilters {
		e.filterEntries = append(e.filterEntries, filterEntry{name: b.name, fn: b.fn})
	}
}

// filterContains：判断字符串是否包含参数。模板: {{ "hi"|contains:"h" }} -> true
func filterContains(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	return pongo2.AsValue(in.Contains(param)), nil
}

// filterTrim：去除字符串两端的空格。模板: {{ " hi "|trim }} -> "hi"
func filterTrim(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	if param.IsNil() || len(param.String()) == 0 {
		return pongo2.AsValue(strings.TrimSpace(in.String())), nil
	}
	return pongo2.AsValue(strings.Trim(in.String(), param.String())), nil
}

// filterTrimLeft：去除字符串左端的空格。模板: {{ " hi "|trim_left }} -> "hi "
func filterTrimLeft(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	return pongo2.AsValue(strings.TrimLeft(in.String(), param.String())), nil
}

// filterTrimRight：去除字符串右端的空格。模板: {{ " hi "|trim_right }} -> " hi"
func filterTrimRight(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	return pongo2.AsValue(strings.TrimRight(in.String(), param.String())), nil
}

// filterReplace：替换字符串。模板: {{ "hi"|replace:"h","j" }} -> "jij"
func filterReplace(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	sep := strings.Split(param.String(), ",")
	from := sep[0]
	to := ""
	if len(sep) > 1 {
		to = sep[1]
	}
	return pongo2.AsValue(strings.ReplaceAll(in.String(), from, to)), nil
}

// filterList：将字符串转换为列表。模板: {{ "hi,hello"|list }} -> ["hi", "hello"]
func filterList(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	income := []rune(strings.TrimSpace(strings.Trim(strings.Trim(in.String(), "'\""), "[]")))
	var result = make([]string, 0, len(income))
	if len(income) == 0 {
		return pongo2.AsValue(result), nil
	}
	start := 0
	var hasComma rune
	if income[0] == '\'' || income[0] == '"' {
		hasComma = income[0]
		start = 1
	}
	for i := 1; i < len(income); i++ {
		if hasComma > 0 && income[i] == hasComma {
			tmp := income[start:i]
			result = append(result, string(tmp))
			start = i + 1
			hasComma = 0
		} else if income[i] == ',' && hasComma == 0 {
			if start < i {
				tmp := income[start:i]
				result = append(result, string(tmp))
				start = i + 1
			} else if start == i {
				start = i + 1
			}
		} else if income[i] == ' ' && hasComma == 0 {
			if start < i {
				tmp := income[start:i]
				result = append(result, string(tmp))
				start = i + 1
			} else if start == i {
				start = i + 1
			}
		} else if i == len(income)-1 && start <= i {
			tmp := income[start:]
			result = append(result, string(tmp))
		} else if (income[i] == '\'' || income[i] == '"') && hasComma == 0 {
			hasComma = income[i]
			start = i + 1
		}
	}
	return pongo2.AsValue(result), nil
}

// filterFields：将字符串转换为字段列表。模板: {{ "hi hello"|fields }} -> ["hi", "hello"]
func filterFields(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	return pongo2.AsValue(strings.Fields(in.String())), nil
}

// filterCount：计算字符串中包含参数的次数。模板: {{ "hi hello"|count:"h" }} -> 1
func filterCount(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	if in.IsString() {
		return pongo2.AsValue(strings.Count(in.String(), param.String())), nil
	}
	total := 0
	if in.CanSlice() {
		// slice
		in.Iterate(func(idx, count int, key, value *pongo2.Value) bool {
			if value != nil {
				if value.EqualValueTo(param) {
					total++
				}
			} else if key.EqualValueTo(param) {
				total++
			}
			return true
		}, func() {})
	}
	return pongo2.AsValue(total), nil
}

// filterIndex：计算字符串中参数的索引。模板: {{ "hi hello"|index:"h" }} -> 0
func filterIndex(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	if in.IsString() {
		return pongo2.AsValue(strings.Index(in.String(), param.String())), nil
	}
	index := -1
	if in.CanSlice() {
		// slice
		in.Iterate(func(idx, count int, key, value *pongo2.Value) bool {
			if value != nil {
				if value.EqualValueTo(param) {
					index = idx
					return false
				}
			} else if key.EqualValueTo(param) {
				index = idx
				return false
			}
			return true
		}, func() {})
	}
	return pongo2.AsValue(index), nil
}

// filterRepeat：重复字符串。模板: {{ "hi"|repeat:3 }} -> "hihihi"
func filterRepeat(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	return pongo2.AsValue(strings.Repeat(in.String(), param.Integer())), nil
}

// filterDump：打印字符串。模板: {{ "hi"|dump }} -> "hi"
func filterDump(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	return pongo2.AsValue(fmt.Sprintf("%+v", in.Interface())), nil
}

// filterSplit：将字符串转换为列表。模板: {{ "hi,hello"|split:"," }} -> ["hi", "hello"]
func filterSplit(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	sep := param.String()
	if sep == "\\n" {
		sep = "\n"
	}
	chunks := strings.Split(in.String(), sep)
	return pongo2.AsValue(chunks), nil
}

// filterJson：将字符串转换为 JSON。模板: {{ "hi"|json }} -> "{"hi"}"
func filterJson(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	s := in.Interface()
	buf, _ := json.Marshal(s)
	return pongo2.AsValue(string(buf)), nil
}

// filterWordwrap：按单词换行。模板: {{ "hi hello"|wordwrap:3 }} -> "hi\nhello"
func filterWordwrap(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	words := strings.Fields(in.String())
	wordsLen := len(words)
	wrapAt := param.Integer()
	if wrapAt <= 0 {
		return in, nil
	}

	linecount := int(math.Ceil(float64(wordsLen) / float64(wrapAt)))
	lines := make([]string, 0, linecount)
	for i := 0; i < linecount; i++ {
		lines = append(lines, strings.Join(words[wrapAt*i:min(wrapAt*(i+1), wordsLen)], " "))
	}
	return pongo2.AsValue(strings.Join(lines, "\n")), nil
}

// filterSuffix：末尾追加参数，无参数时追加 "!"。模板: {{ "hi"|suffix }} 或 {{ "hi"|suffix:"~" }}
func filterSuffix(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	suffix := "!"
	if param != nil && !param.IsNil() {
		suffix = param.String()
	}
	return pongo2.AsValue(in.String() + suffix), nil
}

// filterDefaultEmpty：空字符串时用参数作默认值。模板: {{ name|default_empty:"匿名" }}
func filterDefaultEmpty(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	if strings.TrimSpace(in.String()) != "" {
		return in, nil
	}
	if param == nil || param.IsNil() {
		return pongo2.AsValue(""), nil
	}
	return param, nil
}

// filterTruncateChars：按 rune 截断并加 "..."。模板: {{ content|truncate_chars:20 }}
func filterTruncateChars(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	max := 50
	if param != nil && !param.IsNil() {
		n, err := strconv.Atoi(param.String())
		if err != nil || n < 0 {
			return nil, &pongo2.Error{Sender: "truncate_chars", OrigError: err}
		}
		max = n
	}
	s := []rune(in.String())
	if len(s) <= max {
		return in, nil
	}
	if max <= 3 {
		return pongo2.AsValue(string(s[:max])), nil
	}
	return pongo2.AsValue(string(s[:max-3]) + "..."), nil
}
