package parser

import (
	"github.com/rarnu/goscript/ast"
	"github.com/rarnu/goscript/file"
	"github.com/rarnu/goscript/token"
	"github.com/rarnu/goscript/unistring"
	"strings"
)

func (p *parser) parseIdentifier() *ast.Identifier {
	literal := p.parsedLiteral
	idx := p.idx
	p.next()
	return &ast.Identifier{
		Name: literal,
		Idx:  idx,
	}
}

func (p *parser) parsePrimaryExpression() ast.Expression {
	literal, parsedLiteral := p.literal, p.parsedLiteral
	idx := p.idx
	switch p.token {
	case token.IDENTIFIER:
		p.next()
		return &ast.Identifier{Name: parsedLiteral, Idx: idx}
	case token.NULL:
		p.next()
		return &ast.NullLiteral{Idx: idx, Literal: literal}
	case token.BOOLEAN:
		p.next()
		value := false
		switch parsedLiteral {
		case "true":
			value = true
		case "false":
			value = false
		default:
			_ = p.error(idx, "Illegal boolean literal")
		}
		return &ast.BooleanLiteral{Idx: idx, Literal: literal, Value: value}
	case token.STRING:
		p.next()
		return &ast.StringLiteral{Idx: idx, Literal: literal, Value: parsedLiteral}
	case token.NUMBER:
		p.next()
		value, err := parseNumberLiteral(literal)
		if err != nil {
			_ = p.error(idx, err.Error())
			value = 0
		}
		return &ast.NumberLiteral{Idx: idx, Literal: literal, Value: value}
	case token.SLASH, token.QUOTIENT_ASSIGN:
		return p.parseRegExpLiteral()
	case token.LEFT_BRACE:
		return p.parseObjectLiteral()
	case token.LEFT_BRACKET:
		return p.parseArrayLiteral()
	case token.LEFT_PARENTHESIS:
		return p.parseParenthesisedExpression()
	case token.BACKTICK:
		return p.parseTemplateLiteral(false)
	case token.THIS:
		p.next()
		return &ast.ThisExpression{Idx: idx}
	case token.SUPER:
		return p.parseSuperProperty()
	case token.FUNCTION:
		return p.parseFunction(false)
	case token.CLASS:
		return p.parseClass(false)
	}
	if isBindingId(p.token, parsedLiteral) {
		p.next()
		return &ast.Identifier{Name: parsedLiteral, Idx: idx}
	}
	_ = p.errorUnexpectedToken(p.token)
	p.nextStatement()
	return &ast.BadExpression{From: idx, To: p.idx}
}

func (p *parser) parseSuperProperty() ast.Expression {
	idx := p.idx
	p.next()
	switch p.token {
	case token.PERIOD:
		p.next()
		if !token.IsId(p.token) {
			_ = p.expect(token.IDENTIFIER)
			p.nextStatement()
			return &ast.BadExpression{From: idx, To: p.idx}
		}
		idIdx := p.idx
		parsedLiteral := p.parsedLiteral
		p.next()
		return &ast.DotExpression{
			Left:       &ast.SuperExpression{Idx: idx},
			Identifier: ast.Identifier{Name: parsedLiteral, Idx: idIdx},
		}
	case token.LEFT_BRACKET:
		return p.parseBracketMember(&ast.SuperExpression{Idx: idx})
	case token.LEFT_PARENTHESIS:
		return p.parseCallExpression(&ast.SuperExpression{Idx: idx})
	default:
		_ = p.error(idx, "'super' keyword unexpected here")
		p.nextStatement()
		return &ast.BadExpression{From: idx, To: p.idx}
	}
}

func (p *parser) reinterpretSequenceAsArrowFuncParams(seq *ast.SequenceExpression) *ast.ParameterList {
	firstRestIdx := -1
	params := make([]*ast.Binding, 0, len(seq.Sequence))
	for i, item := range seq.Sequence {
		if _, ok := item.(*ast.SpreadElement); ok {
			if firstRestIdx == -1 {
				firstRestIdx = i
				continue
			}
		}
		if firstRestIdx != -1 {
			_ = p.error(seq.Sequence[firstRestIdx].Idx0(), "Rest parameter must be last formal parameter")
			return &ast.ParameterList{}
		}
		params = append(params, p.reinterpretAsBinding(item))
	}
	var rest ast.Expression
	if firstRestIdx != -1 {
		rest = p.reinterpretAsBindingRestElement(seq.Sequence[firstRestIdx])
	}
	return &ast.ParameterList{
		List: params,
		Rest: rest,
	}
}

func (p *parser) parseParenthesisedExpression() ast.Expression {
	opening := p.idx
	p.expect(token.LEFT_PARENTHESIS)
	var list []ast.Expression
	if p.token != token.RIGHT_PARENTHESIS {
		for {
			if p.token == token.ELLIPSIS {
				start := p.idx
				_ = p.errorUnexpectedToken(token.ELLIPSIS)
				p.next()
				expr := p.parseAssignmentExpression()
				list = append(list, &ast.BadExpression{From: start, To: expr.Idx1()})
			} else {
				list = append(list, p.parseAssignmentExpression())
			}
			if p.token != token.COMMA {
				break
			}
			p.next()
			if p.token == token.RIGHT_PARENTHESIS {
				_ = p.errorUnexpectedToken(token.RIGHT_PARENTHESIS)
				break
			}
		}
	}
	_ = p.expect(token.RIGHT_PARENTHESIS)
	if len(list) == 1 && len(p.errors) == 0 {
		return list[0]
	}
	if len(list) == 0 {
		_ = p.errorUnexpectedToken(token.RIGHT_PARENTHESIS)
		return &ast.BadExpression{From: opening, To: p.idx}
	}
	return &ast.SequenceExpression{Sequence: list}
}

