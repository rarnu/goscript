package parser

import (
	"encoding/base64"
	"fmt"
	"github.com/go-sourcemap/sourcemap"
	"github.com/rarnu/goscript/ast"
	"github.com/rarnu/goscript/file"
	"github.com/rarnu/goscript/token"
	"os"
	"strings"
)

func (p *parser) parseBlockStatement() *ast.BlockStatement {
	node := &ast.BlockStatement{}
	node.LeftBrace = p.expect(token.LEFT_BRACE)
	node.List = p.parseStatementList()
	node.RightBrace = p.expect(token.RIGHT_BRACE)
	return node
}

func (p *parser) parseEmptyStatement() ast.Statement {
	idx := p.expect(token.SEMICOLON)
	return &ast.EmptyStatement{Semicolon: idx}
}

func (p *parser) parseStatementList() (list []ast.Statement) {
	for p.token != token.RIGHT_BRACE && p.token != token.EOF {
		p.scope.allowLet = true
		list = append(list, p.parseStatement())
	}
	return
}

func (p *parser) parseStatement() ast.Statement {
	if p.token == token.EOF {
		_ = p.errorUnexpectedToken(p.token)
		return &ast.BadStatement{From: p.idx, To: p.idx + 1}
	}
	switch p.token {
	case token.SEMICOLON:
		return p.parseEmptyStatement()
	case token.LEFT_BRACE:
		return p.parseBlockStatement()
	case token.IF:
		return p.parseIfStatement()
	case token.DO:
		return p.parseDoWhileStatement()
	case token.WHILE:
		return p.parseWhileStatement()
	case token.FOR:
		return p.parseForOrForInStatement()
	case token.BREAK:
		return p.parseBreakStatement()
	case token.CONTINUE:
		return p.parseContinueStatement()
	case token.DEBUGGER:
		return p.parseDebuggerStatement()
	case token.WITH:
		return p.parseWithStatement()
	case token.VAR:
		return p.parseVariableStatement()
	case token.LET:
		tok := p.peek()
		if tok == token.LEFT_BRACKET || p.scope.allowLet && (token.IsId(tok) || tok == token.LEFT_BRACE) {
			return p.parseLexicalDeclaration(p.token)
		}
		p.insertSemicolon = true
	case token.CONST:
		return p.parseLexicalDeclaration(p.token)
	case token.FUNCTION:
		return &ast.FunctionDeclaration{Function: p.parseFunction(true)}
	case token.CLASS:
		return &ast.ClassDeclaration{Class: p.parseClass(true)}
	case token.SWITCH:
		return p.parseSwitchStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	case token.THROW:
		return p.parseThrowStatement()
	case token.TRY:
		return p.parseTryStatement()
	}

	expression := p.parseExpression()

	if identifier, isIdentifier := expression.(*ast.Identifier); isIdentifier && p.token == token.COLON {
		// 标签语句
		colon := p.idx
		p.next() // :
		label := identifier.Name
		for _, value := range p.scope.labels {
			if label == value {
				_ = p.error(identifier.Idx0(), "Label '%s' already exists", label)
			}
		}
		p.scope.labels = append(p.scope.labels, label)
		p.scope.allowLet = false
		statement := p.parseStatement()
		p.scope.labels = p.scope.labels[:len(p.scope.labels)-1]
		return &ast.LabelledStatement{Label: identifier, Colon: colon, Statement: statement}
	}
	p.optionalSemicolon()
	return &ast.ExpressionStatement{Expression: expression}
}

func (p *parser) parseTryStatement() ast.Statement {
	node := &ast.TryStatement{Try: p.expect(token.TRY), Body: p.parseBlockStatement()}
	if p.token == token.CATCH {
		catch := p.idx
		p.next()
		var parameter ast.BindingTarget
		if p.token == token.LEFT_PARENTHESIS {
			p.next()
			parameter = p.parseBindingTarget()
			p.expect(token.RIGHT_PARENTHESIS)
		}
		node.Catch = &ast.CatchStatement{Catch: catch, Parameter: parameter, Body: p.parseBlockStatement()}
	}
	if p.token == token.FINALLY {
		p.next()
		node.Finally = p.parseBlockStatement()
	}
	if node.Catch == nil && node.Finally == nil {
		_ = p.error(node.Try, "Missing catch or finally after try")
		return &ast.BadStatement{From: node.Try, To: node.Body.Idx1()}
	}
	return node
}

