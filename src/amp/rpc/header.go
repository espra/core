// Public Domain (-) 2011 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

package rpc

import (
	"reflect"
	"unicode"
	"unicode/utf8"
)

type Error string

func (err Error) Error() string {
	return string(err)
}

var (
	ErrNotFound     error = Error("amp header: key not found")
	ErrPtrExpected  error = Error("amp header: expected pointer type")
	ErrTypeMismatch error = Error("amp header: type mismatch")
)

type Header map[string]interface{}

func (header Header) Get(key string, value interface{}) (err error) {
	if resp, ok := header[key]; ok {
		ev := reflect.ValueOf(value)
		if ev.Type().Kind() != reflect.Ptr {
			return ErrPtrExpected
		}
		ev = ev.Elem()
		et := ev.Type()
		rv := reflect.ValueOf(resp)
		rt := rv.Type()
		if et == rt {
			ev.Set(rv)
			return
		}
		ek := et.Kind()
		rk := rt.Kind()
		if ek == rk && ek == reflect.Slice && rt.Elem().Kind() == reflect.Interface {
			l := rv.Len()
			it := et.Elem()
			sl := reflect.MakeSlice(et, l, l)
			for i := 0; i < l; i++ {
				ii := rv.Index(i)
				iv := ii.Elem()
				if iv.Type() != it {
					return ErrTypeMismatch
				}
				sl.Index(i).Set(iv)
			}
			ev.Set(sl)
			return
		}
		if et.Kind() == reflect.Struct && rt.Kind() == reflect.Map && rt.Key().Kind() == reflect.String && rt.Elem().Kind() == reflect.Interface {
			return setValue(rv, ev)
		}
		return ErrTypeMismatch
	}
	return ErrNotFound
}

func (header Header) GetString(key string) (value string, ok bool) {
	if val, ok := header[key]; ok {
		if value, ok := val.(string); ok {
			return value, true
		}
		if value, ok := val.([]byte); ok {
			return string(value), true
		}
	}
	return
}

func setValue(src, dst reflect.Value) (err error) {

	ks := ""
	kv := reflect.ValueOf(&ks).Elem()

	var fv reflect.Value

	for _, mapkey := range src.MapKeys() {
		kv.Set(mapkey)
		rune, _ := utf8.DecodeRuneInString(ks)
		if !unicode.IsUpper(rune) {
			continue
		}
		field := dst.FieldByName(ks)
		if !field.IsValid() {
			continue
		}
		ft := field.Type()
		ptr := false
		if ft.Kind() == reflect.Ptr {
			if field.Elem().IsValid() {
			} else {
				fv = reflect.New(ft.Elem())
				field.Set(fv)
				ptr = true
			}
			ft = ft.Elem()
		}
		value := src.MapIndex(mapkey).Elem()
		vt := value.Type()
		if ft.Kind() == vt.Kind() {
			if ft == vt {
				if ptr {
					fv.Elem().Set(value)
				} else {
					field.Set(value)
				}
			} else {
				switch ft.Kind() {
				case reflect.Bool:
					field.SetBool(value.Internal.(bool))
				case reflect.String:
					field.SetString(value.Internal.(string))
				case reflect.Slice:
					if vt.Elem().Kind() != reflect.Interface {
						return ErrTypeMismatch
					}
					l := value.Len()
					it := ft.Elem()
					sl := reflect.MakeSlice(ft, l, l)
					for i := 0; i < l; i++ {
						ii := value.Index(i)
						iv := ii.Elem()
						if iv.Type() != it {
							return ErrTypeMismatch
						}
						sl.Index(i).Set(iv)
					}
					field.Set(sl)
				default:
					return Error("amp header: unsupported type: " + ft.Kind().String())
				}
			}
		} else if ft.Kind() == reflect.Struct && vt.Kind() == reflect.Map && vt.Key().Kind() == reflect.String && vt.Elem().Kind() == reflect.Interface {
			if ptr {
				err = setValue(value, fv.Elem())
			} else {
				err = setValue(value, field)
			}
			if err != nil {
				return err
			}
		}
	}

	return

}