func (p *parser) parseRegExpLiteral() *ast.RegExpLiteral {

	offset := p.chrOffset - 1 // Opening slash already gotten
	if p.token == token.QUOTIENT_ASSIGN {
		offset -= 1 // =
	}
	idx := p.idxOf(offset)

	pattern, _, err := p.scanString(offset, false)
	endOffset := p.chrOffset

	if err == "" {
		pattern = pattern[1 : len(pattern)-1]
	}

	flags := ""
	if !isLineTerminator(p.chr) && !isLineWhiteSpace(p.chr) {
		p.next()

		if p.token == token.IDENTIFIER { // gim

			flags = p.literal
			p.next()
			endOffset = p.chrOffset - 1
		}
	} else {
		p.next()
	}

	literal := p.str[offset:endOffset]

	return &ast.RegExpLiteral{
		Idx:     idx,
		Literal: literal,
		Pattern: pattern,
		Flags:   flags,
	}
}

func isBindingId(tok token.Token, parsedLiteral unistring.String) bool {
	if tok == token.IDENTIFIER {
		return true
	}
	if token.IsId(tok) {
		switch parsedLiteral {
		case "yield", "await":
			return true
		}
		if token.IsUnreservedWord(tok) {
			return true
		}
	}
	return false
}

func (p *parser) tokenToBindingId() {
	if isBindingId(p.token, p.parsedLiteral) {
		p.token = token.IDENTIFIER
	}
}

func (p *parser) parseBindingTarget() (target ast.BindingTarget) {
	p.tokenToBindingId()
	switch p.token {
	case token.IDENTIFIER:
		target = &ast.Identifier{Name: p.parsedLiteral, Idx: p.idx}
		p.next()
	case token.LEFT_BRACKET:
		target = p.parseArrayBindingPattern()
	case token.LEFT_BRACE:
		target = p.parseObjectBindingPattern()
	default:
		idx := p.expect(token.IDENTIFIER)
		p.nextStatement()
		target = &ast.BadExpression{From: idx, To: p.idx}
	}
	return
}

func (p *parser) parseVariableDeclaration(declarationList *[]*ast.Binding) ast.Expression {
	node := &ast.Binding{Target: p.parseBindingTarget()}
	if declarationList != nil {
		*declarationList = append(*declarationList, node)
	}
	if p.token == token.ASSIGN {
		p.next()
		node.Initializer = p.parseAssignmentExpression()
	}
	return node
}

func (p *parser) parseVariableDeclarationList() (declarationList []*ast.Binding) {
	for {
		p.parseVariableDeclaration(&declarationList)
		if p.token != token.COMMA {
			break
		}
		p.next()
	}
	return
}

func (p *parser) parseVarDeclarationList(v file.Idx) []*ast.Binding {
	declarationList := p.parseVariableDeclarationList()
	p.scope.declare(&ast.VariableDeclaration{Var: v, List: declarationList})
	return declarationList
}

func (p *parser) parseObjectPropertyKey() (string, unistring.String, ast.Expression, token.Token) {
	if p.token == token.LEFT_BRACKET {
		p.next()
		expr := p.parseAssignmentExpression()
		_ = p.expect(token.RIGHT_BRACKET)
		return "", "", expr, token.ILLEGAL
	}
	idx, tkn, literal, parsedLiteral := p.idx, p.token, p.literal, p.parsedLiteral
	var value ast.Expression
	p.next()
	switch tkn {
	case token.IDENTIFIER, token.STRING, token.KEYWORD, token.ESCAPED_RESERVED_WORD:
		value = &ast.StringLiteral{Idx: idx, Literal: literal, Value: parsedLiteral}
	case token.NUMBER:
		num, err := parseNumberLiteral(literal)
		if err != nil {
			_ = p.error(idx, err.Error())
		} else {
			value = &ast.NumberLiteral{Idx: idx, Literal: literal, Value: num}
		}
	case token.PRIVATE_IDENTIFIER:
		value = &ast.PrivateIdentifier{Identifier: ast.Identifier{Idx: idx, Name: parsedLiteral}}
	default:
		if token.IsId(tkn) {
			value = &ast.StringLiteral{Idx: idx, Literal: literal, Value: unistring.String(literal)}
		} else {
			_ = p.errorUnexpectedToken(tkn)
		}
	}
	return literal, parsedLiteral, value, tkn
}

