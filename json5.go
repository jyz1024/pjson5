package pjson5

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const (
	errTrimStringPartLen = 10

	colon     = ':'
	comma     = ','
	space     = ' '
	backslash = '/'

	lineBreak = "\n"
	quot      = "\""
	Root      = "$"
)

var (
	arrayPair   = [2]byte{'[', ']'}
	objectPair  = [2]byte{'{', '}'}
	placeholder = []byte{space, space}
)

var (
	errParseJsonErrorTmpl = "invalid JSON5 value at position %d: %s"
)

const (
	dataTypeComment int32 = 1 << iota
	dataTypeCommentLine
	dataTypeStartFlag
	dataTypeKey
	dataTypeColon
	dataTypeVal
	dataTypeComma
	dataTypeEndFlag
	dataTypeLineBreak
)

type dataBlock struct {
	Typ int32  // 数据类型
	Val string // 数据内容
}

func (db dataBlock) Is(multiTyp int32) bool {
	return db.Typ&multiTyp != 0
}

func (db dataBlock) KeyUnQuot() string {
	if db.Typ != dataTypeKey {
		return db.Val
	}
	return strings.Trim(db.Val, quot)
}

type Type int

const (
	// None not exist
	None Type = iota
	// Null json null
	Null
	// Boolean json boolean
	Boolean
	// Number is json number
	Number
	// String is a json string
	String
	// Array is a json array
	Array
	// Object is a json object
	Object
)

type Node struct {
	raw    string // 原始未解析值，用于懒解析
	parsed bool   // 是否已经解析过了

	typ      Type             // 类型
	block    []dataBlock      // 解析数据块
	val      string           // 解析后的值部分(对于非数组/对象类型为不含注释的raw，数组对象类型为开始位置到结束位置之间的值)
	children map[string]*Node // 子节点元素信息,仅Object结构

	parseIdx int   // 当前解析位置
	err      error // 解析失败信息
}

func New(json string) *Node {
	return &Node{raw: json}
}

func (n *Node) Type() Type {
	return n.parse().typ
}

func (n *Node) Value() string {
	if n.parsed {
		return n.val
	}
	return n.raw
}

func (n *Node) Error() error {
	return n.err
}

func (n *Node) exceptLineBreak(pos int) bool {
	if pos >= len(n.raw) {
		return false
	}
	return n.raw[pos] == lineBreak[0]
}

func (n *Node) except(c byte) bool {
	if n.parseIdx >= len(n.raw) {
		return false
	}
	return n.raw[n.parseIdx] == c
}

func (n *Node) Parse() *Node {
	return n.parse()
}

func (n *Node) parse() *Node {
	if n.parsed {
		return n
	}
	n.parsed = true
parse:
	if n.err != nil {
		return n
	}
	var skipLB, containsLB bool
	n.parseIdx, skipLB = skipWhiteSpace(n.raw, n.parseIdx) // 跳过所有的空白字符
	startIdx := n.parseIdx
	if n.parseIdx >= len(n.raw) {
		n.parseErr(n.parseIdx)
		return n
	}
	switch n.raw[n.parseIdx] {
	case backslash:
		containsLB, _ = n.parseComment(true, skipLB || containsLB)
		goto parse
	case '{':
		n.typ = Object
		n.children = make(map[string]*Node)
		n.parseObject()
	case '[':
		n.typ = Array
		n.parseCombineEnd(arrayPair)
	case '"', '\'':
		n.typ = String
		n.parseString()
	case 't', 'f':
		n.typ = Boolean
		n.parseBoolean()
	case 'n':
		n.typ = Null
		n.parseNull()
	case '-', '+', '.', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'I', 'N':
		n.typ = Number
		n.parseNumber()
	default:
		n.parseErr(n.parseIdx)
	}
	if n.err != nil {
		return n
	}
	if n.typ != Object && startIdx < n.parseIdx {
		n.block = append(n.block, dataBlock{Typ: dataTypeVal})
		n.val = n.raw[startIdx:n.parseIdx]
	}
	// 末尾逗号
	n.parseIdx = skipLineWhiteSpace(n.raw, n.parseIdx)
	if n.except(comma) {
		n.block = append(n.block, dataBlock{Typ: dataTypeComma})
		n.parseIdx++
	}
	// 末尾换行
	n.parseIdx, skipLB = skipWhiteSpace(n.raw, n.parseIdx)
	if skipLB {
		n.block = append(n.block, dataBlock{Typ: dataTypeLineBreak})
	}
	containsLB = false
	// 处理注释
	for n.err == nil && n.parseIdx < len(n.raw) {
		n.parseIdx, skipLB = skipWhiteSpace(n.raw, n.parseIdx)
		if !n.except(backslash) {
			n.parseErr(n.parseIdx)
			break
		}
		containsLB, _ = n.parseComment(true, skipLB || containsLB)
	}
	return n
}

