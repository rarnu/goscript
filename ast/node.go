package ast

import (
	"github.com/rarnu/goscript/file"
	"github.com/rarnu/goscript/token"
	"github.com/rarnu/goscript/unistring"
)

// PropertyKind 指出了属性的类型
type PropertyKind string

const (
	PropertyKindValue  PropertyKind = "value"
	PropertyKindGet    PropertyKind = "get"
	PropertyKindSet    PropertyKind = "set"
	PropertyKindMethod PropertyKind = "method"
)

// Node 节点接口，所有的节点都必须实现这个接口
type Node interface {
	Idx0() file.Idx // 属于该节点的第一个字符的索引
	Idx1() file.Idx // 紧随该节点之后的第一个字符的索引
}

// Expression 表达式接口，所有的表达式都必须实现这个接口
type Expression interface {
	Node
	_expressionNode()
}

// Statement 语句接口，所有的语句都必须实现这个接口
type Statement interface {
	Node
	_statementNode()
}

type BindingTarget interface {
	Expression
	_bindingTarget()
}

type Binding struct {
	Target      BindingTarget
	Initializer Expression
}

type Pattern interface {
	BindingTarget
	_pattern()
}

type ArrayLiteral struct {
	LeftBracket  file.Idx
	RightBracket file.Idx
	Value        []Expression
}

type ArrayPattern struct {
	LeftBracket  file.Idx
	RightBracket file.Idx
	Elements     []Expression
	Rest         Expression
}

type AssignExpression struct {
	Operator token.Token
	Left     Expression
	Right    Expression
}

type BadExpression struct {
	From file.Idx
	To   file.Idx
}

type BinaryExpression struct {
	Operator   token.Token
	Left       Expression
	Right      Expression
	Comparison bool
}

type BooleanLiteral struct {
	Idx     file.Idx
	Literal string
	Value   bool
}

type BracketExpression struct {
	Left         Expression
	Member       Expression
	LeftBracket  file.Idx
	RightBracket file.Idx
}

type CallExpression struct {
	Callee           Expression
	LeftParenthesis  file.Idx
	ArgumentList     []Expression
	RightParenthesis file.Idx
}

type ConditionalExpression struct {
	Test       Expression
	Consequent Expression
	Alternate  Expression
}

type DotExpression struct {
	Left       Expression
	Identifier Identifier
}

type PrivateDotExpression struct {
	Left       Expression
	Identifier PrivateIdentifier
}

type OptionalChain struct {
	Expression
}

type Optional struct {
	Expression
}

type FunctionLiteral struct {
	Function      file.Idx
	Name          *Identifier
	ParameterList *ParameterList
	Body          *BlockStatement
	Source        string

	DeclarationList []*VariableDeclaration
}

type ClassLiteral struct {
	Class      file.Idx
	RightBrace file.Idx
	Name       *Identifier
	SuperClass Expression
	Body       []ClassElement
	Source     string
}

type ConciseBody interface {
	Node
	_conciseBody()
}

type ExpressionBody struct {
	Expression Expression
}

type ArrowFunctionLiteral struct {
	Start           file.Idx
	ParameterList   *ParameterList
	Body            ConciseBody
	Source          string
	DeclarationList []*VariableDeclaration
}

type Identifier struct {
	Name unistring.String
	Idx  file.Idx
}

type PrivateIdentifier struct {
	Identifier
}

type NewExpression struct {
	New              file.Idx
	Callee           Expression
	LeftParenthesis  file.Idx
	ArgumentList     []Expression
	RightParenthesis file.Idx
}

type NullLiteral struct {
	Idx     file.Idx
	Literal string
}

type NumberLiteral struct {
	Idx     file.Idx
	Literal string
	Value   any
}

type ObjectLiteral struct {
	LeftBrace  file.Idx
	RightBrace file.Idx
	Value      []Property
}

type ObjectPattern struct {
	LeftBrace  file.Idx
	RightBrace file.Idx
	Properties []Property
	Rest       Expression
}

type ParameterList struct {
	Opening file.Idx
	List    []*Binding
	Rest    Expression
	Closing file.Idx
}

type Property interface {
	Expression
	_property()
}

type PropertyShort struct {
	Name        Identifier
	Initializer Expression
}