func (p *parser) parseObjectProperty() ast.Property {
	if p.token == token.ELLIPSIS {
		p.next()
		return &ast.SpreadElement{Expression: p.parseAssignmentExpression()}
	}
	keyStartIdx := p.idx
	literal, parsedLiteral, value, tkn := p.parseObjectPropertyKey()
	if value == nil {
		return nil
	}
	if token.IsId(tkn) || tkn == token.STRING || tkn == token.ILLEGAL {
		switch {
		case p.token == token.LEFT_PARENTHESIS:
			parameterList := p.parseFunctionParameterList()
			node := &ast.FunctionLiteral{Function: keyStartIdx, ParameterList: parameterList}
			node.Body, node.DeclarationList = p.parseFunctionBlock()
			node.Source = p.slice(keyStartIdx, node.Body.Idx1())
			return &ast.PropertyKeyed{Key: value, Kind: ast.PropertyKindMethod, Value: node}
		case p.token == token.COMMA || p.token == token.RIGHT_BRACE || p.token == token.ASSIGN:
			if isBindingId(tkn, parsedLiteral) {
				var initializer ast.Expression
				if p.token == token.ASSIGN {
					// 允许有这里使用初始化语法，以防对象的字面量需要被重新解析为赋值模式
					p.next()
					initializer = p.parseAssignmentExpression()
				}
				return &ast.PropertyShort{Name: ast.Identifier{Name: parsedLiteral, Idx: value.Idx0()}, Initializer: initializer}
			} else {
				_ = p.errorUnexpectedToken(p.token)
			}
		case (literal == "get" || literal == "set") && p.token != token.COLON:
			_, _, keyValue, _ := p.parseObjectPropertyKey()
			if keyValue == nil {
				return nil
			}
			var kind ast.PropertyKind
			if literal == "get" {
				kind = ast.PropertyKindGet
			} else {
				kind = ast.PropertyKindSet
			}
			return &ast.PropertyKeyed{Key: keyValue, Kind: kind, Value: p.parseMethodDefinition(keyStartIdx, kind)}
		}
	}
	_ = p.expect(token.COLON)
	return &ast.PropertyKeyed{Key: value, Kind: ast.PropertyKindValue, Value: p.parseAssignmentExpression(), Computed: tkn == token.ILLEGAL}
}

func (p *parser) parseMethodDefinition(keyStartIdx file.Idx, kind ast.PropertyKind) *ast.FunctionLiteral {
	idx1 := p.idx
	parameterList := p.parseFunctionParameterList()
	switch kind {
	case ast.PropertyKindGet:
		if len(parameterList.List) > 0 || parameterList.Rest != nil {
			_ = p.error(idx1, "Getter must not have any formal parameters.")
		}
	case ast.PropertyKindSet:
		if len(parameterList.List) != 1 || parameterList.Rest != nil {
			_ = p.error(idx1, "Setter must have exactly one formal parameter.")
		}
	}
	node := &ast.FunctionLiteral{Function: keyStartIdx, ParameterList: parameterList}
	node.Body, node.DeclarationList = p.parseFunctionBlock()
	node.Source = p.slice(keyStartIdx, node.Body.Idx1())
	return node
}

func (p *parser) parseObjectLiteral() *ast.ObjectLiteral {
	var value []ast.Property
	idx0 := p.expect(token.LEFT_BRACE)
	for p.token != token.RIGHT_BRACE && p.token != token.EOF {
		property := p.parseObjectProperty()
		if property != nil {
			value = append(value, property)
		}
		if p.token != token.RIGHT_BRACE {
			p.expect(token.COMMA)
		} else {
			break
		}
	}
	idx1 := p.expect(token.RIGHT_BRACE)

	return &ast.ObjectLiteral{LeftBrace: idx0, RightBrace: idx1, Value: value}
}

func (p *parser) parseArrayLiteral() *ast.ArrayLiteral {
	idx0 := p.expect(token.LEFT_BRACKET)
	var value []ast.Expression
	for p.token != token.RIGHT_BRACKET && p.token != token.EOF {
		if p.token == token.COMMA {
			p.next()
			value = append(value, nil)
			continue
		}
		if p.token == token.ELLIPSIS {
			p.next()
			value = append(value, &ast.SpreadElement{Expression: p.parseAssignmentExpression()})
		} else {
			value = append(value, p.parseAssignmentExpression())
		}
		if p.token != token.RIGHT_BRACKET {
			p.expect(token.COMMA)
		}
	}
	idx1 := p.expect(token.RIGHT_BRACKET)
	return &ast.ArrayLiteral{LeftBracket: idx0, RightBracket: idx1, Value: value}
}

func (p *parser) parseTemplateLiteral(tagged bool) *ast.TemplateLiteral {
	res := &ast.TemplateLiteral{OpenQuote: p.idx}
	for {
		start := p.offset
		literal, parsed, finished, parseErr, err := p.parseTemplateCharacters()
		if err != "" {
			_ = p.error(p.offset, err)
		}
		res.Elements = append(res.Elements, &ast.TemplateElement{Idx: p.idxOf(start), Literal: literal, Parsed: parsed, Valid: parseErr == ""})
		if !tagged && parseErr != "" {
			_ = p.error(p.offset, parseErr)
		}
		end := p.chrOffset - 1
		p.next()
		if finished {
			res.CloseQuote = p.idxOf(end)
			break
		}
		expr := p.parseExpression()
		res.Expressions = append(res.Expressions, expr)
		if p.token != token.RIGHT_BRACE {
			_ = p.errorUnexpectedToken(p.token)
		}
	}
	return res
}

func (p *parser) parseTaggedTemplateLiteral(tag ast.Expression) *ast.TemplateLiteral {
	l := p.parseTemplateLiteral(true)
	l.Tag = tag
	return l
}