func (n *Node) parseErr(parseIdx int) {
	n.err = fmt.Errorf(errParseJsonErrorTmpl, parseIdx, trimStringPart(n.raw, parseIdx, errTrimStringPartLen))
}

// parseComment 解析注释，返回解析后的位置
func (n *Node) parseComment(wBlock bool, isNotInLine bool) (endWithLB bool, suc bool) {
	pos := n.parseIdx
	if pos+1 >= len(n.raw) {
		n.err = fmt.Errorf(errParseJsonErrorTmpl, pos+1, trimStringPart(n.raw, pos, errTrimStringPartLen))
		return
	}
	var endIdx int
	switch n.raw[pos+1] {
	case backslash:
		endIdx = strings.Index(n.raw[pos+2:], lineBreak)
		if endIdx == -1 {
			n.parseIdx = len(n.raw)
		} else {
			n.parseIdx = pos + 2 + endIdx + 1 // 包括换行符
			endWithLB = true
		}
	case '*':
		endIdx = strings.Index(n.raw[pos+2:], "*/")
		if endIdx == -1 {
			n.err = fmt.Errorf(errParseJsonErrorTmpl, pos+1, trimStringPart(n.raw, pos, errTrimStringPartLen))
			return
		}
		skipWhitePos := skipLineWhiteSpace(n.raw, endIdx)
		if skipWhitePos < len(n.raw) && n.exceptLineBreak(skipWhitePos) {
			n.parseIdx = skipWhitePos + 1
			endWithLB = true
		} else {
			n.parseIdx = pos + 2 + endIdx + 2
		}
	default:
		n.parseIdx++
		return endWithLB, false
	}
	if !wBlock {
		return endWithLB, true
	}
	typ := dataTypeCommentLine
	if isNotInLine {
		typ = dataTypeComment
	}
	n.block = append(n.block, dataBlock{
		Typ: typ,
		Val: n.raw[pos:n.parseIdx],
	})
	return endWithLB, true
}

func (n *Node) parseObject() {
	objStartIdx := n.parseIdx
	n.parseIdx++
	n.block = append(n.block, dataBlock{Typ: dataTypeStartFlag})

	var containsLB, skipLB bool
	pos := skipLineWhiteSpace(n.raw, n.parseIdx)
	if n.exceptLineBreak(pos) {
		n.block = append(n.block, dataBlock{Typ: dataTypeLineBreak})
	}
	keyBlock := dataBlock{Typ: dataTypeKey}
	for n.parseIdx < len(n.raw) && n.err == nil {
		n.parseIdx, skipLB = skipWhiteSpace(n.raw, n.parseIdx)
		if n.raw[n.parseIdx] != backslash {
			containsLB = false
		}
		switch n.raw[n.parseIdx] {
		case '}':
			n.parseIdx++
			n.block = append(n.block, dataBlock{Typ: dataTypeEndFlag})
			n.val = n.raw[objStartIdx:n.parseIdx]
			return
		case backslash:
			containsLB, _ = n.parseComment(true, containsLB || skipLB)
			continue
		case colon:
			n.parseIdx++
			n.block = append(n.block, dataBlock{Typ: dataTypeColon})
			continue
		case comma:
			n.parseIdx++
			n.block = append(n.block, dataBlock{Typ: dataTypeComma})
			continue
		}
		startIdx := n.parseIdx
		var block dataBlock
		// 判断当前是否解析了key
		if keyBlock.Val == "" { // 尝试获取到key
			n.parseObjectKey()
			block = dataBlock{Typ: dataTypeKey}
		} else {
			n.parseObjectVal()
			block = dataBlock{Typ: dataTypeVal}
		}
		if n.err != nil {
			return
		}
		switch block.Typ {
		case dataTypeKey:
			keyBlock.Val = n.raw[startIdx:n.parseIdx]
			block.Val = keyBlock.Val
			if _, ok := n.children[block.KeyUnQuot()]; ok {
				n.err = errors.New("repeat key:" + block.KeyUnQuot())
				return // 重复的key
			}
		case dataTypeVal:
			n.children[keyBlock.KeyUnQuot()] = &Node{raw: n.raw[startIdx:n.parseIdx]}
			keyBlock.Val = ""
		}
		n.block = append(n.block, block)
		if block.Typ == dataTypeVal { // 是否直接换行
			n.parseIdx = skipLineWhiteSpace(n.raw, n.parseIdx)
			if n.except(comma) {
				n.block = append(n.block, dataBlock{Typ: dataTypeComma})
				n.parseIdx++
			}
			n.parseIdx, skipLB = skipWhiteSpace(n.raw, n.parseIdx)
			if skipLB {
				n.block = append(n.block, dataBlock{Typ: dataTypeLineBreak})
			}
		}
	}
}