type PropertyKeyed struct {
	Key      Expression
	Kind     PropertyKind
	Value    Expression
	Computed bool
}

type SpreadElement struct {
	Expression
}

type RegExpLiteral struct {
	Idx     file.Idx
	Literal string
	Pattern string
	Flags   string
}

type SequenceExpression struct {
	Sequence []Expression
}

type StringLiteral struct {
	Idx     file.Idx
	Literal string
	Value   unistring.String
}

type TemplateElement struct {
	Idx     file.Idx
	Literal string
	Parsed  unistring.String
	Valid   bool
}

type TemplateLiteral struct {
	OpenQuote   file.Idx
	CloseQuote  file.Idx
	Tag         Expression
	Elements    []*TemplateElement
	Expressions []Expression
}

type ThisExpression struct {
	Idx file.Idx
}

type SuperExpression struct {
	Idx file.Idx
}

type UnaryExpression struct {
	Operator token.Token
	Idx      file.Idx // 如果有前缀操作符
	Operand  Expression
	Postfix  bool
}

type MetaProperty struct {
	Meta, Property *Identifier
	Idx            file.Idx
}

type BadStatement struct {
	From file.Idx
	To   file.Idx
}

type BlockStatement struct {
	LeftBrace  file.Idx
	List       []Statement
	RightBrace file.Idx
}

type BranchStatement struct {
	Idx   file.Idx
	Token token.Token
	Label *Identifier
}

type CaseStatement struct {
	Case       file.Idx
	Test       Expression
	Consequent []Statement
}

type CatchStatement struct {
	Catch     file.Idx
	Parameter BindingTarget
	Body      *BlockStatement
}

type DebuggerStatement struct {
	Debugger file.Idx
}

type DoWhileStatement struct {
	Do   file.Idx
	Test Expression
	Body Statement
}

type EmptyStatement struct {
	Semicolon file.Idx
}

type ExpressionStatement struct {
	Expression Expression
}

type ForInStatement struct {
	For    file.Idx
	Into   ForInto
	Source Expression
	Body   Statement
}

type ForOfStatement struct {
	For    file.Idx
	Into   ForInto
	Source Expression
	Body   Statement
}

type ForStatement struct {
	For         file.Idx
	Initializer ForLoopInitializer
	Update      Expression
	Test        Expression
	Body        Statement
}

type IfStatement struct {
	If         file.Idx
	Test       Expression
	Consequent Statement
	Alternate  Statement
}

type LabelledStatement struct {
	Label     *Identifier
	Colon     file.Idx
	Statement Statement
}

type ReturnStatement struct {
	Return   file.Idx
	Argument Expression
}

type SwitchStatement struct {
	Switch       file.Idx
	Discriminant Expression
	Default      int
	Body         []*CaseStatement
}

type ThrowStatement struct {
	Throw    file.Idx
	Argument Expression
}

type TryStatement struct {
	Try     file.Idx
	Body    *BlockStatement
	Catch   *CatchStatement
	Finally *BlockStatement
}

type VariableStatement struct {
	Var  file.Idx
	List []*Binding
}

type LexicalDeclaration struct {
	Idx   file.Idx
	Token token.Token
	List  []*Binding
}

type WhileStatement struct {
	While file.Idx
	Test  Expression
	Body  Statement
}

type WithStatement struct {
	With   file.Idx
	Object Expression
	Body   Statement
}

type FunctionDeclaration struct {
	Function *FunctionLiteral
}

type ClassDeclaration struct {
	Class *ClassLiteral
}

type VariableDeclaration struct {
	Var  file.Idx
	List []*Binding
}

type ClassElement interface {
	Node
	_classElement()
}

type FieldDefinition struct {
	Idx         file.Idx
	Key         Expression
	Initializer Expression
	Computed    bool
	Static      bool
}

type MethodDefinition struct {
	Idx      file.Idx
	Key      Expression
	Kind     PropertyKind // 只可以是 method/get/set
	Body     *FunctionLiteral
	Computed bool
	Static   bool
}

type ClassStaticBlock struct {
	Static          file.Idx
	Block           *BlockStatement
	Source          string
	DeclarationList []*VariableDeclaration
}

type ForLoopInitializer interface {
	_forLoopInitializer()
}

type ForLoopInitializerExpression struct {
	Expression Expression
}

