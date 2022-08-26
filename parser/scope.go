package parser

import (
	"github.com/rarnu/goscript/ast"
	"github.com/rarnu/goscript/unistring"
)

type scope struct {
	outer           *scope
	allowIn         bool
	allowLet        bool
	inIteration     bool
	inSwitch        bool
	inFunction      bool
	declarationList []*ast.VariableDeclaration
	labels          []unistring.String
}

func (p *parser) openScope() {
	p.scope = &scope{
		outer:   p.scope,
		allowIn: true,
	}
}

func (p *parser) closeScope() {
	p.scope = p.scope.outer
}

func (s *scope) declare(declaration *ast.VariableDeclaration) {
	s.declarationList = append(s.declarationList, declaration)
}

func (s *scope) hasLabel(name unistring.String) bool {
	for _, label := range s.labels {
		if label == name {
			return true
		}
	}
	if s.outer != nil && !s.inFunction {
		// 不允许越过函数边界寻找标签
		return s.outer.hasLabel(name)
	}
	return false
}