func (n *Node) parseObjectKey() {
	// key中间不允许插入注释
	var endFn func(ch byte) bool
	skipCharNum := 0
	if n.raw[n.parseIdx] == '"' {
		// 找到结束引号的位置
		endFn = func(ch byte) bool {
			return ch == '"'
		}
		skipCharNum = 1
	} else {
		// 找到空白字符或者:的位置
		endFn = func(ch byte) bool {
			return isWhitespaceNLB(ch) || ch == colon
		}
	}
	for i := n.parseIdx + 1; i < len(n.raw); i++ {
		if endFn(n.raw[i]) {
			n.parseIdx = i + skipCharNum
			return
		}
	}
	n.parseErr(len(n.raw))
}

func (n *Node) parseObjectVal() {
	switch n.raw[n.parseIdx] {
	case '{':
		n.parseCombineEnd(objectPair)
	case '[':
		n.parseCombineEnd(arrayPair)
	case '"', '\'':
		n.parseString()
	case 't', 'f':
		n.parseBoolean()
	case 'n':
		n.parseNull()
	case '-', '+', '.', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'I', 'N':
		n.parseNumber()
	default:
		n.parseErr(n.parseIdx)
	}
}

func (n *Node) parseCombineEnd(pair [2]byte) {
	// 寻找对应的结束位置
	leftFlagNum := 1
	n.parseIdx++
	for n.parseIdx < len(n.raw) && leftFlagNum > 0 && n.err == nil {
		switch n.raw[n.parseIdx] {
		case backslash:
			n.parseComment(false, false)
			continue
		case pair[0]:
			leftFlagNum++
		case pair[1]:
			leftFlagNum--
		}
		n.parseIdx++
	}
	if n.err != nil {
		return
	}
	if leftFlagNum > 0 {
		n.parseErr(n.parseIdx)
		return
	}
}

func (n *Node) parseString() {
	rawStr := n.raw
	// expects that the lead character is a '"'
	for i := n.parseIdx + 1; i < len(rawStr); i++ {
		if rawStr[i] > '\\' {
			continue
		}
		if rawStr[i] == '"' {
			n.parseIdx = i + 1
			return
		}
		if rawStr[i] == '\\' {
			i++
			for ; i < len(rawStr); i++ {
				if rawStr[i] > '\\' {
					continue
				}
				if rawStr[i] == '"' {
					// look for an escaped slash
					if rawStr[i-1] == '\\' {
						n := 0
						for j := i - 2; j > 0; j-- {
							if rawStr[j] != '\\' {
								break
							}
							n++
						}
						if n%2 == 0 {
							continue
						}
					}
					n.parseIdx = i + 1
					return
				}
			}
			if i+1 < len(rawStr) {
				n.parseIdx = i + 1
			} else {
				n.parseIdx = i
			}
			return
		}
	}
	n.parseErr(len(rawStr) - 1)
}

func (n *Node) parseBoolean() {
	var checkVal string
	switch n.raw[n.parseIdx] {
	case 't':
		checkVal = "true"
	case 'f':
		checkVal = "false"
	default:
		n.parseErr(n.parseIdx)
		return
	}
	if !strings.HasPrefix(n.raw[n.parseIdx:], checkVal) {
		n.parseErr(n.parseIdx)
		return
	}
	n.parseIdx += len(checkVal)
}

func (n *Node) parseNull() {
	if n.parseIdx+4 <= len(n.raw) && strings.EqualFold(n.raw[n.parseIdx:n.parseIdx+4], "null") {
		n.parseIdx += len(n.val)
		return
	}
	n.parseErr(n.parseIdx)
}

func (n *Node) parseNumber() {
	// 通过空白字符或者非有效字符找到结束位置
	endIdx := n.parseIdx + findEndOfNumber(n.raw[n.parseIdx:])
	_, err := strconv.ParseFloat(n.raw[n.parseIdx:endIdx], 64)
	if err != nil {
		n.parseErr(n.parseIdx)
		return
	}
	n.parseIdx = endIdx
}

