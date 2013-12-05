// Copyright 2013 Fredrik Ehnbom
// Use of this source code is governed by a 2-clause
// BSD-style license that can be found in the LICENSE file.

package text

import (
	"sync"
)

var (
	idCount = Id(0)
	idMutex sync.Mutex
)

type (
	Id          int
	IdInterface interface {
		Id() Id
	}
	// An utility struct typically embedded to give
	// the type a unique id
	HasId struct {
		id Id
	}
)

func (i *HasId) Id() Id {
	if i.id == 0 {
		i.id = nextId()
	}
	return i.id
}

func nextId() Id {
	idMutex.Lock()
	defer idMutex.Unlock()
	idCount++
	return idCount
}