func (p *parser) parseFunctionParameterList() *ast.ParameterList {
	opening := p.expect(token.LEFT_PARENTHESIS)
	var list []*ast.Binding
	var rest ast.Expression
	for p.token != token.RIGHT_PARENTHESIS && p.token != token.EOF {
		if p.token == token.ELLIPSIS {
			p.next()
			rest = p.reinterpretAsDestructBindingTarget(p.parseAssignmentExpression())
			break
		}
		p.parseVariableDeclaration(&list)
		if p.token != token.RIGHT_PARENTHESIS {
			p.expect(token.COMMA)
		}
	}
	closing := p.expect(token.RIGHT_PARENTHESIS)
	return &ast.ParameterList{Opening: opening, List: list, Rest: rest, Closing: closing}
}

func (p *parser) parseFunction(declaration bool) *ast.FunctionLiteral {
	node := &ast.FunctionLiteral{Function: p.expect(token.FUNCTION)}
	p.tokenToBindingId()
	var name *ast.Identifier
	if p.token == token.IDENTIFIER {
		name = p.parseIdentifier()
	} else if declaration {
		// 跳转到异常接收者
		_ = p.expect(token.IDENTIFIER)
	}
	node.Name = name
	node.ParameterList = p.parseFunctionParameterList()
	node.Body, node.DeclarationList = p.parseFunctionBlock()
	node.Source = p.slice(node.Idx0(), node.Idx1())
	return node
}

func (p *parser) parseFunctionBlock() (body *ast.BlockStatement, declarationList []*ast.VariableDeclaration) {
	p.openScope()
	inFunction := p.scope.inFunction
	p.scope.inFunction = true
	defer func() {
		p.scope.inFunction = inFunction
		p.closeScope()
	}()
	body = p.parseBlockStatement()
	declarationList = p.scope.declarationList
	return
}

func (p *parser) parseArrowFunctionBody() (ast.ConciseBody, []*ast.VariableDeclaration) {
	if p.token == token.LEFT_BRACE {
		return p.parseFunctionBlock()
	}
	return &ast.ExpressionBody{Expression: p.parseAssignmentExpression()}, nil
}

func (p *parser) parseClass(declaration bool) *ast.ClassLiteral {
	if !p.scope.allowLet && p.token == token.CLASS {
		_ = p.errorUnexpectedToken(token.CLASS)
	}
	node := &ast.ClassLiteral{Class: p.expect(token.CLASS)}
	p.tokenToBindingId()
	var name *ast.Identifier
	if p.token == token.IDENTIFIER {
		name = p.parseIdentifier()
	} else if declaration {
		_ = p.expect(token.IDENTIFIER)
	}
	node.Name = name
	if p.token != token.LEFT_BRACE {
		_ = p.expect(token.EXTENDS)
		node.SuperClass = p.parseLeftHandSideExpressionAllowCall()
	}
	_ = p.expect(token.LEFT_BRACE)

	for p.token != token.RIGHT_BRACE && p.token != token.EOF {
		if p.token == token.SEMICOLON {
			p.next()
			continue
		}
		start := p.idx
		static := false
		if p.token == token.STATIC {
			switch p.peek() {
			case token.ASSIGN, token.SEMICOLON, token.RIGHT_BRACE, token.LEFT_PARENTHESIS:
				// 作为标识符处理
			default:
				p.next()
				if p.token == token.LEFT_BRACE {
					b := &ast.ClassStaticBlock{Static: start}
					b.Block, b.DeclarationList = p.parseFunctionBlock()
					b.Source = p.slice(b.Block.LeftBrace, b.Block.Idx1())
					node.Body = append(node.Body, b)
					continue
				}
				static = true
			}
		}
		var kind ast.PropertyKind
		methodBodyStart := p.idx
		if p.literal == "get" || p.literal == "set" {
			if p.peek() != token.LEFT_PARENTHESIS {
				if p.literal == "get" {
					kind = ast.PropertyKindGet
				} else {
					kind = ast.PropertyKindSet
				}
				p.next()
			}
		}
		_, keyName, value, tkn := p.parseObjectPropertyKey()
		if value == nil {
			continue
		}
		computed := tkn == token.ILLEGAL
		_, private := value.(*ast.PrivateIdentifier)
		if static && !private && keyName == "prototype" {
			_ = p.error(value.Idx0(), "Classes may not have a static property named 'prototype'")
		}
		if kind == "" && p.token == token.LEFT_PARENTHESIS {
			kind = ast.PropertyKindMethod
		}
		if kind != "" {
			// 方法
			if keyName == "constructor" {
				if !computed && !static && kind != ast.PropertyKindMethod {
					_ = p.error(value.Idx0(), "Class constructor may not be an accessor")
				} else if private {
					_ = p.error(value.Idx0(), "Class constructor may not be a private method")
				}
			}
			md := &ast.MethodDefinition{Idx: start, Key: value, Kind: kind, Body: p.parseMethodDefinition(methodBodyStart, kind), Static: static, Computed: computed}
			node.Body = append(node.Body, md)
		} else {
			// 字段
			isCtor := !computed && keyName == "constructor"
			if !isCtor {
				if n, ok := value.(*ast.PrivateIdentifier); ok {
					isCtor = n.Name == "constructor"
				}
			}
			if isCtor {
				_ = p.error(value.Idx0(), "Classes may not have a field named 'constructor'")
			}
			var initializer ast.Expression
			if p.token == token.ASSIGN {
				p.next()
				initializer = p.parseExpression()
			}
			if !p.implicitSemicolon && p.token != token.SEMICOLON && p.token != token.RIGHT_BRACE {
				_ = p.errorUnexpectedToken(p.token)
				break
			}
			node.Body = append(node.Body, &ast.FieldDefinition{Idx: start, Key: value, Initializer: initializer, Static: static, Computed: computed})
		}
	}
	node.RightBrace = p.expect(token.RIGHT_BRACE)
	node.Source = p.slice(node.Class, node.RightBrace+1)
	return node
}