type ForLoopInitializerVarDeclList struct {
	Var  file.Idx
	List []*Binding
}

type ForLoopInitializerLexicalDecl struct {
	LexicalDeclaration LexicalDeclaration
}

type ForInto interface {
	Node
	_forInto()
}

type ForIntoVar struct {
	Binding *Binding
}

type ForDeclaration struct {
	Idx     file.Idx
	IsConst bool
	Target  BindingTarget
}

type ForIntoExpression struct {
	Expression Expression
}

type Program struct {
	Body            []Statement
	DeclarationList []*VariableDeclaration
	File            *file.File
}

// 表达式节点

func (*ArrayLiteral) _expressionNode()          {}
func (*AssignExpression) _expressionNode()      {}
func (*BadExpression) _expressionNode()         {}
func (*BinaryExpression) _expressionNode()      {}
func (*BooleanLiteral) _expressionNode()        {}
func (*BracketExpression) _expressionNode()     {}
func (*CallExpression) _expressionNode()        {}
func (*ConditionalExpression) _expressionNode() {}
func (*DotExpression) _expressionNode()         {}
func (*PrivateDotExpression) _expressionNode()  {}
func (*FunctionLiteral) _expressionNode()       {}
func (*ClassLiteral) _expressionNode()          {}
func (*ArrowFunctionLiteral) _expressionNode()  {}
func (*Identifier) _expressionNode()            {}
func (*NewExpression) _expressionNode()         {}
func (*NullLiteral) _expressionNode()           {}
func (*NumberLiteral) _expressionNode()         {}
func (*ObjectLiteral) _expressionNode()         {}
func (*RegExpLiteral) _expressionNode()         {}
func (*SequenceExpression) _expressionNode()    {}
func (*StringLiteral) _expressionNode()         {}
func (*TemplateLiteral) _expressionNode()       {}
func (*ThisExpression) _expressionNode()        {}
func (*SuperExpression) _expressionNode()       {}
func (*UnaryExpression) _expressionNode()       {}
func (*MetaProperty) _expressionNode()          {}
func (*ObjectPattern) _expressionNode()         {}
func (*ArrayPattern) _expressionNode()          {}
func (*Binding) _expressionNode()               {}
func (*PropertyShort) _expressionNode()         {}
func (*PropertyKeyed) _expressionNode()         {}

// 语句节点

func (*BadStatement) _statementNode()        {}
func (*BlockStatement) _statementNode()      {}
func (*BranchStatement) _statementNode()     {}
func (*CaseStatement) _statementNode()       {}
func (*CatchStatement) _statementNode()      {}
func (*DebuggerStatement) _statementNode()   {}
func (*DoWhileStatement) _statementNode()    {}
func (*EmptyStatement) _statementNode()      {}
func (*ExpressionStatement) _statementNode() {}
func (*ForInStatement) _statementNode()      {}
func (*ForOfStatement) _statementNode()      {}
func (*ForStatement) _statementNode()        {}
func (*IfStatement) _statementNode()         {}
func (*LabelledStatement) _statementNode()   {}
func (*ReturnStatement) _statementNode()     {}
func (*SwitchStatement) _statementNode()     {}
func (*ThrowStatement) _statementNode()      {}
func (*TryStatement) _statementNode()        {}
func (*VariableStatement) _statementNode()   {}
func (*WhileStatement) _statementNode()      {}
func (*WithStatement) _statementNode()       {}
func (*LexicalDeclaration) _statementNode()  {}
func (*FunctionDeclaration) _statementNode() {}
func (*ClassDeclaration) _statementNode()    {}

func (*ForLoopInitializerExpression) _forLoopInitializer()  {}
func (*ForLoopInitializerVarDeclList) _forLoopInitializer() {}
func (*ForLoopInitializerLexicalDecl) _forLoopInitializer() {}
func (*ForIntoVar) _forInto()                               {}
func (*ForDeclaration) _forInto()                           {}
func (*ForIntoExpression) _forInto()                        {}
func (*ArrayPattern) _pattern()                             {}
func (*ArrayPattern) _bindingTarget()                       {}
func (*ObjectPattern) _pattern()                            {}
func (*ObjectPattern) _bindingTarget()                      {}
func (*BadExpression) _bindingTarget()                      {}
func (*PropertyShort) _property()                           {}
func (*PropertyKeyed) _property()                           {}
func (*SpreadElement) _property()                           {}
func (*Identifier) _bindingTarget()                         {}
func (*BlockStatement) _conciseBody()                       {}
func (*ExpressionBody) _conciseBody()                       {}
func (*FieldDefinition) _classElement()                     {}
func (*MethodDefinition) _classElement()                    {}
func (*ClassStaticBlock) _classElement()                    {}