func (n *Node) Pretty() string {
	if n.err != nil {
		return n.err.Error()
	}
	buf := &strings.Builder{}
	buf.Grow(len(n.raw))
	// 重新组装Node结构返回
	buildNodeData(buf, n, 0)
	return buf.String()
}

func buildNodeData(buf *strings.Builder, node *Node, level int) {
	if !node.parsed {
		buf.WriteString(node.raw)
		return
	}
	preKey := ""
	for idx, block := range node.block {
		switch block.Typ {
		case dataTypeComment:
			buf.Write(bytes.Repeat(placeholder, level))
			fallthrough
		case dataTypeCommentLine:
			buf.WriteString(block.Val)
		case dataTypeStartFlag:
			switch node.typ {
			case Object:
				buf.WriteByte(objectPair[0])
			case Array:
				buf.WriteByte(arrayPair[0])
			}
			if !nextBlockIs(node, idx, dataTypeLineBreak) {
				buf.WriteByte(space)
			}
			level++
		case dataTypeKey:
			buf.Write(bytes.Repeat(placeholder, level))
			buf.WriteString(block.Val)
			preKey = strings.Trim(block.Val, quot)
		case dataTypeColon:
			buf.WriteByte(colon)
			buf.WriteByte(space)
		case dataTypeVal:
			if node.typ != Object {
				buf.WriteString(node.val)
				continue
			}
			buildNodeData(buf, node.children[preKey], level)
		case dataTypeComma:
			buf.WriteByte(comma)
			if nextBlockIs(node, idx, dataTypeKey) {
				buf.WriteString(lineBreak)
			} else {
				buf.WriteByte(space)
			}
		case dataTypeEndFlag:
			level--
			buf.Write(bytes.Repeat(placeholder, level))
			switch node.typ {
			case Object:
				buf.WriteByte(objectPair[1])
			case Array:
				buf.WriteByte(arrayPair[1])
			}
		case dataTypeLineBreak:
			buf.WriteString(lineBreak)
		}
	}
}

func nextBlockIs(node *Node, idx int, typ int32) bool {
	if idx >= len(node.block)-1 {
		return false
	}
	return node.block[idx+1].Typ == typ
}

func (n *Node) Exists(path string) bool {
	node := n.Get(path)
	return node.typ != None
}

func (n *Node) IsExist() bool {
	return n.Type() != None
}

func (n *Node) IsArray() bool {
	return n.parse().typ == Array
}

func (n *Node) IsObject() bool {
	return n.parse().typ == Object
}

func (n *Node) Get(path string) *Node {
	pPath := parsePath(path)
	if pPath.onlyRoot() {
		return n
	}
	pathNode := n
	for _, nodePath := range pPath.PathNoe {
		if n.err = pathNode.parse().Error(); n.err != nil {
			return &Node{}
		}
		node, ok := pathNode.children[nodePath]
		if !ok { // 没找到节点，直接返回
			return &Node{}
		}
		pathNode = node
	}
	if n.err = pathNode.parse().Error(); n.err != nil {
		return &Node{}
	}
	return pathNode
}

func (n *Node) Delete(path string) *Node {
	pPath := parsePath(path)
	if pPath.onlyRoot() {
		*n = Node{raw: "", parsed: false}
		return n
	}

	pathDepth := len(pPath.PathNoe)
	pathNode := n
	for depth, nodePath := range pPath.PathNoe {
		if n.err = pathNode.parse().Error(); n.err != nil {
			return n
		}
		node, ok := pathNode.children[nodePath]
		if !ok { // 没找到节点，直接返回
			return n
		}
		if depth < pathDepth-1 { // 非最后一级时，继续向后查找
			pathNode = node
			continue
		}
		pathNode.deleteObjectNode(nodePath)
	}
	return n
}

func (n *Node) insertObjectNode(nodePath string, node *Node) *Node {
	n.children[nodePath] = node
	endFlagIdx := len(n.block) - 1
	for endFlagIdx >= 0 {
		if n.block[endFlagIdx].Typ == dataTypeEndFlag {
			break
		}
		endFlagIdx--
	}
	if endFlagIdx < 0 {
		n.err = errors.New("inner error: end flag not found")
		return n
	}
	// 插入新增的block
	insertBlocks := []dataBlock{
		{Typ: dataTypeKey, Val: "\"" + nodePath + "\""},
		{Typ: dataTypeColon},
		{Typ: dataTypeVal},
		{Typ: dataTypeLineBreak},
	}
	n.block = append(n.block[:endFlagIdx], append(insertBlocks, n.block[endFlagIdx:]...)...)
	for i := endFlagIdx - 1; i >= 0; i-- { // 上一个val元素最后添加逗号
		if n.block[i].Typ == dataTypeComma {
			break
		}
		if n.block[i].Typ == dataTypeVal {
			n.block = append(n.block[:i+1], append([]dataBlock{{Typ: dataTypeComma}}, n.block[i+1:]...)...)
			break
		}
	}
	return n
}