func (p *parser) parseDebuggerStatement() ast.Statement {
	idx := p.expect(token.DEBUGGER)
	node := &ast.DebuggerStatement{Debugger: idx}
	p.semicolon()
	return node
}

func (p *parser) parseReturnStatement() ast.Statement {
	idx := p.expect(token.RETURN)
	if !p.scope.inFunction {
		_ = p.error(idx, "Illegal return statement")
		p.nextStatement()
		return &ast.BadStatement{From: idx, To: p.idx}
	}
	node := &ast.ReturnStatement{Return: idx}
	if !p.implicitSemicolon && p.token != token.SEMICOLON && p.token != token.RIGHT_BRACE && p.token != token.EOF {
		node.Argument = p.parseExpression()
	}
	p.semicolon()
	return node
}

func (p *parser) parseThrowStatement() ast.Statement {
	idx := p.expect(token.THROW)
	if p.implicitSemicolon {
		if p.chr == -1 { // Hackish
			_ = p.error(idx, "Unexpected end of input")
		} else {
			_ = p.error(idx, "Illegal newline after throw")
		}
		p.nextStatement()
		return &ast.BadStatement{From: idx, To: p.idx}
	}
	node := &ast.ThrowStatement{Throw: idx, Argument: p.parseExpression()}
	p.semicolon()
	return node
}

func (p *parser) parseSwitchStatement() ast.Statement {
	_ = p.expect(token.SWITCH)
	_ = p.expect(token.LEFT_PARENTHESIS)
	node := &ast.SwitchStatement{Discriminant: p.parseExpression(), Default: -1}
	_ = p.expect(token.RIGHT_PARENTHESIS)
	_ = p.expect(token.LEFT_BRACE)
	inSwitch := p.scope.inSwitch
	p.scope.inSwitch = true
	defer func() {
		p.scope.inSwitch = inSwitch
	}()

	for index := 0; p.token != token.EOF; index++ {
		if p.token == token.RIGHT_BRACE {
			p.next()
			break
		}
		clause := p.parseCaseStatement()
		if clause.Test == nil {
			if node.Default != -1 {
				_ = p.error(clause.Case, "Already saw a default in switch")
			}
			node.Default = index
		}
		node.Body = append(node.Body, clause)
	}
	return node
}

func (p *parser) parseWithStatement() ast.Statement {
	_ = p.expect(token.WITH)
	_ = p.expect(token.LEFT_PARENTHESIS)
	node := &ast.WithStatement{Object: p.parseExpression()}
	_ = p.expect(token.RIGHT_PARENTHESIS)
	p.scope.allowLet = false
	node.Body = p.parseStatement()
	return node
}

func (p *parser) parseCaseStatement() *ast.CaseStatement {
	node := &ast.CaseStatement{Case: p.idx}
	if p.token == token.DEFAULT {
		p.next()
	} else {
		_ = p.expect(token.CASE)
		node.Test = p.parseExpression()
	}
	_ = p.expect(token.COLON)
	for {
		if p.token == token.EOF || p.token == token.RIGHT_BRACE || p.token == token.CASE || p.token == token.DEFAULT {
			break
		}
		node.Consequent = append(node.Consequent, p.parseStatement())
	}
	return node
}