func (p *parser) parseArgumentList() (argumentList []ast.Expression, idx0, idx1 file.Idx) {
	idx0 = p.expect(token.LEFT_PARENTHESIS)
	for p.token != token.RIGHT_PARENTHESIS {
		var item ast.Expression
		if p.token == token.ELLIPSIS {
			p.next()
			item = &ast.SpreadElement{Expression: p.parseAssignmentExpression()}
		} else {
			item = p.parseAssignmentExpression()
		}
		argumentList = append(argumentList, item)
		if p.token != token.COMMA {
			break
		}
		p.next()
	}
	idx1 = p.expect(token.RIGHT_PARENTHESIS)
	return
}

func (p *parser) parseCallExpression(left ast.Expression) ast.Expression {
	argumentList, idx0, idx1 := p.parseArgumentList()
	return &ast.CallExpression{Callee: left, LeftParenthesis: idx0, ArgumentList: argumentList, RightParenthesis: idx1}
}

func (p *parser) parseDotMember(left ast.Expression) ast.Expression {
	period := p.idx
	p.next()
	literal := p.parsedLiteral
	idx := p.idx
	if p.token == token.PRIVATE_IDENTIFIER {
		p.next()
		return &ast.PrivateDotExpression{Left: left, Identifier: ast.PrivateIdentifier{Identifier: ast.Identifier{Idx: idx, Name: literal}}}
	}
	if !token.IsId(p.token) {
		p.expect(token.IDENTIFIER)
		p.nextStatement()
		return &ast.BadExpression{From: period, To: p.idx}
	}
	p.next()
	return &ast.DotExpression{Left: left, Identifier: ast.Identifier{Idx: idx, Name: literal}}
}

func (p *parser) parseBracketMember(left ast.Expression) ast.Expression {
	idx0 := p.expect(token.LEFT_BRACKET)
	member := p.parseExpression()
	idx1 := p.expect(token.RIGHT_BRACKET)
	return &ast.BracketExpression{LeftBracket: idx0, Left: left, Member: member, RightBracket: idx1}
}

func (p *parser) parseNewExpression() ast.Expression {
	idx := p.expect(token.NEW)
	if p.token == token.PERIOD {
		p.next()
		if p.literal == "target" {
			return &ast.MetaProperty{Meta: &ast.Identifier{Name: unistring.String(token.NEW.String()), Idx: idx}, Property: p.parseIdentifier()}
		}
		_ = p.errorUnexpectedToken(token.IDENTIFIER)
	}
	callee := p.parseLeftHandSideExpression()
	if bad, ok := callee.(*ast.BadExpression); ok {
		bad.From = idx
		return bad
	}
	node := &ast.NewExpression{New: idx, Callee: callee}
	if p.token == token.LEFT_PARENTHESIS {
		argumentList, idx0, idx1 := p.parseArgumentList()
		node.ArgumentList = argumentList
		node.LeftParenthesis = idx0
		node.RightParenthesis = idx1
	}
	return node
}

func (p *parser) parseLeftHandSideExpression() ast.Expression {
	var left ast.Expression
	if p.token == token.NEW {
		left = p.parseNewExpression()
	} else {
		left = p.parsePrimaryExpression()
	}
L:
	for {
		switch p.token {
		case token.PERIOD:
			left = p.parseDotMember(left)
		case token.LEFT_BRACKET:
			left = p.parseBracketMember(left)
		case token.BACKTICK:
			left = p.parseTaggedTemplateLiteral(left)
		default:
			break L
		}
	}
	return left
}

func (p *parser) parseLeftHandSideExpressionAllowCall() ast.Expression {
	allowIn := p.scope.allowIn
	p.scope.allowIn = true
	defer func() {
		p.scope.allowIn = allowIn
	}()
	var left ast.Expression
	start := p.idx
	if p.token == token.NEW {
		left = p.parseNewExpression()
	} else {
		left = p.parsePrimaryExpression()
	}
	optionalChain := false
L:
	for {
		switch p.token {
		case token.PERIOD:
			left = p.parseDotMember(left)
		case token.LEFT_BRACKET:
			left = p.parseBracketMember(left)
		case token.LEFT_PARENTHESIS:
			left = p.parseCallExpression(left)
		case token.BACKTICK:
			if optionalChain {
				_ = p.error(p.idx, "Invalid template literal on optional chain")
				p.nextStatement()
				return &ast.BadExpression{From: start, To: p.idx}
			}
			left = p.parseTaggedTemplateLiteral(left)
		case token.QUESTION_DOT:
			optionalChain = true
			left = &ast.Optional{Expression: left}
			switch p.peek() {
			case token.LEFT_BRACKET, token.LEFT_PARENTHESIS, token.BACKTICK:
				p.next()
			default:
				left = p.parseDotMember(left)
			}
		default:
			break L
		}
	}
	if optionalChain {
		left = &ast.OptionalChain{Expression: left}
	}
	return left
}

func (p *parser) parsePostfixExpression() ast.Expression {
	operand := p.parseLeftHandSideExpressionAllowCall()
	switch p.token {
	case token.INCREMENT, token.DECREMENT:
		// 确保这里没有结束符
		if p.implicitSemicolon {
			break
		}
		tkn := p.token
		idx := p.idx
		p.next()
		switch operand.(type) {
		case *ast.Identifier, *ast.DotExpression, *ast.PrivateDotExpression, *ast.BracketExpression:
		default:
			_ = p.error(idx, "Invalid left-hand side in assignment")
			p.nextStatement()
			return &ast.BadExpression{From: idx, To: p.idx}
		}
		return &ast.UnaryExpression{Operator: tkn, Idx: idx, Operand: operand, Postfix: true}
	}
	return operand
}