func (i *ArrayLiteral) Idx0() file.Idx                  { return i.LeftBracket }
func (i *ArrayPattern) Idx0() file.Idx                  { return i.LeftBracket }
func (i *ObjectPattern) Idx0() file.Idx                 { return i.LeftBrace }
func (i *AssignExpression) Idx0() file.Idx              { return i.Left.Idx0() }
func (i *BadExpression) Idx0() file.Idx                 { return i.From }
func (i *BinaryExpression) Idx0() file.Idx              { return i.Left.Idx0() }
func (i *BooleanLiteral) Idx0() file.Idx                { return i.Idx }
func (i *BracketExpression) Idx0() file.Idx             { return i.Left.Idx0() }
func (i *CallExpression) Idx0() file.Idx                { return i.Callee.Idx0() }
func (i *ConditionalExpression) Idx0() file.Idx         { return i.Test.Idx0() }
func (i *DotExpression) Idx0() file.Idx                 { return i.Left.Idx0() }
func (i *PrivateDotExpression) Idx0() file.Idx          { return i.Left.Idx0() }
func (i *FunctionLiteral) Idx0() file.Idx               { return i.Function }
func (i *ClassLiteral) Idx0() file.Idx                  { return i.Class }
func (i *ArrowFunctionLiteral) Idx0() file.Idx          { return i.Start }
func (i *Identifier) Idx0() file.Idx                    { return i.Idx }
func (i *NewExpression) Idx0() file.Idx                 { return i.New }
func (i *NullLiteral) Idx0() file.Idx                   { return i.Idx }
func (i *NumberLiteral) Idx0() file.Idx                 { return i.Idx }
func (i *ObjectLiteral) Idx0() file.Idx                 { return i.LeftBrace }
func (i *RegExpLiteral) Idx0() file.Idx                 { return i.Idx }
func (i *SequenceExpression) Idx0() file.Idx            { return i.Sequence[0].Idx0() }
func (i *StringLiteral) Idx0() file.Idx                 { return i.Idx }
func (i *TemplateLiteral) Idx0() file.Idx               { return i.OpenQuote }
func (i *ThisExpression) Idx0() file.Idx                { return i.Idx }
func (i *SuperExpression) Idx0() file.Idx               { return i.Idx }
func (i *UnaryExpression) Idx0() file.Idx               { return i.Idx }
func (i *MetaProperty) Idx0() file.Idx                  { return i.Idx }
func (i *BadStatement) Idx0() file.Idx                  { return i.From }
func (i *BlockStatement) Idx0() file.Idx                { return i.LeftBrace }
func (i *BranchStatement) Idx0() file.Idx               { return i.Idx }
func (i *CaseStatement) Idx0() file.Idx                 { return i.Case }
func (i *CatchStatement) Idx0() file.Idx                { return i.Catch }
func (i *DebuggerStatement) Idx0() file.Idx             { return i.Debugger }
func (i *DoWhileStatement) Idx0() file.Idx              { return i.Do }
func (i *EmptyStatement) Idx0() file.Idx                { return i.Semicolon }
func (i *ExpressionStatement) Idx0() file.Idx           { return i.Expression.Idx0() }
func (i *ForInStatement) Idx0() file.Idx                { return i.For }
func (i *ForOfStatement) Idx0() file.Idx                { return i.For }
func (i *ForStatement) Idx0() file.Idx                  { return i.For }
func (i *IfStatement) Idx0() file.Idx                   { return i.If }
func (i *LabelledStatement) Idx0() file.Idx             { return i.Label.Idx0() }
func (i *Program) Idx0() file.Idx                       { return i.Body[0].Idx0() }
func (i *ReturnStatement) Idx0() file.Idx               { return i.Return }
func (i *SwitchStatement) Idx0() file.Idx               { return i.Switch }
func (i *ThrowStatement) Idx0() file.Idx                { return i.Throw }
func (i *TryStatement) Idx0() file.Idx                  { return i.Try }
func (i *VariableStatement) Idx0() file.Idx             { return i.Var }
func (i *WhileStatement) Idx0() file.Idx                { return i.While }
func (i *WithStatement) Idx0() file.Idx                 { return i.With }
func (i *LexicalDeclaration) Idx0() file.Idx            { return i.Idx }
func (i *FunctionDeclaration) Idx0() file.Idx           { return i.Function.Idx0() }
func (i *ClassDeclaration) Idx0() file.Idx              { return i.Class.Idx0() }
func (i *Binding) Idx0() file.Idx                       { return i.Target.Idx0() }
func (i *ForLoopInitializerVarDeclList) Idx0() file.Idx { return i.List[0].Idx0() }
func (i *PropertyShort) Idx0() file.Idx                 { return i.Name.Idx }
func (i *PropertyKeyed) Idx0() file.Idx                 { return i.Key.Idx0() }
func (i *ExpressionBody) Idx0() file.Idx                { return i.Expression.Idx0() }
func (i *FieldDefinition) Idx0() file.Idx               { return i.Idx }
func (i *MethodDefinition) Idx0() file.Idx              { return i.Idx }
func (i *ClassStaticBlock) Idx0() file.Idx              { return i.Static }
func (i *ForDeclaration) Idx0() file.Idx                { return i.Idx }
func (i *ForIntoVar) Idx0() file.Idx                    { return i.Binding.Idx0() }
func (i *ForIntoExpression) Idx0() file.Idx             { return i.Expression.Idx0() }

