// map.go
package utl

import (
	"sync"
)

type FileAttr struct {
	ContentType string
	SavePath    string
	Saved       bool
	Error       bool
}

type tMapSavedFiles map[string]FileAttr

var SavedFiles = make(tMapSavedFiles)
var sfm sync.RWMutex

func (sf *tMapSavedFiles) GetElementAndRLock(mapkey string) (*FileAttr, error) {
	sfm.RLock()
	if el, ok := (*sf)[mapkey]; ok {
		return &el, nil
	} else {
		sfm.RUnlock()
		return nil, SPrtErr("map GetUnsafeElement: record with key=%q not exist\n", mapkey)
	}
}

func (sf *tMapSavedFiles) RUnlock() {
	sfm.RUnlock()
}
func (sf *tMapSavedFiles) Lock() {
	sfm.Lock()
}
func (sf *tMapSavedFiles) Unlock() {
	sfm.Unlock()
}
func (sf *tMapSavedFiles) Add(mapkey string, ContentType string, SavePath string, er bool) error {
	sfm.Lock()
	if _, ok := (*sf)[mapkey]; ok { //уже есть - ошибка
		sfm.Unlock()
		return SPrtErr("map: record with key=%q alredy exist\n", mapkey)
	} else {
		fa := FileAttr{ContentType, SavePath, false, er}
		(*sf)[mapkey] = fa
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
		ret := el.Saved
		sfm.RUnlock()
		return ret, nil
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
		ret := el.Error
		sfm.RUnlock()
		return ret
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