func (p *parser) parseUnaryExpression() ast.Expression {
	switch p.token {
	case token.PLUS, token.MINUS, token.NOT, token.BITWISE_NOT:
		fallthrough
	case token.DELETE, token.VOID, token.TYPEOF:
		tkn := p.token
		idx := p.idx
		p.next()
		return &ast.UnaryExpression{Operator: tkn, Idx: idx, Operand: p.parseUnaryExpression()}
	case token.INCREMENT, token.DECREMENT:
		tkn := p.token
		idx := p.idx
		p.next()
		operand := p.parseUnaryExpression()
		switch operand.(type) {
		case *ast.Identifier, *ast.DotExpression, *ast.PrivateDotExpression, *ast.BracketExpression:
		default:
			_ = p.error(idx, "Invalid left-hand side in assignment")
			p.nextStatement()
			return &ast.BadExpression{From: idx, To: p.idx}
		}
		return &ast.UnaryExpression{Operator: tkn, Idx: idx, Operand: operand}
	}
	return p.parsePostfixExpression()
}

func isUpdateExpression(expr ast.Expression) bool {
	if ux, ok := expr.(*ast.UnaryExpression); ok {
		return ux.Operator == token.INCREMENT || ux.Operator == token.DECREMENT
	}
	return true
}

func (p *parser) parseExponentiationExpression() ast.Expression {
	left := p.parseUnaryExpression()
	for p.token == token.EXPONENT && isUpdateExpression(left) {
		p.next()
		left = &ast.BinaryExpression{Operator: token.EXPONENT, Left: left, Right: p.parseExponentiationExpression()}
	}
	return left
}

func (p *parser) parseMultiplicativeExpression() ast.Expression {
	left := p.parseExponentiationExpression()
	for p.token == token.MULTIPLY || p.token == token.SLASH ||
		p.token == token.REMAINDER {
		tkn := p.token
		p.next()
		left = &ast.BinaryExpression{Operator: tkn, Left: left, Right: p.parseExponentiationExpression()}
	}
	return left
}

func (p *parser) parseAdditiveExpression() ast.Expression {
	left := p.parseMultiplicativeExpression()
	for p.token == token.PLUS || p.token == token.MINUS {
		tkn := p.token
		p.next()
		left = &ast.BinaryExpression{Operator: tkn, Left: left, Right: p.parseMultiplicativeExpression()}
	}
	return left
}

func (p *parser) parseShiftExpression() ast.Expression {
	left := p.parseAdditiveExpression()
	for p.token == token.SHIFT_LEFT || p.token == token.SHIFT_RIGHT ||
		p.token == token.UNSIGNED_SHIFT_RIGHT {
		tkn := p.token
		p.next()
		left = &ast.BinaryExpression{Operator: tkn, Left: left, Right: p.parseAdditiveExpression()}
	}
	return left
}

func (p *parser) parseRelationalExpression() ast.Expression {
	if p.scope.allowIn && p.token == token.PRIVATE_IDENTIFIER {
		left := &ast.PrivateIdentifier{Identifier: ast.Identifier{Idx: p.idx, Name: p.parsedLiteral}}
		p.next()
		if p.token == token.IN {
			p.next()
			return &ast.BinaryExpression{Operator: p.token, Left: left, Right: p.parseShiftExpression()}
		}
		return left
	}
	left := p.parseShiftExpression()
	allowIn := p.scope.allowIn
	p.scope.allowIn = true
	defer func() {
		p.scope.allowIn = allowIn
	}()
	switch p.token {
	case token.LESS, token.LESS_OR_EQUAL, token.GREATER, token.GREATER_OR_EQUAL:
		tkn := p.token
		p.next()
		return &ast.BinaryExpression{Operator: tkn, Left: left, Right: p.parseRelationalExpression(), Comparison: true}
	case token.INSTANCEOF:
		tkn := p.token
		p.next()
		return &ast.BinaryExpression{Operator: tkn, Left: left, Right: p.parseRelationalExpression()}
	case token.IN:
		if !allowIn {
			return left
		}
		tkn := p.token
		p.next()
		return &ast.BinaryExpression{Operator: tkn, Left: left, Right: p.parseRelationalExpression()}
	}
	return left
}

func (p *parser) parseEqualityExpression() ast.Expression {
	left := p.parseRelationalExpression()
	for p.token == token.EQUAL || p.token == token.NOT_EQUAL ||
		p.token == token.STRICT_EQUAL || p.token == token.STRICT_NOT_EQUAL {
		tkn := p.token
		p.next()
		left = &ast.BinaryExpression{Operator: tkn, Left: left, Right: p.parseRelationalExpression(), Comparison: true}
	}
	return left
}

func (p *parser) parseBitwiseAndExpression() ast.Expression {
	left := p.parseEqualityExpression()
	for p.token == token.AND {
		tkn := p.token
		p.next()
		left = &ast.BinaryExpression{Operator: tkn, Left: left, Right: p.parseEqualityExpression()}
	}
	return left
}