func (p *parser) parseIterationStatement() ast.Statement {
	inIteration := p.scope.inIteration
	p.scope.inIteration = true
	defer func() {
		p.scope.inIteration = inIteration
	}()
	p.scope.allowLet = false
	return p.parseStatement()
}

func (p *parser) parseForIn(idx file.Idx, into ast.ForInto) *ast.ForInStatement {
	// 已经消费了 "<into> in"
	source := p.parseExpression()
	_ = p.expect(token.RIGHT_PARENTHESIS)
	return &ast.ForInStatement{For: idx, Into: into, Source: source, Body: p.parseIterationStatement()}
}

func (p *parser) parseForOf(idx file.Idx, into ast.ForInto) *ast.ForOfStatement {
	// 已经消费了 "<into> of"
	source := p.parseAssignmentExpression()
	_ = p.expect(token.RIGHT_PARENTHESIS)
	return &ast.ForOfStatement{For: idx, Into: into, Source: source, Body: p.parseIterationStatement()}
}

func (p *parser) parseFor(idx file.Idx, initializer ast.ForLoopInitializer) *ast.ForStatement {
	// 已经消费了 "<initializer> ;"
	var test, update ast.Expression
	if p.token != token.SEMICOLON {
		test = p.parseExpression()
	}
	_ = p.expect(token.SEMICOLON)
	if p.token != token.RIGHT_PARENTHESIS {
		update = p.parseExpression()
	}
	_ = p.expect(token.RIGHT_PARENTHESIS)
	return &ast.ForStatement{For: idx, Initializer: initializer, Test: test, Update: update, Body: p.parseIterationStatement()}
}

func (p *parser) parseForOrForInStatement() ast.Statement {
	idx := p.expect(token.FOR)
	_ = p.expect(token.LEFT_PARENTHESIS)
	var initializer ast.ForLoopInitializer
	forIn := false
	forOf := false
	var into ast.ForInto
	if p.token != token.SEMICOLON {
		allowIn := p.scope.allowIn
		p.scope.allowIn = false
		tok := p.token
		if tok == token.LET {
			switch p.peek() {
			case token.IDENTIFIER, token.LEFT_BRACKET, token.LEFT_BRACE:
			default:
				tok = token.IDENTIFIER
			}
		}
		if tok == token.VAR || tok == token.LET || tok == token.CONST {
			idx := p.idx
			p.next()
			var list []*ast.Binding
			if tok == token.VAR {
				list = p.parseVarDeclarationList(idx)
			} else {
				list = p.parseVariableDeclarationList()
			}
			if len(list) == 1 {
				if p.token == token.IN {
					p.next()
					forIn = true
				} else if p.token == token.IDENTIFIER && p.literal == "of" {
					p.next()
					forOf = true
				}
			}
			if forIn || forOf {
				if list[0].Initializer != nil {
					_ = p.error(list[0].Initializer.Idx0(), "for-in loop variable declaration may not have an initializer")
				}
				if tok == token.VAR {
					into = &ast.ForIntoVar{Binding: list[0]}
				} else {
					into = &ast.ForDeclaration{Idx: idx, IsConst: tok == token.CONST, Target: list[0].Target}
				}
			} else {
				p.ensurePatternInit(list)
				if tok == token.VAR {
					initializer = &ast.ForLoopInitializerVarDeclList{List: list}
				} else {
					initializer = &ast.ForLoopInitializerLexicalDecl{LexicalDeclaration: ast.LexicalDeclaration{Idx: idx, Token: tok, List: list}}
				}
			}
		} else {
			expr := p.parseExpression()
			if p.token == token.IN {
				p.next()
				forIn = true
			} else if p.token == token.IDENTIFIER && p.literal == "of" {
				p.next()
				forOf = true
			}
			if forIn || forOf {
				switch e := expr.(type) {
				case *ast.Identifier, *ast.DotExpression, *ast.PrivateDotExpression, *ast.BracketExpression, *ast.Binding:
					// 都是可以接受的类型
				case *ast.ObjectLiteral:
					expr = p.reinterpretAsObjectAssignmentPattern(e)
				case *ast.ArrayLiteral:
					expr = p.reinterpretAsArrayAssignmentPattern(e)
				default:
					_ = p.error(idx, "Invalid left-hand side in for-in or for-of")
					p.nextStatement()
					return &ast.BadStatement{From: idx, To: p.idx}
				}
				into = &ast.ForIntoExpression{Expression: expr}
			} else {
				initializer = &ast.ForLoopInitializerExpression{Expression: expr}
			}
		}
		p.scope.allowIn = allowIn
	}
	if forIn {
		return p.parseForIn(idx, into)
	}
	if forOf {
		return p.parseForOf(idx, into)
	}
	_ = p.expect(token.SEMICOLON)
	return p.parseFor(idx, initializer)
}

