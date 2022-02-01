// map.go
package utl

import (
	"sync"
)

type FileAttr struct {
	Lck         *sync.RWMutex
	ContentType string
	SavePath    string
	Saved       bool
	Error       bool
}

type tMapSavedFiles map[string]FileAttr

var SavedFiles = make(tMapSavedFiles)
var sfm sync.RWMutex

func (sf *tMapSavedFiles) GetUnsafeElement(mapkey string) (*FileAttr, error) {
	if el, ok := (*sf)[mapkey]; ok {
		return &el, nil
	} else {
		return nil, SPrtErr("map GetUnsafeElement: record with key=%q not exist\n", mapkey)
	}
}

func (sf *tMapSavedFiles) RLlock(mapkey string) error {
	sfm.Lock()
	if el, ok := (*sf)[mapkey]; ok {
		el.Lck.RLock()
		sfm.Unlock()
		return nil
	} else {
		sfm.Unlock()
		return SPrtErr("map Rlock: record with key=%q not exist\n", mapkey)
	}
}
func (sf *tMapSavedFiles) RUnlock(mapkey string) error {
	sfm.Lock()
	if el, ok := (*sf)[mapkey]; ok {
		el.Lck.RUnlock()
		sfm.Unlock()
		return nil
	} else {
		sfm.Unlock()
		return SPrtErr("map RUnlock: record with key=%q not exist\n", mapkey)
	}
}
func (sf *tMapSavedFiles) Lock(mapkey string) error {
	sfm.Lock()
	if el, ok := (*sf)[mapkey]; ok {
		el.Lck.Lock()
		sfm.Unlock()
		return nil
	} else {
		sfm.Unlock()
		return SPrtErr("map Lock: record with key=%q not exist\n", mapkey)
	}
}
func (sf *tMapSavedFiles) Unlock(mapkey string) error {
	sfm.Lock()
	if el, ok := (*sf)[mapkey]; ok {
		el.Lck.Unlock()
		sfm.Unlock()
		return nil
	} else {
		sfm.Unlock()
		return SPrtErr("map Unlock: record with key=%q not exist\n", mapkey)
	}
}
func (sf *tMapSavedFiles) Add(mapkey string, ContentType string, SavePath string, er bool) error {
	sfm.Lock()
	if _, ok := (*sf)[mapkey]; ok { //уже есть - ошибка
		sfm.Unlock()
		return SPrtErr("map: record with key=%q alredy exist\n", mapkey)
	} else {
		var locker sync.RWMutex
		locker.Lock()
		fa := FileAttr{&locker, ContentType, SavePath, false, er}
		(*sf)[mapkey] = fa
		(*sf)[mapkey].Lck.Unlock()
		sfm.Unlock()
		return nil
	}
}
func (sf *tMapSavedFiles) SetSaved(mapkey string) error {
	sfm.Lock()
	if el, ok := (*sf)[mapkey]; ok { //ok
		el.Saved = true
		sfm.Unlock()
		return nil
	} else {
		sfm.Unlock()
		return SPrtErr("map SetSaved: record with key=%q not exist\n", mapkey)
	}
}
func (sf *tMapSavedFiles) IsSaved(mapkey string) (bool, error) {
	sfm.RLock()
	if el, ok := (*sf)[mapkey]; ok { //ok
		sfm.RUnlock()
		return el.Saved, nil
	} else {
		sfm.RUnlock()
		return false, SPrtErr("map IsSaved: record with key=%q not exist\n", mapkey)
	}
}
func (sf *tMapSavedFiles) SetError(mapkey string) error {
	sfm.Lock()
	if el, ok := (*sf)[mapkey]; ok { //ok
		el.Error = true
		sfm.Unlock()
		return nil
	} else {
		sfm.Unlock()
		return SPrtErr("map SetError: record with key=%q not exist\n", mapkey)
	}
}
func (sf *tMapSavedFiles) IsError(mapkey string) bool {
	sfm.RLock()
	if el, ok := (*sf)[mapkey]; ok { //ok
		sfm.RUnlock()
		return el.Error
	} else {
		sfm.RUnlock()
		return false
	}
}
func (sf *tMapSavedFiles) IsExist(mapkey string) bool {
	sfm.RLock()
	if _, ok := (*sf)[mapkey]; ok { //ok
		sfm.RUnlock()
		return true
	} else {
		sfm.RUnlock()
		return false
	}
}
func (sf *tMapSavedFiles) GetSavePath(mapkey string) (string, error) {
	sfm.RLock()
	if el, ok := (*sf)[mapkey]; ok { //ok
		st := el.SavePath
		sfm.RUnlock()
		return st, nil
	} else {
		sfm.RUnlock()
		return "", SPrtErr("map GetSavePath: record with key=%q not exist\n", mapkey)
	}
}
func (sf *tMapSavedFiles) GetContentType(mapkey string) (string, error) {
	sfm.RLock()
	if el, ok := (*sf)[mapkey]; ok { //ok
		st := el.ContentType
		sfm.RUnlock()
		return st, nil
	} else {
		sfm.RUnlock()
		return "", SPrtErr("map GetContentType: record with key=%q not exist\n", mapkey)
	}
}
