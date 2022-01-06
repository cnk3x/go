package htmlquery

import (
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	errors "golang.org/x/xerrors"
)

var (
	timeType     = reflect.TypeOf(time.Time{})
	durationType = reflect.TypeOf(time.Duration(0))
	bytesType    = reflect.TypeOf([]byte{})
)

var (
	ErrValueCannotAddress  = errors.New("value can not address")
	ErrValueNotBasicKind   = errors.New("value not a basic kind")
	ErrValueCast           = errors.New("value string cast error")
	ErrValueKindNotSupport = errors.New("value kind not support")
)

type Options struct {
	FindTag   string // "find"
	FormatTag string // "format"
}

func BindSelection(sel *goquery.Selection, out any, options ...Options) error {
	var opt Options
	if len(options) > 0 {
		opt = options[0]
	}
	return bindSelection(reflect.ValueOf(out), sel, "", "", opt)
}

func bindSelection(rv reflect.Value, sel *goquery.Selection, attr, format string, options Options) error {
	if rv = reflect.Indirect(rv); !rv.CanAddr() {
		return ErrValueCannotAddress
	}
	rt := rv.Type()

	if isBasicType(rt) {
		err := setBasicValue(rv, rt, getSelectionText(sel, attr), format)
		if err != nil {
			return errors.Errorf("set basic type: %s: %w", rt, err)
		}
		return nil
	}

	kind := rt.Kind()
	switch kind {
	case reflect.Struct:
		for i := 0; i < rt.NumField(); i++ {
			fs := rt.Field(i)
			if fs.Anonymous {
				if err := bindSelection(rv.Field(i), sel, attr, format, options); err != nil {
					return errors.Errorf("set anonymous struct: %w", err)
				}
			}
			find, attr, format, ok := parseOptions(fs, options)
			if !ok {
				continue
			}
			if err := bindSelection(rv.Field(i), sel.Find(find), attr, format, options); err != nil {
				return errors.Errorf("set struct: %w", err)
			}
		}
	case reflect.Slice:
		if sel.Length() == 0 {
			return nil
		}
		it := rt.Elem()
		var ip bool
		if ip = it.Kind() == reflect.Ptr; ip {
			it = it.Elem()
		}
		var err error
		sel.EachWithBreak(func(_ int, el *goquery.Selection) bool {
			iv := reflect.New(it).Elem()
			if err = bindSelection(iv, el, attr, format, options); err == nil {
				if ip {
					iv = iv.Addr()
				}
				rv.Set(reflect.Append(rv, iv))
			}
			return err == nil
		})
		if err != nil {
			return errors.Errorf("set slice: %w", err)
		}
	default:
		return errors.Errorf("%w: %s", ErrValueCannotAddress, kind)
	}
	return nil
}

func setBasicValue(fv reflect.Value, ft reflect.Type, value, format string) error {
	if fv = reflect.Indirect(fv); !fv.CanAddr() {
		return ErrValueCannotAddress
	}

	if value == "" {
		return nil
	}

	switch ft {
	case durationType:
		if d, err := time.ParseDuration(value); err != nil {
			return errors.Errorf("%w: %s", ErrValueCast, err.Error())
		} else {
			fv.SetInt(int64(d))
		}
		return nil
	case timeType:
		if format == "" {
			format = time.RFC3339
		}
		if t, err := time.Parse(format, value); err != nil {
			return errors.Errorf("%w: %s", ErrValueCast, err.Error())
		} else {
			fv.Set(reflect.ValueOf(t))
		}
	case bytesType:
		fv.Set(reflect.ValueOf([]byte(value)))
	}

	kind := fv.Kind()

	switch kind {
	case reflect.String:
		fv.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		x, err := strconv.ParseInt(value, 0, 0)
		if err != nil {
			return errors.Errorf("%w: %s", ErrValueCast, err.Error())
		}
		fv.SetInt(x)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		x, err := strconv.ParseUint(value, 0, 0)
		if err != nil {
			return errors.Errorf("%w: %s", ErrValueCast, err.Error())
		}
		fv.SetUint(x)
	case reflect.Float32, reflect.Float64:
		x, err := strconv.ParseFloat(value, 0)
		if err != nil {
			return errors.Errorf("%w: %s", ErrValueCast, err.Error())
		}
		fv.SetFloat(x)
	case reflect.Bool:
		x, err := strconv.ParseBool(value)
		if err != nil {
			return errors.Errorf("%w: %s", ErrValueCast, err.Error())
		}
		fv.SetBool(x)
	default:
		return errors.Errorf("%w: %s", ErrValueNotBasicKind, kind)
	}
	return nil
}

func isBasicType(ft reflect.Type) bool {
	switch ft {
	case durationType, timeType, bytesType:
		return true
	default:
		return isBasicKind(ft.Kind())
	}
}

func isBasicKind(kind reflect.Kind) bool {
	switch kind {
	case reflect.String:
	case reflect.Int64:
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
	case reflect.Float32, reflect.Float64:
	case reflect.Bool:
	default:
		return false
	}
	return true
}

func getSelectionText(sel *goquery.Selection, attr string) string {
	sel = sel.First()
	if attr == "text" || attr == "" {
		return sel.Text()
	} else {
		return sel.AttrOr(attr, "")
	}
}

func parseOptions(fs reflect.StructField, options Options) (find, attr, format string, ok bool) {
	if options.FindTag == "" {
		options.FindTag = "find"
	}
	if options.FormatTag == "" {
		options.FormatTag = "format"
	}

	if find = fs.Tag.Get(options.FindTag); find != "" {
		ok = true
		if idx := strings.LastIndex(find, ","); idx >= 0 {
			attr = strings.TrimSpace(find[idx+1:])
			find = strings.TrimSpace(find[:idx])
		}
		format = fs.Tag.Get(options.FormatTag)
	}
	return
}
