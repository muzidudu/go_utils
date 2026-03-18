package utils

import (
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

// EncodeHtmlEntities 将中文文字动态编码为HTML实体格式
// 参数 text: 原始中文文本
// 返回值: 编码后的文本
func EncodeHtmlEntities(text string) string {
	// 动态编码算法：将中文文字转换为编码格式
	// 生成与前端JavaScript期望格式匹配的编码：类似 JHD-PC7-MQN-NIL

	if text == "" {
		return ""
	}

	// 将字符串转换为rune切片以正确处理Unicode字符
	runes := []rune(text)
	encoded := make([]string, 0, len(runes))

	for i, char := range runes {
		// 生成编码：使用Unicode码点和位置生成32进制编码
		code := generateCodeFromUnicode(int(char), i)
		encoded = append(encoded, code)
	}

	// 使用 - 作为分隔符连接所有编码
	return strings.Join(encoded, "-")
}

// generateCodeFromUnicode 从Unicode码点生成编码
// 参数 unicode: Unicode码点
// 参数 position: 字符位置
// 返回值: 生成的编码
func generateCodeFromUnicode(unicode int, position int) string {
	// 正确的算法：直接使用Unicode码点转换为32进制
	// 从反向工程得知：JHD=20013, PC7=25991, MQN=23383, NIL=24149

	// 直接使用Unicode码点
	base := unicode

	// 转换为32进制（0-9, A-V）
	hex := strconv.FormatInt(int64(base), 32)

	// 转换为大写
	hex = strings.ToUpper(hex)

	// 确保编码长度至少为3位
	if len(hex) < 3 {
		hex = strings.Repeat("0", 3-len(hex)) + hex
	}

	// 如果长度超过4位，截取前4位
	if len(hex) > 4 {
		hex = hex[:4]
	}

	return hex
}

// DecodeHtmlEntities 将编码后的文本解码为原始中文文本（可选）
// 参数 encoded: 编码后的文本
// 返回值: 解码后的原始文本
func DecodeHtmlEntities(encoded string) string {
	if encoded == "" {
		return ""
	}

	// 按分隔符分割编码
	codes := strings.Split(encoded, "-")
	decoded := make([]rune, 0, len(codes))

	for _, code := range codes {
		// 将32进制编码转换回Unicode码点
		unicode, err := strconv.ParseInt(code, 32, 64)
		if err != nil {
			// 如果解码失败，跳过该字符
			continue
		}
		decoded = append(decoded, rune(unicode))
	}

	return string(decoded)
}

// AlphaID 将数字ID和字母代码互相转换（支持密码混淆）
// 参数 in: 输入值（数字或编码字符串）
// 参数 toNum: true表示解码为数字，false表示编码为字符串
// 参数 padUp: 填充长度（未使用）
// 参数 passKey: 密码密钥（可选）
// 返回值: 转换后的结果
func AlphaID(in interface{}, toNum bool, padUp int, passKey string) string {
	index := "abcdefghijklmnopqrstuvwxyz0123456789"

	// 如果有密码，混淆字符集
	if passKey != "" {
		index = shuffleIndex(index, passKey)
	}

	base := len(index)

	if toNum {
		// 字母代码 -> 数字
		return decodeAlphaID(in.(string), index, base, padUp)
	} else {
		// 数字 -> 字母代码
		return encodeAlphaID(in.(int64), index, base, padUp)
	}
}

// shuffleIndex 使用密码混淆字符集
func shuffleIndex(index, passKey string) string {
	// 计算密码哈希
	hash256 := sha256.Sum256([]byte(passKey))
	passhash := fmt.Sprintf("%x", hash256)

	// 如果 SHA256 哈希长度不够，使用 SHA512
	if len(passhash) < len(index) {
		hash512 := sha512.Sum512([]byte(passKey))
		passhash = fmt.Sprintf("%x", hash512)
	}

	// 准备排序
	type pair struct {
		char byte
		hash byte
	}
	pairs := make([]pair, len(index))
	for i := 0; i < len(index); i++ {
		pairs[i] = pair{
			char: index[i],
			hash: passhash[i],
		}
	}

	// 按哈希值降序排序字符集
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].hash > pairs[j].hash
	})

	// 重组字符集
	result := make([]byte, len(index))
	for i, p := range pairs {
		result[i] = p.char
	}

	return string(result)
}