func (i *ArrayLiteral) Idx1() file.Idx          { return i.RightBracket + 1 }
func (i *ArrayPattern) Idx1() file.Idx          { return i.RightBracket + 1 }
func (i *AssignExpression) Idx1() file.Idx      { return i.Right.Idx1() }
func (i *BadExpression) Idx1() file.Idx         { return i.To }
func (i *BinaryExpression) Idx1() file.Idx      { return i.Right.Idx1() }
func (i *BooleanLiteral) Idx1() file.Idx        { return file.Idx(int(i.Idx) + len(i.Literal)) }
func (i *BracketExpression) Idx1() file.Idx     { return i.RightBracket + 1 }
func (i *CallExpression) Idx1() file.Idx        { return i.RightParenthesis + 1 }
func (i *ConditionalExpression) Idx1() file.Idx { return i.Test.Idx1() }
func (i *DotExpression) Idx1() file.Idx         { return i.Identifier.Idx1() }
func (i *PrivateDotExpression) Idx1() file.Idx  { return i.Identifier.Idx1() }
func (i *FunctionLiteral) Idx1() file.Idx       { return i.Body.Idx1() }
func (i *ClassLiteral) Idx1() file.Idx          { return i.RightBrace + 1 }
func (i *ArrowFunctionLiteral) Idx1() file.Idx  { return i.Body.Idx1() }
func (i *Identifier) Idx1() file.Idx            { return file.Idx(int(i.Idx) + len(i.Name)) }
func (i *NewExpression) Idx1() file.Idx {
	if i.ArgumentList != nil {
		return i.RightParenthesis + 1
	} else {
		return i.Callee.Idx1()
	}
}
func (i *NullLiteral) Idx1() file.Idx        { return file.Idx(int(i.Idx) + 4) } // "null"
func (i *NumberLiteral) Idx1() file.Idx      { return file.Idx(int(i.Idx) + len(i.Literal)) }
func (i *ObjectLiteral) Idx1() file.Idx      { return i.RightBrace + 1 }
func (i *ObjectPattern) Idx1() file.Idx      { return i.RightBrace + 1 }
func (i *RegExpLiteral) Idx1() file.Idx      { return file.Idx(int(i.Idx) + len(i.Literal)) }
func (i *SequenceExpression) Idx1() file.Idx { return i.Sequence[len(i.Sequence)-1].Idx1() }
func (i *StringLiteral) Idx1() file.Idx      { return file.Idx(int(i.Idx) + len(i.Literal)) }
func (i *TemplateLiteral) Idx1() file.Idx    { return i.CloseQuote + 1 }
func (i *ThisExpression) Idx1() file.Idx     { return i.Idx + 4 }
func (i *SuperExpression) Idx1() file.Idx    { return i.Idx + 5 }
func (i *UnaryExpression) Idx1() file.Idx {
	if i.Postfix {
		return i.Operand.Idx1() + 2 // ++ --
	}
	return i.Operand.Idx1()
}
func (i *MetaProperty) Idx1() file.Idx {
	return i.Property.Idx1()
}
func (i *BadStatement) Idx1() file.Idx        { return i.To }
func (i *BlockStatement) Idx1() file.Idx      { return i.RightBrace + 1 }
func (i *BranchStatement) Idx1() file.Idx     { return i.Idx }
func (i *CaseStatement) Idx1() file.Idx       { return i.Consequent[len(i.Consequent)-1].Idx1() }
func (i *CatchStatement) Idx1() file.Idx      { return i.Body.Idx1() }
func (i *DebuggerStatement) Idx1() file.Idx   { return i.Debugger + 8 }
func (i *DoWhileStatement) Idx1() file.Idx    { return i.Test.Idx1() }
func (i *EmptyStatement) Idx1() file.Idx      { return i.Semicolon + 1 }
func (i *ExpressionStatement) Idx1() file.Idx { return i.Expression.Idx1() }
func (i *ForInStatement) Idx1() file.Idx      { return i.Body.Idx1() }
func (i *ForOfStatement) Idx1() file.Idx      { return i.Body.Idx1() }
func (i *ForStatement) Idx1() file.Idx        { return i.Body.Idx1() }
func (i *IfStatement) Idx1() file.Idx {
	if i.Alternate != nil {
		return i.Alternate.Idx1()
	}
	return i.Consequent.Idx1()
}
func (i *LabelledStatement) Idx1() file.Idx { return i.Colon + 1 }
func (i *Program) Idx1() file.Idx           { return i.Body[len(i.Body)-1].Idx1() }
func (i *ReturnStatement) Idx1() file.Idx   { return i.Return + 6 }
func (i *SwitchStatement) Idx1() file.Idx   { return i.Body[len(i.Body)-1].Idx1() }
func (i *ThrowStatement) Idx1() file.Idx    { return i.Argument.Idx1() }
func (i *TryStatement) Idx1() file.Idx {
	if i.Finally != nil {
		return i.Finally.Idx1()
	}
	if i.Catch != nil {
		return i.Catch.Idx1()
	}
	return i.Body.Idx1()
}
func (i *VariableStatement) Idx1() file.Idx   { return i.List[len(i.List)-1].Idx1() }
func (i *WhileStatement) Idx1() file.Idx      { return i.Body.Idx1() }
func (i *WithStatement) Idx1() file.Idx       { return i.Body.Idx1() }
func (i *LexicalDeclaration) Idx1() file.Idx  { return i.List[len(i.List)-1].Idx1() }
func (i *FunctionDeclaration) Idx1() file.Idx { return i.Function.Idx1() }
func (i *ClassDeclaration) Idx1() file.Idx    { return i.Class.Idx1() }
func (i *Binding) Idx1() file.Idx {
	if i.Initializer != nil {
		return i.Initializer.Idx1()
	}
	return i.Target.Idx1()
}
func (i *ForLoopInitializerVarDeclList) Idx1() file.Idx { return i.List[len(i.List)-1].Idx1() }
func (i *PropertyShort) Idx1() file.Idx {
	if i.Initializer != nil {
		return i.Initializer.Idx1()
	}
	return i.Name.Idx1()
}
func (i *PropertyKeyed) Idx1() file.Idx  { return i.Value.Idx1() }
func (i *ExpressionBody) Idx1() file.Idx { return i.Expression.Idx1() }
func (i *FieldDefinition) Idx1() file.Idx {
	if i.Initializer != nil {
		return i.Initializer.Idx1()
	}
	return i.Key.Idx1()
}
func (i *MethodDefinition) Idx1() file.Idx {
	return i.Body.Idx1()
}
func (i *ClassStaticBlock) Idx1() file.Idx {
	return i.Block.Idx1()
}
func (i *ForDeclaration) Idx1() file.Idx    { return i.Target.Idx1() }
func (i *ForIntoVar) Idx1() file.Idx        { return i.Binding.Idx1() }
func (i *ForIntoExpression) Idx1() file.Idx { return i.Expression.Idx1() }
