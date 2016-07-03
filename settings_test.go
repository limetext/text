// Copyright 2013 Fredrik Ehnbom
// Use of this source code is governed by a 2-clause
// BSD-style license that can be found in the LICENSE file.

package text

import (
	"fmt"
	"testing"
)

func TestSettings(t *testing.T) {
	var (
		s1, s2 HasSettings
		called bool
	)
	s1.Settings().SetParent(&s2)

	if v, ok := s1.Settings().Get("test", true).(bool); !ok || !v {
		t.Error(ok, v)
	}
	s2.Settings().Set("test", false)
	if v, ok := s1.Settings().Get("test", true).(bool); !ok || v {
		t.Error(ok, v)
	}

	s1.Settings().AddOnChange("something", func(name string) {
		called = true
	})
	s2.Settings().Set("test", true)
	if !called {
		t.Error("Should have been called..")
	}
	called = false
	s1.Settings().ClearOnChange("something")
	s2.Settings().Set("test", true)
	if called {
		t.Error("Should not have been called..")
	}
}

func TestCallbacksOnUnmarshal(t *testing.T) {
	tests := []struct {
		before string
		after  string
		cbs    []string
		exp    map[string]bool
	}{
		{
			`{"font_size": 14}`,
			`{"font_size": 12}`,
			[]string{"font_size"},
			map[string]bool{"font_size": true},
		},
		{
			`{"font_size": 14}`,
			`{"font_size": 14}`,
			[]string{"font_size"},
			map[string]bool{},
		},
		{
			`{"a": "t1", "b": 1, "c": true}`,
			`{"a": "t2", "b": 1, "c": false}`,
			[]string{"a", "c"},
			map[string]bool{"a": true, "c": true},
		},
		{
			`{"font_size": 14}`,
			`{"font_size": 12}`,
			[]string{},
			map[string]bool{},
		},
		{
			`{"a": "t1", "b": 1}`,
			`{"a": "t1", "c": false}`,
			[]string{"b", "c"},
			map[string]bool{"b": true, "c": true},
		},
	}
	cnt := make(map[string]bool)
	cb := func(key string) {
		cnt[key] = true
	}

	for i, test := range tests {
		cnt = make(map[string]bool)
		val := NewSettings()
		set := &val
		if err := set.UnmarshalJSON([]byte(test.before)); err != nil {
			t.Errorf("Test %d: Error on unmarshaling before data %s", i, err)
		}
		for _, k := range test.cbs {
			set.AddOnChange(fmt.Sprintf("Test.%d.%s", i, k), cb)
		}

		if err := set.UnmarshalJSON([]byte(test.after)); err != nil {
			t.Errorf("Test %d: Error on unmarshaling after data %s", i, err)
		}
		if lcnt, lexp := len(cnt), len(test.exp); lexp != lcnt {
			t.Errorf("Test %d: map and counted map length difference expected %d, but got %d\nexp: %s\ncnt: %s", i, lexp, lcnt, test.exp, cnt)
			continue
		}
		for k, _ := range test.exp {
			if !cnt[k] {
				t.Errorf("Test %d: Expected %s key get called, but it didn't", i, k)
			}
		}
	}
}