// encodeAlphaID 将数字编码为字母代码
func encodeAlphaID(in int64, index string, base, padUp int) string {
	out := ""

	// 如果指定了填充长度
	if padUp > 0 {
		padUp--
		if padUp > 0 {
			in += pow(base, padUp)
		}
	}

	// 转换为指定进制
	if in == 0 {
		return string(index[0])
	}

	// PHP的逻辑: for ($t = floor(log($in, $base)); $t >= 0; $t--)
	num := big.NewInt(in)
	bigBase := big.NewInt(int64(base))

	// 计算最高位数: 找到最大的 t 使得 base^t <= in
	maxDigits := int64(0)
	for {
		testPow := new(big.Int).Exp(bigBase, big.NewInt(maxDigits+1), nil)
		if testPow.Cmp(num) > 0 {
			break
		}
		maxDigits++
	}

	// 从高位到低位构建编码
	for t := maxDigits; t >= 0; t-- {
		// bcp = bcpow($base, $t)
		bcp := new(big.Int).Exp(bigBase, big.NewInt(t), nil)

		// $a = floor($in / $bcp) % $base
		quotient := new(big.Int).Div(num, bcp)
		remainder := new(big.Int).Mod(quotient, bigBase)

		out = out + string(index[remainder.Int64()])

		// $in = $in - ($a * $bcp)
		subtract := new(big.Int).Mul(remainder, bcp)
		num.Sub(num, subtract)
	}

	// PHP代码中最后反转字符串: $out = strrev($out)
	return reverseString(out)
}

// decodeAlphaID 将字母代码解码为数字
func decodeAlphaID(in, index string, base, padUp int) string {
	// 反转字符串
	reversed := reverseString(in)

	// 计算数值
	total := big.NewInt(0)
	bigBase := big.NewInt(int64(base))

	for t := 0; t < len(reversed); t++ {
		char := reversed[t]
		pos := strings.IndexByte(index, char)
		if pos == -1 {
			return "0"
		}

		// base^(len-t-1)
		power := big.NewInt(int64(len(reversed) - t - 1))
		bcpow := new(big.Int).Exp(bigBase, power, nil)

		// pos * bcpow
		value := new(big.Int).Mul(big.NewInt(int64(pos)), bcpow)

		total.Add(total, value)
	}

	// 处理填充
	if padUp > 0 {
		padUp--
		if padUp > 0 {
			subtract := pow(base, padUp)
			total.Sub(total, big.NewInt(int64(subtract)))
		}
	}

	return total.String()
}

// pow 计算乘方
func pow(base, exp int) int64 {
	result := int64(1)
	for i := 0; i < exp; i++ {
		result *= int64(base)
	}
	return result
}

// reverseString 反转字符串
func reverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// GetRemoteThumb 获取远程缩略图地址（使用配置文件）
// 参数 thumb: 缩略图地址
// 返回值: 远程缩略图地址
// func GetRemoteThumb(thumb string) string {
// 	return GetRemoteThumbWithRemoteThumb(thumb, config.AppConfig.Site.RemoteThumb)
// }

// // GetRemoteThumbWithRemoteThumb 获取远程缩略图地址（使用指定的 remoteThumb）
// // 参数 thumb: 缩略图路径
// // 参数 remoteThumb: 远程缩略图基础地址
// // 返回值: 远程缩略图地址
// func GetRemoteThumbWithRemoteThumb(thumb string, remoteThumb string) string {
// 	if thumb == "" {
// 		return ""
// 	}
// 	if strings.HasPrefix(thumb, "http") {
// 		return thumb
// 	}
// 	if remoteThumb == "" {
// 		remoteThumb = config.AppConfig.Site.RemoteThumb
// 	}
// 	return strings.TrimSuffix(remoteThumb, "/") + "/" + strings.TrimPrefix(thumb, "/")
// }

// ID2Hash 将整数ID转换为哈希字符串
// 参数 id: 整数ID
// 参数 siteIndex: 站点索引（默认0）
// 参数 length: 返回长度（0表示不限制）
// 返回值: 编码后的字符串
func ID2Hash(id int64, siteIndex int64, length int) string {
	in := id
	if length <= 0 {
		length = 0
	}

	passKey := fmt.Sprintf("key:3a3g8y%d", siteIndex)
	return AlphaID(in, false, 0, passKey)
}

// Hash2ID 将哈希字符串转换为整数ID
// 参数 val: 编码后的字符串
// 参数 siteIndex: 站点索引（默认0）
// 参数 length: 原始长度（未使用）
// 返回值: 解码后的整数ID
func Hash2ID(val string, siteIndex int64, length int) int64 {
	passKey := fmt.Sprintf("key:3a3g8y%d", siteIndex)
	result := AlphaID(val, true, 0, passKey)
	id, _ := strconv.ParseInt(result, 10, 64)
	return id
}

// GetRandomWords 获取随机关键词
// 参数 words: 关键词列表
// 参数 num: 返回数量
// 参数 encode: 是否编码
// 参数 siteID: 站点ID（用于生成种子）
// 参数 seed: 随机种子（可选，如果为0则使用siteID和时间）
// 返回值: 随机关键词列表
func GetRandomWords(words []string, num int, encode bool, siteID int64, seed int64) []string {
	if len(words) == 0 || num <= 0 {
		return []string{}
	}

	// 限制返回数量不超过列表长度
	if num > len(words) {
		num = len(words)
	}

	// 生成随机种子
	var rngSeed int64
	if seed != 0 {
		rngSeed = seed
	} else {
		rngSeed = siteID + time.Now().Unix()
	}

	// 创建随机生成器
	rng := rand.New(rand.NewSource(rngSeed))

	// 复制列表以避免修改原始数据
	shuffled := make([]string, len(words))
	copy(shuffled, words)

	// 随机打乱列表
	rng.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	// 取前num个
	result := shuffled[:num]

	// 如果需要编码
	if encode {
		encoded := make([]string, len(result))
		for i, word := range result {
			encoded[i] = EncodeHtmlEntities(word)
		}
		return encoded
	}

	return result
}