func (p *parser) ensurePatternInit(list []*ast.Binding) {
	for _, item := range list {
		if _, ok := item.Target.(ast.Pattern); ok {
			if item.Initializer == nil {
				_ = p.error(item.Idx1(), "Missing initializer in destructuring declaration")
				break
			}
		}
	}
}

func (p *parser) parseVariableStatement() *ast.VariableStatement {
	idx := p.expect(token.VAR)
	list := p.parseVarDeclarationList(idx)
	p.ensurePatternInit(list)
	p.semicolon()
	return &ast.VariableStatement{Var: idx, List: list}
}

func (p *parser) parseLexicalDeclaration(tok token.Token) *ast.LexicalDeclaration {
	idx := p.expect(tok)
	if !p.scope.allowLet {
		_ = p.error(idx, "Lexical declaration cannot appear in a single-statement context")
	}
	list := p.parseVariableDeclarationList()
	p.ensurePatternInit(list)
	p.semicolon()
	return &ast.LexicalDeclaration{Idx: idx, Token: tok, List: list}
}

func (p *parser) parseDoWhileStatement() ast.Statement {
	inIteration := p.scope.inIteration
	p.scope.inIteration = true
	defer func() {
		p.scope.inIteration = inIteration
	}()
	_ = p.expect(token.DO)
	node := &ast.DoWhileStatement{}
	if p.token == token.LEFT_BRACE {
		node.Body = p.parseBlockStatement()
	} else {
		p.scope.allowLet = false
		node.Body = p.parseStatement()
	}
	_ = p.expect(token.WHILE)
	_ = p.expect(token.LEFT_PARENTHESIS)
	node.Test = p.parseExpression()
	_ = p.expect(token.RIGHT_PARENTHESIS)
	if p.token == token.SEMICOLON {
		p.next()
	}
	return node
}

func (p *parser) parseWhileStatement() ast.Statement {
	_ = p.expect(token.WHILE)
	_ = p.expect(token.LEFT_PARENTHESIS)
	node := &ast.WhileStatement{Test: p.parseExpression()}
	_ = p.expect(token.RIGHT_PARENTHESIS)
	node.Body = p.parseIterationStatement()
	return node
}

func (p *parser) parseIfStatement() ast.Statement {
	_ = p.expect(token.IF)
	_ = p.expect(token.LEFT_PARENTHESIS)
	node := &ast.IfStatement{Test: p.parseExpression()}
	_ = p.expect(token.RIGHT_PARENTHESIS)
	if p.token == token.LEFT_BRACE {
		node.Consequent = p.parseBlockStatement()
	} else {
		p.scope.allowLet = false
		node.Consequent = p.parseStatement()
	}
	if p.token == token.ELSE {
		p.next()
		p.scope.allowLet = false
		node.Alternate = p.parseStatement()
	}
	return node
}

func (p *parser) parseSourceElements() (body []ast.Statement) {
	for p.token != token.EOF {
		p.scope.allowLet = true
		body = append(body, p.parseStatement())
	}
	return body
}

func (p *parser) parseProgram() *ast.Program {
	p.openScope()
	defer p.closeScope()
	prg := &ast.Program{Body: p.parseSourceElements(), DeclarationList: p.scope.declarationList, File: p.file}
	p.file.SetSourceMap(p.parseSourceMap())
	return prg
}

func extractSourceMapLine(str string) string {
	for {
		p := strings.LastIndexByte(str, '\n')
		line := str[p+1:]
		if line != "" && line != "})" {
			if strings.HasPrefix(line, "//# sourceMappingURL=") {
				return line
			}
			break
		}
		if p >= 0 {
			str = str[:p]
		} else {
			break
		}
	}
	return ""
}

