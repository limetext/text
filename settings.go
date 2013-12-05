// Copyright 2013 Fredrik Ehnbom
// Use of this source code is governed by a 2-clause
// BSD-style license that can be found in the LICENSE file.

package text

import (
	"fmt"
	"sync"
)

type (
	// An utility struct that is typically embedded in
	// other type structs to make that type implement the SettingsInterface
	HasSettings struct {
		settings Settings
	}

	// Defines an interface for types that have settings
	SettingsInterface interface {
		Settings() *Settings
	}
	OnChangeCallback func()
	settingsMap      map[string]interface{}
	Settings         struct {
		HasId
		lock              sync.Mutex
		onChangeCallbacks map[string]OnChangeCallback
		data              settingsMap
		parent            SettingsInterface
	}
)

func (s *HasSettings) Settings() *Settings {
	if s.settings.data == nil {
		s.settings = NewSettings()
	}
	return &s.settings
}

func NewSettings() Settings {
	return Settings{onChangeCallbacks: make(map[string]OnChangeCallback), data: make(settingsMap), parent: nil}
}

// Returns the parent Settings of this Settings object
func (s *Settings) Parent() SettingsInterface {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.parent
}

// Sets the parent Settings of this Settings object
func (s *Settings) SetParent(p SettingsInterface) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.parent != nil {
		old := s.parent.Settings()
		old.ClearOnChange(fmt.Sprintf("settings.child.%d", s.Id()))
	}
	s.parent = p

	if s.parent != nil {
		ns := s.parent.Settings()
		ns.AddOnChange(fmt.Sprintf("settings.child.%d", s.Id()), s.onChange)
	}
}

// Adds a OnChangeCallback identified with the given key.
// If a callback is already defined for that name, it is overwritten
func (s *Settings) AddOnChange(key string, cb OnChangeCallback) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.onChangeCallbacks[key] = cb
}

// Removes the OnChangeCallback associated with the given key.
func (s *Settings) ClearOnChange(key string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	delete(s.onChangeCallbacks, key)
}

// Get the setting identified with the given name.
// An optional default value may be specified.
// If the setting does not exist in this object,
// the parent if available will be queried.
func (s *Settings) Get(name string, def ...interface{}) interface{} {
	s.lock.Lock()
	defer s.lock.Unlock()
	if v, ok := s.data[name]; ok {
		return v
	} else if s.parent != nil {
		return s.parent.Settings().Get(name, def...)
	} else if len(def) > 0 {
		return def[0]
	}
	return nil
}

// Sets the setting identified with the given key to
// the specified value
func (s *Settings) Set(name string, val interface{}) {
	s.lock.Lock()
	s.data[name] = val
	s.lock.Unlock()
	s.onChange()
}

// Returns whether the setting identified by this key
// exists in this settings object
func (s *Settings) Has(name string) bool {
	s.lock.Lock()
	defer s.lock.Unlock()
	_, ok := s.data[name]
	return ok
}

func (s *Settings) onChange() {
	for _, v := range s.onChangeCallbacks {
		v()
	}
}

// Erases the setting associated with the given key
// from this settings object
func (s *Settings) Erase(name string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	delete(s.data, name)
}

func (s *Settings) merge(other settingsMap) {
	s.lock.Lock()
	for k, v := range other {
		s.data[k] = v
	}
	s.lock.Unlock()
	s.onChange()
}