func (p *parser) parseBitwiseExclusiveOrExpression() ast.Expression {
	left := p.parseBitwiseAndExpression()
	for p.token == token.EXCLUSIVE_OR {
		tkn := p.token
		p.next()
		left = &ast.BinaryExpression{Operator: tkn, Left: left, Right: p.parseBitwiseAndExpression()}
	}
	return left
}

func (p *parser) parseBitwiseOrExpression() ast.Expression {
	left := p.parseBitwiseExclusiveOrExpression()
	for p.token == token.OR {
		tkn := p.token
		p.next()
		left = &ast.BinaryExpression{Operator: tkn, Left: left, Right: p.parseBitwiseExclusiveOrExpression()}
	}
	return left
}

func (p *parser) parseLogicalAndExpression() ast.Expression {
	left := p.parseBitwiseOrExpression()
	for p.token == token.LOGICAL_AND {
		tkn := p.token
		p.next()
		left = &ast.BinaryExpression{Operator: tkn, Left: left, Right: p.parseBitwiseOrExpression()}
	}
	return left
}

func isLogicalAndExpr(expr ast.Expression) bool {
	if bexp, ok := expr.(*ast.BinaryExpression); ok && bexp.Operator == token.LOGICAL_AND {
		return true
	}
	return false
}

func (p *parser) parseLogicalOrExpression() ast.Expression {
	var idx file.Idx
	parenthesis := p.token == token.LEFT_PARENTHESIS
	left := p.parseLogicalAndExpression()
	if p.token == token.LOGICAL_OR || !parenthesis && isLogicalAndExpr(left) {
		for {
			switch p.token {
			case token.LOGICAL_OR:
				p.next()
				left = &ast.BinaryExpression{Operator: token.LOGICAL_OR, Left: left, Right: p.parseLogicalAndExpression()}
			case token.COALESCE:
				idx = p.idx
				goto mixed
			default:
				return left
			}
		}
	} else {
		for {
			switch p.token {
			case token.COALESCE:
				idx = p.idx
				p.next()
				par := p.token == token.LEFT_PARENTHESIS
				right := p.parseLogicalAndExpression()
				if !par && isLogicalAndExpr(right) {
					goto mixed
				}
				left = &ast.BinaryExpression{Operator: token.COALESCE, Left: left, Right: right}
			case token.LOGICAL_OR:
				idx = p.idx
				goto mixed
			default:
				return left
			}
		}
	}
mixed:
	_ = p.error(idx, "Logical expressions and coalesce expressions cannot be mixed. Wrap either by parentheses")
	return left
}

func (p *parser) parseConditionalExpression() ast.Expression {
	left := p.parseLogicalOrExpression()
	if p.token == token.QUESTION_MARK {
		p.next()
		consequent := p.parseAssignmentExpression()
		p.expect(token.COLON)
		return &ast.ConditionalExpression{Test: left, Consequent: consequent, Alternate: p.parseAssignmentExpression()}
	}
	return left
}

func (p *parser) parseAssignmentExpression() ast.Expression {
	start := p.idx
	parenthesis := false
	var state parserState
	if p.token == token.LEFT_PARENTHESIS {
		p.mark(&state)
		parenthesis = true
	} else {
		p.tokenToBindingId()
	}
	left := p.parseConditionalExpression()
	var operator token.Token
	switch p.token {
	case token.ASSIGN:
		operator = p.token
	case token.ADD_ASSIGN:
		operator = token.PLUS
	case token.SUBTRACT_ASSIGN:
		operator = token.MINUS
	case token.MULTIPLY_ASSIGN:
		operator = token.MULTIPLY
	case token.EXPONENT_ASSIGN:
		operator = token.EXPONENT
	case token.QUOTIENT_ASSIGN:
		operator = token.SLASH
	case token.REMAINDER_ASSIGN:
		operator = token.REMAINDER
	case token.AND_ASSIGN:
		operator = token.AND
	case token.OR_ASSIGN:
		operator = token.OR
	case token.EXCLUSIVE_OR_ASSIGN:
		operator = token.EXCLUSIVE_OR
	case token.SHIFT_LEFT_ASSIGN:
		operator = token.SHIFT_LEFT
	case token.SHIFT_RIGHT_ASSIGN:
		operator = token.SHIFT_RIGHT
	case token.UNSIGNED_SHIFT_RIGHT_ASSIGN:
		operator = token.UNSIGNED_SHIFT_RIGHT
	case token.ARROW:
		var paramList *ast.ParameterList
		if id, ok := left.(*ast.Identifier); ok {
			paramList = &ast.ParameterList{Opening: id.Idx, Closing: id.Idx1(), List: []*ast.Binding{{Target: id}}}
		} else if parenthesis {
			if seq, ok := left.(*ast.SequenceExpression); ok && len(p.errors) == 0 {
				paramList = p.reinterpretSequenceAsArrowFuncParams(seq)
			} else {
				p.restore(&state)
				paramList = p.parseFunctionParameterList()
			}
		} else {
			_ = p.error(left.Idx0(), "Malformed arrow function parameter list")
			return &ast.BadExpression{From: left.Idx0(), To: left.Idx1()}
		}
		p.expect(token.ARROW)
		node := &ast.ArrowFunctionLiteral{Start: start, ParameterList: paramList}
		node.Body, node.DeclarationList = p.parseArrowFunctionBody()
		node.Source = p.slice(node.Start, node.Body.Idx1())
		return node
	}
	if operator != 0 {
		idx := p.idx
		p.next()
		ok := false
		switch l := left.(type) {
		case *ast.Identifier, *ast.DotExpression, *ast.PrivateDotExpression, *ast.BracketExpression:
			ok = true
		case *ast.ArrayLiteral:
			if !parenthesis && operator == token.ASSIGN {
				left = p.reinterpretAsArrayAssignmentPattern(l)
				ok = true
			}
		case *ast.ObjectLiteral:
			if !parenthesis && operator == token.ASSIGN {
				left = p.reinterpretAsObjectAssignmentPattern(l)
				ok = true
			}
		}
		if ok {
			return &ast.AssignExpression{Left: left, Operator: operator, Right: p.parseAssignmentExpression()}
		}
		_ = p.error(left.Idx0(), "Invalid left-hand side in assignment")
		p.nextStatement()
		return &ast.BadExpression{From: idx, To: p.idx}
	}
	return left
}

