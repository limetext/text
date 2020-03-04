// Copyright 2013 Fredrik Ehnbom
// Use of this source code is governed by a 2-clause
// BSD-style license that can be found in the LICENSE file.

package text

import (
	"encoding/json"
	"fmt"
	"reflect"
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
	OnChangeCallback func(name string)
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
	return Settings{
		onChangeCallbacks: make(map[string]OnChangeCallback),
		data:              make(settingsMap),
		parent:            nil,
	}
}

// Returns the parent Settings of this Settings object
func (s *Settings) Parent() SettingsInterface {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.parent
}

func (s *Settings) UnmarshalJSON(data []byte) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	// copying settings data
	old := make(settingsMap)
	for k, v := range s.data {
		old[k] = v
	}
	// clearing settings data before unmarshalling the new data
	s.data = make(settingsMap)
	if err := json.Unmarshal(data, &s.data); err != nil {
		return err
	}
	// checking for any new, modified, deleted setting and calling callbacks
	for k, v := range old {
		if v2, ok := s.data[k]; !ok || !reflect.DeepEqual(v, v2) {
			s.lock.Unlock()
			s.onChange(k)
			s.lock.Lock()
		}
	}
	for k, _ := range s.data {
		if _, ok := old[k]; !ok {
			s.lock.Unlock()
			s.onChange(k)
			s.lock.Lock()
		}
	}
	return nil
}

func (s *Settings) MarshalJSON() (data []byte, err error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	return json.Marshal(&s.data)
}

// Sets the parent Settings of this Settings object
func (s *Settings) SetParent(p SettingsInterface) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.parent != nil {
		old := s.parent.Settings()
		old.ClearOnChange(fmt.Sprintf("settings.child.%d", s.Id()))
	}
	if s.parent = p; s.parent != nil {
		ns := s.parent.Settings()
		ns.AddOnChange(fmt.Sprintf("settings.child.%d", s.Id()), s.onChange)
	}
}

// Adds a OnChangeCallback identified with the given key.
// If a callback is already defined for that name, it is overwritten
func (s *Settings) AddOnChange(key string, cb OnChangeCallback) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.onChangeCallbacks == nil {
		s.onChangeCallbacks = make(map[string]OnChangeCallback)
	}
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

func (s *Settings) Int(name string, def ...interface{}) int {
	value := s.Get(name, def...)
	switch val := value.(type) {
	case int64:
		return int(val)
	case int:
		return val
	case uint64:
		return int(val)
	case uint32:
		return int(val)
	case uintptr:
		return int(val)
	case float32:
		return int(val)
	case float64:
		return int(val)
	}
	panic(fmt.Sprintf("value of %s cannot be represented as an int: %#v", name, value))
}

func (s *Settings) String(name string, def ...interface{}) string {
	value, ok := s.Get(name, def...).(string)
	if ok {
		return value
	}
	panic(fmt.Sprintf("value of %s cannot be represented as an string: %#v", name, value))
}

func (s *Settings) Bool(name string, def ...interface{}) bool {
	value, ok := s.Get(name, def...).(bool)
	if ok {
		return value
	}
	panic(fmt.Sprintf("value of %s cannot be represented as an bool: %#v", name, value))
}

// Sets the setting identified with the given key to
// the specified value
func (s *Settings) Set(name string, val interface{}) {
	s.lock.Lock()
	s.data[name] = val
	s.lock.Unlock()
	s.onChange(name)
}

// Returns whether the setting identified by this key
// exists in this settings object
func (s *Settings) Has(name string) bool {
	s.lock.Lock()
	defer s.lock.Unlock()
	_, ok := s.data[name]
	return ok
}

func (s *Settings) onChange(name string) {
	for _, cb := range s.onChangeCallbacks {
		cb(name)
	}
}

// Erases the setting associated with the given key
// from this settings object
func (s *Settings) Erase(name string) {
	s.lock.Lock()
	delete(s.data, name)
	s.lock.Unlock()
	s.onChange(name)
}
