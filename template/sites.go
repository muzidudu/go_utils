// Package template 基于 pongo2 的多站点模板引擎，支持 RegisterTag
package template

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"

	"github.com/flosch/pongo2/v6"
	core "github.com/gofiber/template/v2"
)

// SitesEngine 多站点 pongo2 模板引擎
// 目录结构: baseDir/template/{theme}/index.django, baseDir/template/{theme}/layouts/main.django
// binding 中需包含 Theme 或 siteBase.Template 以确定当前站点模板
type SitesEngine struct {
	core.Engine
	autoEscape     bool
	Templates      map[string]*pongo2.Template // key: "theme/name"
	tagParsers     []tagParser
	tagsRegistered bool
	tagsRegMu      sync.Mutex
}

type tagParser struct {
	name string
	fn   pongo2.TagParser
}

// New 创建多站点模板引擎
// baseDir: 模板根目录，如 "views"
// extension: 模板扩展名，如 ".django"
func New(baseDir, extension string) *SitesEngine {
	e := &SitesEngine{
		Templates:  make(map[string]*pongo2.Template),
		tagParsers: nil,
	}
	e.Engine.Left = "{{"
	e.Engine.Right = "}}"
	e.Engine.Directory = baseDir
	e.Engine.Extension = extension
	e.Engine.LayoutName = "embed"
	e.Engine.Funcmap = make(map[string]interface{})
	e.autoEscape = true
	return e
}

// RegisterTag 注册自定义 pongo2 tag，需在 Load 之前调用
func (e *SitesEngine) RegisterTag(name string, parser pongo2.TagParser) {
	e.tagParsers = append(e.tagParsers, tagParser{name: name, fn: parser})
}

// AddFunc 添加模板函数
func (e *SitesEngine) AddFunc(name string, fn interface{}) *SitesEngine {
	e.Engine.AddFunc(name, fn)
	return e
}

// SetAutoEscape 设置自动转义
func (e *SitesEngine) SetAutoEscape(v bool) *SitesEngine {
	e.autoEscape = v
	return e
}

// getThemeFromBinding 从 binding 中提取 theme 名称
func getThemeFromBinding(binding interface{}) string {
	if binding == nil {
		return "default"
	}
	val := reflect.ValueOf(binding)
	if val.Kind() == reflect.Ptr && !val.IsNil() {
		val = val.Elem()
	}
	if val.Kind() != reflect.Map || val.IsNil() {
		return "default"
	}
	// Theme（fiber.Map 为 map[string]interface{}，值 Kind 为 Interface 需 Elem 解包）
	if k := val.MapIndex(reflect.ValueOf("Theme")); k.IsValid() && !k.IsNil() {
		if k.Kind() == reflect.Interface {
			k = k.Elem()
		}
		if k.Kind() == reflect.String {
			s := strings.TrimSpace(k.String())
			if s != "" {
				return s
			}
		}
	}
	// siteBase.Template
	if siteBase := val.MapIndex(reflect.ValueOf("siteBase")); siteBase.IsValid() && !siteBase.IsNil() {
		sb := siteBase
		if sb.Kind() == reflect.Interface {
			sb = sb.Elem()
		}
		if sb.Kind() == reflect.Ptr && !sb.IsNil() {
			sb = sb.Elem()
		}
		if sb.Kind() == reflect.Struct {
			if f := sb.FieldByName("Template"); f.IsValid() && f.Kind() == reflect.String {
				s := strings.TrimSpace(f.String())
				if s != "" {
					return s
				}
			}
		}
	}
	return "default"
}

