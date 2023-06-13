package appenv

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"

	"github.com/joho/godotenv"
)

var DefaultAPP_ENV = "production"

func Load(v any, dir string) error {
	return load(v, openOS, dir)
}

func LoadFS(v any, fsys fs.FS, dir string) error {
	return load(v, openFS(fsys), dir)
}

func LoadOnAPP_ENV(v any, dir string, APP_ENV string) error {
	return loadOnAPP_ENV(v, openOS, dir, APP_ENV)
}

func LoadFSOnAPP_ENV(v any, fsys fs.FS, dir string, APP_ENV string) error {
	return loadOnAPP_ENV(v, openFS(fsys), dir, APP_ENV)
}

type openFunc func(dir, base string) (fs.File, error)

func openOS(dir, base string) (fs.File, error) {
	return os.Open(filepath.Join(dir, base))
}

func openFS(fsys fs.FS) openFunc {
	return func(dir, base string) (fs.File, error) {
		return fsys.Open(path.Join(dir, base))
	}
}

func load(v any, open openFunc, dir string) error {
	appenv := os.Getenv("APP_ENV")
	if appenv == "" {
		appenv = DefaultAPP_ENV
	}
	return loadOnAPP_ENV(v, open, dir, appenv)
}

func loadOnAPP_ENV(v any, open openFunc, dir, appenv string) error {
	err := setFromFile(v, open, dir, appenv+".env")
	if err != nil {
		return err
	}

	err = setFromFile(v, open, dir, ".env")
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	return setFromEnv(v)
}

func setFromEnv(v any) error {
	return setFields(v, os.LookupEnv)
}

func setFromFile(v any, open openFunc, dir, base string) error {
	f, err := open(dir, base)
	if err != nil {
		return err
	}
	defer f.Close()

	m, err := godotenv.Parse(f)
	if err != nil {
		return err
	}

	return setFields(v, func(s string) (string, bool) {
		v, ok := m[s]
		return v, ok
	})
}

var trueRe = regexp.MustCompile("^(?i:1|true)$")

var unmarshaller = map[reflect.Kind]func(string) (any, error){
	reflect.Int: func(s string) (any, error) {
		i, err := strconv.ParseInt(s, 10, 0)
		return int(i), err
	},
	reflect.Bool: func(s string) (any, error) {
		return trueRe.MatchString(s), nil
	},
	reflect.String: func(s string) (any, error) {
		return s, nil
	},
}

type lookupFunc func(string) (string, bool)

func setFields(v any, lookup lookupFunc) error {
	rv := mustValueOfPtrStruct(v)

	for i := 0; i < rv.NumField(); i++ {
		tag, ok := structTag(rv.Type().Field(i))
		if !ok {
			continue
		}
		ev, ok := lookup(tag)
		if !ok {
			continue
		}
		if err := setField(rv.Field(i), ev); err != nil {
			return err
		}
	}
	return nil
}

func setField(v reflect.Value, str string) error {
	if !v.CanSet() {
		return nil
	}

	unmarshal, ok := unmarshaller[v.Kind()]
	if !ok {
		panic(fmt.Sprintf("set field: unsupported type: %s", v.Type()))
	}

	u, err := unmarshal(str)
	if err != nil {
		return fmt.Errorf("set field: cannot unmarshal %s to type %s", str, v.Type())
	}
	v.Set(reflect.ValueOf(u))
	return nil
}

func structTag(f reflect.StructField) (string, bool) {
	return f.Tag.Lookup("env")
}

func mustValueOfPtrStruct(v any) reflect.Value {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer {
		panic("not a pointer")
	}
	rv = rv.Elem()
	if rv.Kind() != reflect.Struct {
		panic("not a struct")
	}
	return rv
}