// PathToSeed 将 URL 路径转换为 int64 seed
// 使用 SHA256 哈希路径，然后取前8字节转换为 int64
func PathToSeed(path string) int64 {
	if path == "" {
		return 0
	}

	// 使用 SHA256 哈希路径
	hash := sha256.Sum256([]byte(path))

	// 取前8字节转换为 int64
	var seed int64
	for i := 0; i < 8 && i < len(hash); i++ {
		seed = seed<<8 | int64(hash[i])
	}

	// 确保 seed 为正数
	if seed < 0 {
		seed = -seed
	}

	return seed
}

// MatchServerList 匹配 server_list（使用配置文件）
// 参数 matchStr: 匹配字符串，如果基础URL或标识中包含此字符串，则匹配对应的服务器
// 返回值: 匹配到的服务器字符串，格式为 "标识,路径@完整URL"，如果未匹配返回空字符串
func MatchServerList(matchStr string) (string, string, string) {
	return MatchServerListWithSettings(matchStr, nil)
}

// MatchServerListWithSettings 匹配 server_list（使用站点配置）
// 参数 matchStr: 匹配字符串，如果基础URL或标识中包含此字符串，则匹配对应的服务器
// 参数 settings: 站点配置（*model.SiteSettings），如果为 nil 则使用配置文件中的默认值
// 返回值: 匹配到的服务器字符串，格式为 "标识,路径@完整URL"，如果未匹配返回空字符串
func MatchServerListWithSettings(matchStr string, settings interface{}) (string, string, string) {
	var serverList map[string]string
	var playURL string

	// 如果提供了 settings，尝试从中获取配置
	if settings != nil {
		// 使用反射获取字段值
		v := reflect.ValueOf(settings)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		if v.Kind() == reflect.Struct {
			if f := v.FieldByName("ServerList"); f.IsValid() && f.CanInterface() {
				if sl, ok := f.Interface().(map[string]string); ok {
					serverList = sl
				}
			}
			if f := v.FieldByName("PlayURL"); f.IsValid() && f.CanInterface() {
				if pu, ok := f.Interface().(string); ok {
					playURL = pu
				}
			}
		}
	}

	// 如果 settings 中没有配置，使用配置文件中的默认值
	// if len(serverList) == 0 {
	// 	serverList = config.AppConfig.Site.ServerList
	// }
	// if playURL == "" {
	// 	playURL = config.AppConfig.Site.PlayUrl
	// }

	if matchStr == "" || len(serverList) == 0 {
		return matchStr, "", matchStr
	}
	// 循环 server_list
	for identifier, baseURL := range serverList {
		// 检查标识或基础URL是否包含匹配字符串
		if strings.Contains(matchStr, baseURL) {
			// 构建完整URL
			baseURL = strings.Replace(matchStr, baseURL, "", 1)
			fullURL := playURL + identifier + baseURL

			// 返回格式：完整URL
			return strings.TrimSuffix(playURL, "/") + "/" + strings.TrimPrefix(identifier, "/"), baseURL, fullURL
		}
	}
	return matchStr, "", matchStr // 未匹配返回原始字符串
}

/** 分页
 * @param page 当前页
 * @param pagesize 每页条数
 * @param total 总条数
 * @return *Paging 分页对象
 */

// 创建分页
func CreatePaging(page, pagesize, total int64) *Paging {
	if page < 1 {
		page = 1
	}
	if pagesize < 1 {
		pagesize = 10
	}

	page_count := math.Ceil(float64(total) / float64(pagesize))

	paging := new(Paging)
	paging.Page = page
	paging.Pagesize = pagesize
	paging.Total = total
	paging.PageCount = int64(page_count)
	paging.NumsCount = 7
	paging.setNums()
	return paging
}

type Paging struct {
	Page      int64   //当前页
	Pagesize  int64   //每页条数
	Total     int64   //总条数
	PageCount int64   //总页数
	Nums      []int64 //分页序数
	NumsCount int64   //总页序数
}

// 设置分页序数
func (p *Paging) setNums() {

	p.Nums = []int64{}
	if p.PageCount == 0 {
		return
	}

	half := math.Floor(float64(p.NumsCount) / float64(2))
	begin := p.Page - int64(half)
	if begin < 1 {
		begin = 1
	}

	end := begin + p.NumsCount - 1
	if end >= p.PageCount {
		begin = p.PageCount - p.NumsCount + 1
		if begin < 1 {
			begin = 1
		}
		end = p.PageCount
	}

	for i := begin; i <= end; i++ {
		p.Nums = append(p.Nums, i)
	}
}
