package aiotieba

import (
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"strconv"
	"strings"
	"testing"
)

// TestLogCallArgsAreUniform 断言日志参数的两条不变量。
//
// Python 版的 handle_exception 装饰器包裹整个调用，因此每次调用只会产生一条日志，
// 且无论失败发生在哪一步，args/kwargs 都是调用方传入的实参。Go 版没有装饰器，只能在
// 方法内逐点记录，所以必须靠约定保证：
//
//  1. 同一个方法内所有 logCallError 站点的参数完全一致；
//  2. 同一个 API 的成功日志与失败日志参数完全一致。
//
// 该测试直接解析 client.go 的语法树，新增站点若破坏约定会立刻失败。
func TestLogCallArgsAreUniform(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "client.go", nil, 0)
	if err != nil {
		t.Fatalf("parse client.go: %v", err)
	}

	// 方法名 -> api 名 -> 参数签名
	errorSites := map[string]map[string]string{}
	successSites := map[string]map[string]string{}
	var sites int

	ast.Inspect(file, func(n ast.Node) bool {
		fn, ok := n.(*ast.FuncDecl)
		if !ok || fn.Recv == nil {
			return true
		}
		ast.Inspect(fn.Body, func(inner ast.Node) bool {
			call, ok := inner.(*ast.CallExpr)
			if !ok {
				return true
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			bucket := errorSites
			var rest []ast.Expr
			switch sel.Sel.Name {
			case "logCallError":
				if len(call.Args) < 2 {
					return true
				}
				// 第 2 个参数是 error 变量（err 或 reqErr 等），不参与比较。
				rest = call.Args[2:]
			case "logCallSuccess":
				bucket = successSites
				if len(call.Args) < 1 {
					return true
				}
				rest = call.Args[1:]
			default:
				return true
			}

			api, ok := stringArg(call.Args[0])
			if !ok {
				t.Errorf("%s: 日志调用的第一个参数必须是字符串字面量", fn.Name.Name)
				return true
			}
			sig := renderExprs(fset, rest)
			sites++

			if bucket[fn.Name.Name] == nil {
				bucket[fn.Name.Name] = map[string]string{}
			}
			prev, seen := bucket[fn.Name.Name][api]
			if !seen {
				bucket[fn.Name.Name][api] = sig
				return true
			}
			if prev != sig {
				t.Errorf("%s: %q 的参数不一致，同一次调用必须产生同一形状的日志\n  已有: %s\n  此处: %s",
					fn.Name.Name, api, prev, sig)
			}
			return true
		})
		return true
	})

	if sites == 0 {
		t.Fatal("没有扫描到任何日志站点，该测试会变成空断言")
	}

	// 成功日志与失败日志必须使用同一组参数。
	for method, apis := range successSites {
		for api, sucSig := range apis {
			errSig, ok := errorSites[method][api]
			if !ok {
				continue
			}
			if errSig != sucSig {
				t.Errorf("%s: %q 的成功日志与失败日志参数不一致\n  失败: %s\n  成功: %s",
					method, api, errSig, sucSig)
			}
		}
	}
}

// stringArg 提取字符串字面量的值。
func stringArg(e ast.Expr) (string, bool) {
	lit, ok := e.(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return "", false
	}
	v, err := strconv.Unquote(lit.Value)
	if err != nil {
		return "", false
	}
	return v, true
}

// renderExprs 把参数表达式渲染成可直接比较的字符串。
func renderExprs(fset *token.FileSet, exprs []ast.Expr) string {
	parts := make([]string, len(exprs))
	for i, e := range exprs {
		var b strings.Builder
		if err := format.Node(&b, fset, e); err != nil {
			parts[i] = "?"
			continue
		}
		parts[i] = b.String()
	}
	return strings.Join(parts, ", ")
}

// TestBoolResponseMethodsLogSuccess 断言「返回 exception.BoolResponse 的公开方法都会记录成功日志」。
//
// 规则来自 Python 版：handle_exception 的 ok_log_level 是逐方法显式标注的，v4.7.1 中所有
// 返回 BoolResponse 的方法都标了 ok_log_level=logging.INFO，唯二例外是 init_websocket 与
// join_chatroom。Go 版没有装饰器，成功日志必须手写在方法末尾，因此很容易漏掉——而漏掉
// 不影响功能，只会在日志里悄悄少一行，所以用测试钉住。
func TestBoolResponseMethodsLogSuccess(t *testing.T) {
	// 例外：Python 的 join_chatroom 未标注 ok_log_level，Go 版保持同样行为。
	// （init_websocket 在 Go 版不返回 BoolResponse，不需要在这里列出。）
	exceptions := map[string]bool{
		"JoinChatroom": true,
	}

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "client.go", nil, 0)
	if err != nil {
		t.Fatalf("parse client.go: %v", err)
	}

	var checked int
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Recv == nil || !fn.Name.IsExported() {
			continue
		}
		if !returnsBoolResponse(fn) {
			continue
		}
		checked++
		if exceptions[fn.Name.Name] {
			continue
		}
		if !callsLogCallSuccess(fn.Body) {
			t.Errorf("%s: 返回 exception.BoolResponse 却没有调用 logCallSuccess\n"+
				"  Python 版对应方法标注了 ok_log_level=logging.INFO，Go 版应在成功分支补上\n"+
				"  若确认不记成功日志，请加入本测试的 exceptions 并写明依据",
				fn.Name.Name)
		}
	}

	if checked == 0 {
		t.Fatal("没有扫描到返回 BoolResponse 的方法，该测试会变成空断言")
	}
}

// returnsBoolResponse 判断方法的返回值中是否含有 exception.BoolResponse。
func returnsBoolResponse(fn *ast.FuncDecl) bool {
	if fn.Type.Results == nil {
		return false
	}
	for _, result := range fn.Type.Results.List {
		sel, ok := result.Type.(*ast.SelectorExpr)
		if !ok || sel.Sel.Name != "BoolResponse" {
			continue
		}
		if pkg, ok := sel.X.(*ast.Ident); ok && pkg.Name == "exception" {
			return true
		}
	}
	return false
}

// callsLogCallSuccess 判断方法体内是否调用了 c.logCallSuccess。
func callsLogCallSuccess(body *ast.BlockStmt) bool {
	found := false
	ast.Inspect(body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		if sel.Sel.Name == "logCallSuccess" {
			found = true
			return false
		}
		return true
	})
	return found
}