func (p *parser) parseSourceMap() *sourcemap.Consumer {
	if p.opts.disableSourceMaps {
		return nil
	}
	if smLine := extractSourceMapLine(p.str); smLine != "" {
		urlIndex := strings.Index(smLine, "=")
		urlStr := smLine[urlIndex+1:]
		var data []byte
		var err error
		if strings.HasPrefix(urlStr, "data:application/json") {
			b64Index := strings.Index(urlStr, ",")
			b64 := urlStr[b64Index+1:]
			data, err = base64.StdEncoding.DecodeString(b64)
		} else {
			if sourceURL := file.ResolveSourcemapURL(p.file.Name(), urlStr); sourceURL != nil {
				if p.opts.sourceMapLoader != nil {
					data, err = p.opts.sourceMapLoader(sourceURL.String())
				} else {
					if sourceURL.Scheme == "" || sourceURL.Scheme == "file" {
						data, err = os.ReadFile(sourceURL.Path)
					} else {
						err = fmt.Errorf("unsupported source map URL scheme: %s", sourceURL.Scheme)
					}
				}
			}
		}

		if err != nil {
			_ = p.error(file.Idx(0), "Could not load source map: %v", err)
			return nil
		}
		if data == nil {
			return nil
		}
		if sm, err := sourcemap.Parse(p.file.Name(), data); err == nil {
			return sm
		} else {
			_ = p.error(file.Idx(0), "Could not parse source map: %v", err)
		}
	}
	return nil
}

func (p *parser) parseBreakStatement() ast.Statement {
	idx := p.expect(token.BREAK)
	semicolon := p.implicitSemicolon
	if p.token == token.SEMICOLON {
		semicolon = true
		p.next()
	}
	if semicolon || p.token == token.RIGHT_BRACE {
		p.implicitSemicolon = false
		if !p.scope.inIteration && !p.scope.inSwitch {
			goto illegal
		}
		return &ast.BranchStatement{Idx: idx, Token: token.BREAK}
	}
	p.tokenToBindingId()
	if p.token == token.IDENTIFIER {
		identifier := p.parseIdentifier()
		if !p.scope.hasLabel(identifier.Name) {
			_ = p.error(idx, "Undefined label '%s'", identifier.Name)
			return &ast.BadStatement{From: idx, To: identifier.Idx1()}
		}
		p.semicolon()
		return &ast.BranchStatement{Idx: idx, Token: token.BREAK, Label: identifier}
	}
	_ = p.expect(token.IDENTIFIER)

illegal:
	_ = p.error(idx, "Illegal break statement")
	p.nextStatement()
	return &ast.BadStatement{From: idx, To: p.idx}
}

func (p *parser) parseContinueStatement() ast.Statement {
	idx := p.expect(token.CONTINUE)
	semicolon := p.implicitSemicolon
	if p.token == token.SEMICOLON {
		semicolon = true
		p.next()
	}
	if semicolon || p.token == token.RIGHT_BRACE {
		p.implicitSemicolon = false
		if !p.scope.inIteration {
			goto illegal
		}
		return &ast.BranchStatement{Idx: idx, Token: token.CONTINUE}
	}

	p.tokenToBindingId()
	if p.token == token.IDENTIFIER {
		identifier := p.parseIdentifier()
		if !p.scope.hasLabel(identifier.Name) {
			_ = p.error(idx, "Undefined label '%s'", identifier.Name)
			return &ast.BadStatement{From: idx, To: identifier.Idx1()}
		}
		if !p.scope.inIteration {
			goto illegal
		}
		p.semicolon()
		return &ast.BranchStatement{Idx: idx, Token: token.CONTINUE, Label: identifier}
	}
	_ = p.expect(token.IDENTIFIER)
illegal:
	_ = p.error(idx, "Illegal continue statement")
	p.nextStatement()
	return &ast.BadStatement{From: idx, To: p.idx}
}

// nextStatement 在出错后寻找下一条语句(recover)
func (p *parser) nextStatement() {
	for {
		switch p.token {
		case token.BREAK, token.CONTINUE,
			token.FOR, token.IF, token.RETURN, token.SWITCH,
			token.VAR, token.DO, token.TRY, token.WITH,
			token.WHILE, token.THROW, token.CATCH, token.FINALLY:
			// 只有当 parser 自上次同步以来发生过变化，或者下一个 10 次调用都没有进展时，才会返回
			// 否则至少要消耗一个令牌，以避免 parser 陷入死循环
			if p.idx == p.recover.idx && p.recover.count < 10 {
				p.recover.count++
				return
			}
			if p.idx > p.recover.idx {
				p.recover.idx = p.idx
				p.recover.count = 0
				return
			}
			// 如果代码走到这里，表示 parser 产生了错误，很可能是这个函数中的标记列表不正确
		case token.EOF:
			return
		}
		p.next()
	}
}