func (n *Node) deleteObjectNode(nodePath string) *Node {
	// 寻找删除的节点
	_, ok := n.children[nodePath]
	if !ok {
		return n
	}
	delete(n.children, nodePath)
	// 删除关联的block信息
	keyIdx := 0
	for ; keyIdx < len(n.block); keyIdx++ {
		if n.block[keyIdx].Typ == dataTypeKey && n.block[keyIdx].KeyUnQuot() == nodePath {
			break
		}
	}
	// 没找到key所在的block，直接返回
	if keyIdx >= len(n.block) {
		return n
	}
	// 寻找开始删除的位置和结束删除的位置
	startIdx := keyIdx - 1 // 开始位置定位到最后一个Val/comma/CommentLine/StartFlag/LineBreak
	for ; startIdx > 0; startIdx-- {
		if n.block[startIdx].Is(dataTypeVal | dataTypeComma | dataTypeCommentLine | dataTypeStartFlag | dataTypeLineBreak) {
			break
		}
	}
	startIdx += 1
	// 寻找结束位置, 结束位置定位到下一个Key/EndFlag/Comment/LineBreak
	endIdx := keyIdx + 1
	for ; endIdx < len(n.block); endIdx++ {
		if n.block[endIdx].Is(dataTypeKey | dataTypeEndFlag | dataTypeComment | dataTypeLineBreak) {
			break
		}
	}
	// 删除节点
	n.block = append(n.block[:startIdx], n.block[endIdx:]...)
	return n
}

func (n *Node) Set(path string, val any) *Node {
	// val根据类型序列化
	data, err := json.Marshal(val)
	if err != nil {
		n.err = fmt.Errorf("marshal data error:%w", err)
		return n
	}
	return n.SetString(path, string(data))
}

func (n *Node) SetString(path string, val string) *Node {
	pPath := parsePath(path)
	if pPath.onlyRoot() {
		*n = Node{raw: val, parsed: false}
		return n
	}
	// 寻找插入位置，如果中间位置不存在，直接创建
	pathNode := n
	for i, nodePath := range pPath.PathNoe {
		if pathNode.parse().Error() != nil {
			n.err = pathNode.err
			return n
		}
		if pathNode.typ != Object {
			n.err = errors.New("path not found")
			return n
		}
		node, ok := pathNode.children[nodePath]
		if !ok {
			node = buildObjectNode()
			pathNode.children[nodePath] = node
			pathNode.insertObjectNode(nodePath, node)
		}
		pathNode = node
		if i != len(pPath.PathNoe)-1 {
			continue
		}
		// 最后一个节点，直接赋值
		pathNode.raw = val
		pathNode.parsed = false
	}
	return n
}

func buildObjectNode() *Node {
	return &Node{
		parsed:   true,
		typ:      Object,
		children: map[string]*Node{},
		block: []dataBlock{
			{Typ: dataTypeStartFlag},
			{Typ: dataTypeEndFlag},
		},
	}
}

type parsedPath struct {
	Root    bool
	PathNoe []string
}

func (pp parsedPath) onlyRoot() bool {
	return pp.Root && len(pp.PathNoe) == 0
}

func parsePath(path string) parsedPath {
	pathList := strings.Split(path, ".")
	if len(pathList) == 0 {
		return parsedPath{PathNoe: make([]string, 0)}
	}
	pPath := parsedPath{PathNoe: pathList}
	if pathList[0] == Root || (len(pathList) == 1 && pathList[0] == "") {
		pPath.Root = true
		pPath.PathNoe = pathList[1:]
	}
	return pPath
}

func (n *Node) ForEach(iterator func(key string, value *Node) bool) {
	if n.parse().Error() != nil {
		return
	}
	if n.typ != Object {
		iterator("", n)
		return
	}
	for _, blockInfo := range n.block {
		if blockInfo.Typ != dataTypeKey {
			continue
		}
		rawKey := blockInfo.KeyUnQuot()
		if !iterator(rawKey, n.children[rawKey]) {
			return
		}
	}
}