// Load 加载所有站点的模板
func (e *SitesEngine) Load() error {
	e.tagsRegMu.Lock()
	if !e.tagsRegistered && len(e.tagParsers) > 0 {
		for _, tp := range e.tagParsers {
			_ = pongo2.RegisterTag(tp.name, tp.fn)
		}
		e.tagsRegistered = true
	}
	e.tagsRegMu.Unlock()

	e.Mutex.Lock()
	defer e.Mutex.Unlock()

	e.Templates = make(map[string]*pongo2.Template)
	baseDir := e.Directory
	templateDir := filepath.Join(baseDir, "template")
	pongo2.SetAutoescape(e.autoEscape)

	entries, err := os.ReadDir(templateDir)
	if err != nil {
		if os.IsNotExist(err) {
			return e.loadTheme(baseDir, "default")
		}
		return fmt.Errorf("read template dir: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		theme := entry.Name()
		themePath := filepath.Join(templateDir, theme)
		if err := e.loadTheme(themePath, theme); err != nil {
			return fmt.Errorf("load theme %s: %w", theme, err)
		}
	}

	if len(e.Templates) == 0 {
		return e.loadTheme(baseDir, "default")
	}

	e.Loaded = true
	return nil
}

// loadTheme 加载单个 theme 目录下的模板
func (e *SitesEngine) loadTheme(themePath, theme string) error {
	pongoloader := pongo2.MustNewLocalFileSystemLoader(themePath)
	pongoset := pongo2.NewSet("theme:"+theme, pongoloader)
	pongoset.Globals.Update(e.Funcmap)
	walkFn := func(path string, info os.FileInfo, err error) error {
		if err != nil || info == nil || info.IsDir() {
			return err
		}
		if len(e.Extension) >= len(path) || path[len(path)-len(e.Extension):] != e.Extension {
			return nil
		}
		rel, err := filepath.Rel(themePath, path)
		if err != nil {
			return err
		}
		name := filepath.ToSlash(rel)
		name = strings.TrimSuffix(name, e.Extension)
		key := theme + "/" + name

		buf, err := core.ReadFile(path, e.FileSystem)
		if err != nil {
			return err
		}
		tmpl, err := pongoset.FromBytes(buf)
		if err != nil {
			return err
		}
		e.Templates[key] = tmpl
		if e.Verbose {
			log.Printf("template: parsed %s", key)
		}
		return nil
	}

	if err := filepath.Walk(themePath, walkFn); err != nil {
		return err
	}
	e.Loaded = true
	return nil
}

// Render 渲染模板，根据 binding 中的 Theme/siteBase 选择站点模板
func (e *SitesEngine) Render(out io.Writer, name string, binding interface{}, layout ...string) error {
	if e.PreRenderCheck() {
		if err := e.Load(); err != nil {
			return err
		}
	}

	theme := getThemeFromBinding(binding)
	nameKey := theme + "/" + name
	var layoutKey string
	if len(layout) > 0 && layout[0] != "" {
		layoutKey = theme + "/" + layout[0]
	}
	e.Mutex.RLock()
	tmpl, ok := e.Templates[nameKey]
	e.Mutex.RUnlock()

	if !ok {
		if theme != "default" {
			nameKey = "default/" + name
			e.Mutex.RLock()
			tmpl, ok = e.Templates[nameKey]
			e.Mutex.RUnlock()
			if ok && layoutKey != "" {
				layoutKey = "default/" + layout[0]
			}
		}
	}
	if !ok {
		return fmt.Errorf("template %s (theme=%s) does not exist", name, theme)
	}

	bind := getPongoBinding(binding)
	parsed, err := tmpl.Execute(bind)
	if err != nil {
		return err
	}

	if layoutKey != "" {
		e.Mutex.RLock()
		lay, ok := e.Templates[layoutKey]
		e.Mutex.RUnlock()
		if !ok {
			return fmt.Errorf("layout %s (theme=%s) does not exist", layout[0], theme)
		}
		if bind == nil {
			bind = make(pongo2.Context, 1)
		}
		bind[e.LayoutName] = pongo2.AsSafeValue(parsed)
		return lay.ExecuteWriter(bind, out)
	}
	_, err = out.Write([]byte(parsed))
	return err
}

func getPongoBinding(binding interface{}) pongo2.Context {
	if binding == nil {
		return nil
	}
	switch binds := binding.(type) {
	case pongo2.Context:
		return sanitizePongoContext(binds)
	case map[string]interface{}:
		return sanitizePongoContext(binds)
	}
	val := reflect.ValueOf(binding)
	if val.Kind() == reflect.Ptr && !val.IsNil() {
		val = val.Elem()
	}
	if val.Kind() != reflect.Map || val.IsNil() {
		return nil
	}
	if val.Type().Key().Kind() != reflect.String {
		return nil
	}
	bind := make(pongo2.Context, val.Len())
	for _, key := range val.MapKeys() {
		strKey := key.String()
		if isValidKey(strKey) {
			bind[strKey] = val.MapIndex(key).Interface()
		}
	}
	return bind
}

func sanitizePongoContext(data map[string]interface{}) pongo2.Context {
	if len(data) == 0 {
		return make(pongo2.Context)
	}
	bind := make(pongo2.Context, len(data))
	for key, value := range data {
		if isValidKey(key) {
			bind[key] = value
		}
	}
	return bind
}

func isValidKey(key string) bool {
	for _, ch := range key {
		if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_') {
			return false
		}
	}
	return true
}
