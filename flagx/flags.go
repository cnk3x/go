package flagx

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

// Parse 解析参数到struct
func Parse(strcutVal any, args []string, options ...func(flagSet *flag.FlagSet)) ([]string, error) {
	flagSet := flag.NewFlagSet("", flag.ContinueOnError)
	flag.ErrHelp = errors.New("")
	for _, apply := range options {
		apply(flagSet)
	}

	items := Struct(strcutVal)
	var kl int

	for _, item := range items {
		flagSet.Var(item, item.Name, item.Usage)
		for _, name := range item.Aliases {
			flagSet.Var(item, name, item.Usage)
		}

		pl := len(item.Name) + 2
		for _, name := range item.Aliases {
			pl += len(name) + 3
		}
		if !item.IsBoolFlag() {
			pl += len(item.KindString()) + 1
		}
		if pl > kl {
			kl = pl
		}
	}

	flagSet.Usage = func() {
		fmt.Fprintf(flagSet.Output(), "命令格式: %s [...参数]\n", flagSet.Name())
		fmt.Fprintf(flagSet.Output(), "参数说明:\n")

		for _, item := range items {
			if item.Hidden {
				continue
			}
			n := item.Name
			alias := strings.Join(item.Aliases, ", -")
			if len(item.Aliases) > 0 {
				n += ", -" + alias
			}
			if !item.IsBoolFlag() {
				n += " " + item.KindString()
			}

			envVar := strings.Join(item.EnvVars, ", $")
			if len(item.EnvVars) > 0 {
				envVar = " [$" + envVar + "]"
			}

			defs := item.Default
			if defs != "" {
				defs = " (默认 " + defs + ")"
			}

			fmt.Fprintf(flagSet.Output(), "  %*s\t%s%s%s\n", -kl, "--"+n, item.Usage, defs, envVar)
		}
		fmt.Fprintln(flagSet.Output())
	}

	err := flagSet.Parse(args)
	return flagSet.Args(), err
}

// 将struct转换为参数标记集合
func Struct(strcutVal any) (flags []Flag) {
	Walk(strcutVal, func(item Flag) { flags = append(flags, item) })
	return
}

func Walk(strcutVal any, walkFn func(item Flag)) {
	walk(reflect.Indirect(reflect.ValueOf(strcutVal)), walkFn)
}

func walk(rv reflect.Value, walkFn func(item Flag)) {
	rv = reflect.Indirect(rv)
	rt := rv.Type()

	if rt.Kind() != reflect.Struct {
		return
	}

	for i := 0; i < rt.NumField(); i++ {
		ft := rt.Field(i)
		if ft.Type.Kind() == reflect.Struct && ft.Anonymous {
			walk(rv.Field(i), walkFn)
			continue
		}
		if item, ok := getFlag(ft); ok {
			item.value = value{rv.Field(i)}
			if ev := getEnv(item.EnvVars...); ev != "" {
				item.Set(ev)
				item.Default = ev
			} else if item.Default != "" {
				item.Set(item.Default)
			} else if !item.IsZero() {
				item.Default = item.String()
			}
			walkFn(item)
		}
	}
}

type Flag struct {
	value                // 值
	Name        string   // 名称
	Aliases     []string // 别名
	Usage       string   // 用法说明
	EnvVars     []string // 关联环境变量
	Default     string   // 默认值字符串
	Deprecated  string   // 过期描述
	NoOptDefVal string   // 不带参数
	Required    bool     // 必须
	Hidden      bool     // 隐藏参数
}

func getFlag(ft reflect.StructField) (item Flag, ok bool) {
	if !ft.IsExported() {
		return
	}

	item.Name = ft.Tag.Get("flag")
	ok = item.Name != "-"
	if !ok {
		return
	}

	if item.Name != "" {
		if names := strings.Split(item.Name, ","); len(names) > 1 {
			item.Name = strings.TrimSpace(names[0])
			for i := 1; i < len(names); i++ {
				switch strings.ToLower(strings.TrimSpace(names[i])) {
				case "hidden":
					item.Hidden = true
				case "required":
					item.Required = true
				}
			}
		}
	}

	if item.Name == "" {
		item.Name = strings.ToLower(ft.Name)
	}

	item.Aliases = getTags(string(ft.Tag), "alias", "short")
	item.Usage = getTag(string(ft.Tag), "usage", "description")
	item.EnvVars = getTags(string(ft.Tag), "env")
	item.Default = getTag(string(ft.Tag), "default")
	item.Deprecated = getTag(string(ft.Tag), "deprecated")
	if ft.Type.Kind() == reflect.Bool {
		item.NoOptDefVal = "true"
	}
	return
}

