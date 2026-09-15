package logging

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// PyKw 标记一个 Python 关键字参数。
//
// 传参时未被标记的值按位置参数渲染，对应 Python 的 *args；
// 被标记的值进入 kwargs，对应 Python 的 **kwargs。
//
//	c.logCallError("get_posts", err, tid, logging.PyKw{Name: "pn", Value: args.Pn})
//	// args=(123,) kwargs={'pn': 2}
type PyKw struct {
	Name  string
	Value any
}

// pyReprer 由实现了该方法的值提供自身的 Python repr。
//
// 用于让 ForumRef / UserRef 之类的引用类型在日志中输出解析后的可读值，
// 而不是 Go 结构体的 %+v 形式。
type pyReprer interface {
	PyRepr() string
}

// pyExcArgs 由库内异常实现，返回与 Python 版异常参数元组等价的参数。
type pyExcArgs interface {
	PyArgs() []any
}

// PyRepr 把 Go 值渲染为 Python repr 风格。
//
//	nil     -> None
//	true    -> True
//	"abc"   -> 'abc'
//	1       -> 1
//	[]int{} -> [1, 2]
func PyRepr(v any) string {
	if v == nil {
		return "None"
	}
	if r, ok := v.(pyReprer); ok {
		return r.PyRepr()
	}
	switch x := v.(type) {
	case string:
		return pyQuote(x)
	case []byte:
		return pyQuote(string(x))
	case error:
		// 异常出现在参数位置时按 Python 的 str(err) 渲染。
		return PyErr(x)
	}

	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Bool:
		if rv.Bool() {
			return "True"
		}
		return "False"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(rv.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(rv.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(rv.Float(), 'g', -1, rv.Type().Bits())
	case reflect.Pointer, reflect.Interface:
		if rv.IsNil() {
			return "None"
		}
		return PyRepr(rv.Elem().Interface())
	case reflect.Slice, reflect.Array:
		items := make([]string, rv.Len())
		for i := range items {
			items[i] = PyRepr(rv.Index(i).Interface())
		}
		return "[" + strings.Join(items, ", ") + "]"
	case reflect.Map:
		items := make([]string, 0, rv.Len())
		iter := rv.MapRange()
		for iter.Next() {
			items = append(items, PyRepr(iter.Key().Interface())+": "+PyRepr(iter.Value().Interface()))
		}
		return "{" + strings.Join(items, ", ") + "}"
	default:
		return fmt.Sprintf("%+v", v)
	}
}

// PyArgs 把参数渲染为 Python 风格的后缀 "args=(...) kwargs={...}"。
//
// 未被 PyKw 标记的值按位置参数渲染并保持传入顺序，PyKw 则进入关键字参数。
// 单个位置参数带有与 Python 一致的尾逗号，如 args=('盗墓笔记',)。
func PyArgs(items ...any) string {
	args := make([]string, 0, len(items))
	kwargs := make([]string, 0, len(items))
	for _, item := range items {
		kw, ok := item.(PyKw)
		if !ok {
			args = append(args, PyRepr(item))
			continue
		}
		kwargs = append(kwargs, pyQuote(kw.Name)+": "+PyRepr(kw.Value))
	}

	var b strings.Builder
	b.WriteString("args=(")
	b.WriteString(strings.Join(args, ", "))
	if len(args) == 1 {
		b.WriteByte(',')
	}
	b.WriteString(") kwargs={")
	b.WriteString(strings.Join(kwargs, ", "))
	b.WriteByte('}')
	return b.String()
}

// PyErr 把异常渲染为 Python 的 str(err)。
//
// 库内异常通过 PyArgs 提供与 Python 等价的参数元组：多参数渲染为参数元组
// （如 TiebaServerError 的 "(340011, ”)"），单参数直接渲染该参数，
// 与 CPython 的 BaseException.__str__ 一致。非库内异常退化为 err.Error()。
func PyErr(err error) string {
	if err == nil {
		return "None"
	}
	e, ok := err.(pyExcArgs)
	if !ok {
		return err.Error()
	}
	vals := e.PyArgs()
	switch len(vals) {
	case 0:
		return "()"
	case 1:
		return PyRepr(vals[0])
	default:
		parts := make([]string, len(vals))
		for i, v := range vals {
			parts[i] = PyRepr(v)
		}
		return "(" + strings.Join(parts, ", ") + ")"
	}
}

// pyQuote 复刻 Python repr 对字符串的引号选择：默认单引号，
// 内含单引号且不含双引号时改用双引号，否则转义单引号。
func pyQuote(s string) string {
	if !strings.ContainsRune(s, '\'') {
		return "'" + strings.ReplaceAll(s, `\`, `\\`) + "'"
	}
	if !strings.ContainsRune(s, '"') {
		return `"` + strings.ReplaceAll(s, `\`, `\\`) + `"`
	}
	return "'" + strings.NewReplacer(`\`, `\\`, `'`, `\'`).Replace(s) + "'"
}