func (p *parser) parseExpression() ast.Expression {
	left := p.parseAssignmentExpression()
	if p.token == token.COMMA {
		sequence := []ast.Expression{left}
		for {
			if p.token != token.COMMA {
				break
			}
			p.next()
			sequence = append(sequence, p.parseAssignmentExpression())
		}
		return &ast.SequenceExpression{Sequence: sequence}
	}
	return left
}

func (p *parser) checkComma(from, to file.Idx) {
	if pos := strings.IndexByte(p.str[int(from)-p.base:int(to)-p.base], ','); pos >= 0 {
		_ = p.error(from+file.Idx(pos), "Comma is not allowed here")
	}
}

func (p *parser) reinterpretAsArrayAssignmentPattern(left *ast.ArrayLiteral) ast.Expression {
	value := left.Value
	var rest ast.Expression
	for i, item := range value {
		if spread, ok := item.(*ast.SpreadElement); ok {
			if i != len(value)-1 {
				_ = p.error(item.Idx0(), "Rest element must be last element")
				return &ast.BadExpression{From: left.Idx0(), To: left.Idx1()}
			}
			p.checkComma(spread.Expression.Idx1(), left.RightBracket)
			rest = p.reinterpretAsDestructAssignTarget(spread.Expression)
			value = value[:len(value)-1]
		} else {
			value[i] = p.reinterpretAsAssignmentElement(item)
		}
	}
	return &ast.ArrayPattern{LeftBracket: left.LeftBracket, RightBracket: left.RightBracket, Elements: value, Rest: rest}
}

func (p *parser) reinterpretArrayAssignPatternAsBinding(pattern *ast.ArrayPattern) *ast.ArrayPattern {
	for i, item := range pattern.Elements {
		pattern.Elements[i] = p.reinterpretAsDestructBindingTarget(item)
	}
	if pattern.Rest != nil {
		pattern.Rest = p.reinterpretAsDestructBindingTarget(pattern.Rest)
	}
	return pattern
}

func (p *parser) reinterpretAsArrayBindingPattern(left *ast.ArrayLiteral) ast.BindingTarget {
	value := left.Value
	var rest ast.Expression
	for i, item := range value {
		if spread, ok := item.(*ast.SpreadElement); ok {
			if i != len(value)-1 {
				_ = p.error(item.Idx0(), "Rest element must be last element")
				return &ast.BadExpression{From: left.Idx0(), To: left.Idx1()}
			}
			p.checkComma(spread.Expression.Idx1(), left.RightBracket)
			rest = p.reinterpretAsDestructBindingTarget(spread.Expression)
			value = value[:len(value)-1]
		} else {
			value[i] = p.reinterpretAsBindingElement(item)
		}
	}
	return &ast.ArrayPattern{LeftBracket: left.LeftBracket, RightBracket: left.RightBracket, Elements: value, Rest: rest}
}

func (p *parser) parseArrayBindingPattern() ast.BindingTarget {
	return p.reinterpretAsArrayBindingPattern(p.parseArrayLiteral())
}

func (p *parser) parseObjectBindingPattern() ast.BindingTarget {
	return p.reinterpretAsObjectBindingPattern(p.parseObjectLiteral())
}

func (p *parser) reinterpretArrayObjectPatternAsBinding(pattern *ast.ObjectPattern) *ast.ObjectPattern {
	for _, prop := range pattern.Properties {
		if keyed, ok := prop.(*ast.PropertyKeyed); ok {
			keyed.Value = p.reinterpretAsBindingElement(keyed.Value)
		}
	}
	if pattern.Rest != nil {
		pattern.Rest = p.reinterpretAsBindingRestElement(pattern.Rest)
	}
	return pattern
}

func (p *parser) reinterpretAsObjectBindingPattern(expr *ast.ObjectLiteral) ast.BindingTarget {
	var rest ast.Expression
	value := expr.Value
	for i, prop := range value {
		ok := false
		switch pr := prop.(type) {
		case *ast.PropertyKeyed:
			if pr.Kind == ast.PropertyKindValue {
				pr.Value = p.reinterpretAsBindingElement(pr.Value)
				ok = true
			}
		case *ast.PropertyShort:
			ok = true
		case *ast.SpreadElement:
			if i != len(expr.Value)-1 {
				_ = p.error(pr.Idx0(), "Rest element must be last element")
				return &ast.BadExpression{From: expr.Idx0(), To: expr.Idx1()}
			}
			// 确保没有尾部的逗号
			rest = p.reinterpretAsBindingRestElement(pr.Expression)
			value = value[:i]
			ok = true
		}
		if !ok {
			_ = p.error(prop.Idx0(), "Invalid destructuring binding target")
			return &ast.BadExpression{From: expr.Idx0(), To: expr.Idx1()}
		}
	}
	return &ast.ObjectPattern{LeftBrace: expr.LeftBrace, RightBrace: expr.RightBrace, Properties: value, Rest: rest}
}