type value struct {
	rv reflect.Value
}

func (v value) SetValue(x reflect.Value) {
	v.rv.Set(x)
}

func (v value) Set(s string) error {
	return setValue(v.rv, s)
}

func (v value) GetSlice() []string {
	return getValue(v.rv)
}

func (v value) String() string {
	return strings.Join(v.GetSlice(), ", ")
}

func (v value) Get() any {
	return reflect.Indirect(v.rv).Interface()
}

func (v value) IsBoolFlag() bool { return reflect.Indirect(v.rv).Kind() == reflect.Bool }

func (v value) KindString() string {
	return reflect.Indirect(v.rv).Kind().String()
}

func (v value) IsZero() bool {
	return reflect.Indirect(v.rv).IsZero()
}

var (
	_ flag.Value  = Flag{}
	_ flag.Getter = Flag{}
)

func setValue(rv reflect.Value, s string) error {
	switch rv.Kind() {
	case reflect.Ptr:
		nv := reflect.New(rv.Type().Elem())
		if err := setValue(nv.Elem(), s); err != nil {
			return err
		}
		rv.Set(nv)
	case reflect.String:
		rv.SetString(s)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		x, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return err
		}
		rv.SetInt(x)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		x, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return err
		}
		rv.SetUint(x)
	case reflect.Float32, reflect.Float64:
		x, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return err
		}
		rv.SetFloat(x)
	case reflect.Bool:
		x, err := strconv.ParseBool(s)
		if err != nil {
			return err
		}
		rv.SetBool(x)
	case reflect.Slice, reflect.Array:
		iv := reflect.New(rv.Type().Elem()).Elem()
		if err := setValue(iv, s); err != nil {
			return err
		}
		rv.Set(reflect.Append(rv, iv))
	default:
		return fmt.Errorf("unsupported kind: %s", rv.Kind())
	}
	return nil
}

func getValue(rv reflect.Value) (arr []string) {
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr,
		reflect.UnsafePointer, reflect.Interface, reflect.Slice:
		if rv.IsNil() {
			return
		}
	}

	switch rv.Kind() {
	case reflect.Ptr:
		if rv.IsZero() {
			return
		}
		return getValue(rv.Elem())
	case reflect.String:
		return []string{rv.String()}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return []string{strconv.FormatInt(rv.Int(), 10)}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return []string{strconv.FormatUint(rv.Uint(), 10)}
	case reflect.Float32, reflect.Float64:
		return []string{strconv.FormatFloat(rv.Float(), 'f', 2, 64)}
	case reflect.Bool:
		return []string{strconv.FormatBool(rv.Bool())}
	case reflect.Slice, reflect.Array:
		for i := 0; i < rv.Len(); i++ {
			arr = append(arr, getValue(rv.Index(i))...)
		}
		return
	default:
		return
	}
}

func getTags(s string, tags ...string) (r []string) {
	st := reflect.StructTag(s)
	for _, tag := range tags {
		for _, v := range strings.Split(st.Get(tag), ",") {
			if v = strings.TrimSpace(v); v != "" {
				r = append(r, v)
			}
		}
	}
	return
}

func getTag(s string, tags ...string) string {
	st := reflect.StructTag(s)
	for _, tag := range tags {
		if vs := strings.Split(st.Get(tag), ","); len(vs) > 0 {
			if v := strings.TrimSpace(vs[0]); v != "" {
				return v
			}
		}
	}
	return ""
}

func getEnv(keys ...string) string {
	for _, key := range keys {
		if v := os.Getenv(key); v != "" {
			return v
		}
	}
	return ""
}
