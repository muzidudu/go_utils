// Package template 自定义 tag 示例
package template

import (
	"bytes"
	"strings"

	"github.com/flosch/pongo2/v6"
)

// TagUppercaseParser 示例：uppercase tag 解析器
// 使用: engine.RegisterTag("uppercase", template.TagUppercaseParser)
// 模板: {% uppercase %}hello{% enduppercase %} -> HELLO

type tagUppercaseNode struct {
	body pongo2.INode
}

func (n *tagUppercaseNode) Execute(ctx *pongo2.ExecutionContext, writer pongo2.TemplateWriter) *pongo2.Error {
	buf := bytes.NewBuffer(nil)
	if err := n.body.Execute(ctx, buf); err != nil {
		return err
	}
	_, err := writer.WriteString(strings.ToUpper(buf.String()))
	if err != nil {
		return ctx.Error(err.Error(), nil)
	}
	return nil
}

// TagUppercaseParser 解析 uppercase 标签
func TagUppercaseParser(doc *pongo2.Parser, start *pongo2.Token, arguments *pongo2.Parser) (pongo2.INodeTag, *pongo2.Error) {
	wrapper, endargs, err := doc.WrapUntilTag("enduppercase")
	if err != nil {
		return nil, err
	}
	if arguments.Count() != 0 {
		return nil, arguments.Error("Tag 'uppercase' does not take any argument.", nil)
	}
	if endargs.Count() != 0 {
		return nil, endargs.Error("Arguments not allowed here.", nil)
	}
	return &tagUppercaseNode{body: wrapper}, nil
}