func (p *parser) reinterpretAsObjectAssignmentPattern(l *ast.ObjectLiteral) ast.Expression {
	var rest ast.Expression
	value := l.Value
	for i, prop := range value {
		ok := false
		switch pr := prop.(type) {
		case *ast.PropertyKeyed:
			if pr.Kind == ast.PropertyKindValue {
				pr.Value = p.reinterpretAsAssignmentElement(pr.Value)
				ok = true
			}
		case *ast.PropertyShort:
			ok = true
		case *ast.SpreadElement:
			if i != len(l.Value)-1 {
				_ = p.error(pr.Idx0(), "Rest element must be last element")
				return &ast.BadExpression{From: l.Idx0(), To: l.Idx1()}
			}
			// 确保这里没有尾部的逗号s
			rest = pr.Expression
			value = value[:i]
			ok = true
		}
		if !ok {
			_ = p.error(prop.Idx0(), "Invalid destructuring assignment target")
			return &ast.BadExpression{From: l.Idx0(), To: l.Idx1()}
		}
	}
	return &ast.ObjectPattern{LeftBrace: l.LeftBrace, RightBrace: l.RightBrace, Properties: value, Rest: rest}
}

func (p *parser) reinterpretAsAssignmentElement(expr ast.Expression) ast.Expression {
	switch ex := expr.(type) {
	case *ast.AssignExpression:
		if ex.Operator == token.ASSIGN {
			ex.Left = p.reinterpretAsDestructAssignTarget(ex.Left)
			return ex
		} else {
			_ = p.error(ex.Idx0(), "Invalid destructuring assignment target")
			return &ast.BadExpression{From: ex.Idx0(), To: ex.Idx1()}
		}
	default:
		return p.reinterpretAsDestructAssignTarget(ex)
	}
}

func (p *parser) reinterpretAsBindingElement(expr ast.Expression) ast.Expression {
	switch ex := expr.(type) {
	case *ast.AssignExpression:
		if ex.Operator == token.ASSIGN {
			ex.Left = p.reinterpretAsDestructBindingTarget(ex.Left)
			return ex
		} else {
			_ = p.error(ex.Idx0(), "Invalid destructuring assignment target")
			return &ast.BadExpression{From: ex.Idx0(), To: ex.Idx1()}
		}
	default:
		return p.reinterpretAsDestructBindingTarget(ex)
	}
}

func (p *parser) reinterpretAsBinding(expr ast.Expression) *ast.Binding {
	switch ex := expr.(type) {
	case *ast.AssignExpression:
		if ex.Operator == token.ASSIGN {
			return &ast.Binding{Target: p.reinterpretAsDestructBindingTarget(ex.Left), Initializer: ex.Right}
		} else {
			_ = p.error(ex.Idx0(), "Invalid destructuring assignment target")
			return &ast.Binding{Target: &ast.BadExpression{From: ex.Idx0(), To: ex.Idx1()}}
		}
	default:
		return &ast.Binding{Target: p.reinterpretAsDestructBindingTarget(ex)}
	}
}

func (p *parser) reinterpretAsDestructAssignTarget(item ast.Expression) ast.Expression {
	switch it := item.(type) {
	case nil:
		return nil
	case *ast.ArrayLiteral:
		return p.reinterpretAsArrayAssignmentPattern(it)
	case *ast.ObjectLiteral:
		return p.reinterpretAsObjectAssignmentPattern(it)
	case ast.Pattern, *ast.Identifier, *ast.DotExpression, *ast.PrivateDotExpression, *ast.BracketExpression:
		return it
	}
	_ = p.error(item.Idx0(), "Invalid destructuring assignment target")
	return &ast.BadExpression{From: item.Idx0(), To: item.Idx1()}
}

func (p *parser) reinterpretAsDestructBindingTarget(item ast.Expression) ast.BindingTarget {
	switch it := item.(type) {
	case nil:
		return nil
	case *ast.ArrayPattern:
		return p.reinterpretArrayAssignPatternAsBinding(it)
	case *ast.ObjectPattern:
		return p.reinterpretArrayObjectPatternAsBinding(it)
	case *ast.ArrayLiteral:
		return p.reinterpretAsArrayBindingPattern(it)
	case *ast.ObjectLiteral:
		return p.reinterpretAsObjectBindingPattern(it)
	case *ast.Identifier:
		return it
	}
	_ = p.error(item.Idx0(), "Invalid destructuring binding target")
	return &ast.BadExpression{From: item.Idx0(), To: item.Idx1()}
}

func (p *parser) reinterpretAsBindingRestElement(expr ast.Expression) ast.Expression {
	if _, ok := expr.(*ast.Identifier); ok {
		return expr
	}
	_ = p.error(expr.Idx0(), "Invalid binding rest")
	return &ast.BadExpression{From: expr.Idx0(), To: expr.Idx1()}
}
