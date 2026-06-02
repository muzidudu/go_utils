package template

import (
	"fmt"

	"github.com/flosch/pongo2/v6"
)

type filterEntry struct {
	name    string
	fn      pongo2.FilterFunction
	replace bool
}

// RegisterFilter 注册自定义 pongo2 filter，需在 Load 之前调用（Load 之后调用则立即注册到全局）。
func (e *SitesEngine) RegisterFilter(name string, fn pongo2.FilterFunction) error {
	e.filtersRegMu.Lock()
	defer e.filtersRegMu.Unlock()
	if e.filtersRegistered {
		return pongo2.RegisterFilter(name, fn)
	}
	if e.filterExistsLocked(name) {
		return fmt.Errorf("filter with name '%s' is already registered", name)
	}
	e.filterEntries = append(e.filterEntries, filterEntry{name: name, fn: fn, replace: false})
	return nil
}

// ReplaceFilter 替换已存在的 filter 实现；Load 之前可覆盖本引擎待注册的同名 filter，或覆盖 pongo2 已内置的 filter。
func (e *SitesEngine) ReplaceFilter(name string, fn pongo2.FilterFunction) error {
	e.filtersRegMu.Lock()
	defer e.filtersRegMu.Unlock()
	if e.filtersRegistered {
		return pongo2.ReplaceFilter(name, fn)
	}
	for i := range e.filterEntries {
		if e.filterEntries[i].name == name {
			e.filterEntries[i].fn = fn
			e.filterEntries[i].replace = true
			return nil
		}
	}
	if !pongo2.FilterExists(name) {
		return fmt.Errorf("filter with name '%s' does not exist (therefore cannot be overridden)", name)
	}
	e.filterEntries = append(e.filterEntries, filterEntry{name: name, fn: fn, replace: true})
	return nil
}

// FilterExists 判断 filter 是否已注册（含尚未 Load 的待注册项）。
func (e *SitesEngine) FilterExists(name string) bool {
	e.filtersRegMu.Lock()
	defer e.filtersRegMu.Unlock()
	return e.filterExistsLocked(name)
}

func (e *SitesEngine) filterExistsLocked(name string) bool {
	for _, fe := range e.filterEntries {
		if fe.name == name {
			return true
		}
	}
	return pongo2.FilterExists(name)
}

func (e *SitesEngine) registerFilters() error {
	e.filtersRegMu.Lock()
	defer e.filtersRegMu.Unlock()
	if e.filtersRegistered {
		return nil
	}
	for _, fe := range e.filterEntries {
		var err error
		switch {
		case fe.replace:
			err = pongo2.ReplaceFilter(fe.name, fe.fn)
		case pongo2.FilterExists(fe.name):
			// pongo2 filter 为进程级全局表，避免多引擎 Load 时重复注册
			continue
		default:
			err = pongo2.RegisterFilter(fe.name, fe.fn)
		}
		if err != nil {
			return fmt.Errorf("filter %q: %w", fe.name, err)
		}
	}
	e.filtersRegistered = true
	return nil
}
