// worklink
package utl

import (
	"net/http"
	"reflect"
	"strings"

	"golang.org/x/net/html"
)

type Wlstruct struct {
	Link        string
	Udoc        interface{}
	Docaddr     *string
	ContentType string
	Savepath    string
	Resp        *http.Response
	NeedSave    bool
	Error       bool
	Lvl         uint16
}

func (wls *Wlstruct) SetLink(lnk string) {
	wls.Link = lnk
}
func (wls *Wlstruct) GetLink() string {
	return wls.Link
}
func (wls *Wlstruct) IsUdocNil() bool {
	if wls.Udoc == nil || (reflect.ValueOf(wls.Udoc).Kind() == reflect.Ptr &&
		reflect.ValueOf(wls.Udoc).IsNil()) {
		return true
	} else {
		return false
	}
}
func (wls *Wlstruct) IsUdocHTMLNode() (*html.Node, bool) {
	if doc, ok := wls.Udoc.(*html.Node); ok {
		return doc, true
	} else {
		return doc, false
	}
}
func (wls *Wlstruct) GetUdocHTMLNode() *html.Node {
	if doc, ok := wls.Udoc.(*html.Node); ok {
		return doc
	} else {
		return nil
	}
}
func (wls *Wlstruct) GetUdocPBytes() (*[]byte, error) {
	if body, ok := wls.Udoc.(*[]byte); ok {
		return body, nil
	} else {
		return nil, SPrtErr("worllenks GetUdocPBytes wls.Udoc.(*[]byte) is not ok\n")
	}
}
func (wls *Wlstruct) SetUdoc(u interface{}) {
	wls.Udoc = u
}

func (wls *Wlstruct) SetDocaddr(da *string) {
	wls.Docaddr = da
}
func (wls *Wlstruct) GetDocaddr() *string {
	return wls.Docaddr
}

func (wls *Wlstruct) SetContentType(sc string) {
	wls.ContentType = sc
}
func (wls *Wlstruct) GetContentType() string {
	return wls.ContentType
}
func (wls *Wlstruct) IsContentType() bool {
	if wls.ContentType == "" {
		return false
	} else {
		return true
	}
}
func (wls *Wlstruct) IsContentTypeHTML() bool {
	if strings.Index(wls.ContentType, "text/html;") > -1 {
		return true
	} else {
		return false
	}
}

func (wls *Wlstruct) SetSavepath(sp string) {
	wls.Savepath = sp
}
func (wls *Wlstruct) GetSavepath() string {
	return wls.Savepath
}
func (wls *Wlstruct) IsSavepath() bool {
	if wls.Savepath == "" {
		return false
	} else {
		return true
	}
}
func (wls *Wlstruct) SetResponse(resp *http.Response) {
	wls.Resp = resp
}
func (wls *Wlstruct) GetResponse() *http.Response {
	return wls.Resp
}
func (wls *Wlstruct) IsNeedSave() bool {
	return wls.NeedSave
}
func (wls *Wlstruct) SetNeedSave(f bool) {
	wls.NeedSave = f
}
func (wls *Wlstruct) IsError() bool {
	return wls.Error
}
func (wls *Wlstruct) SetError() {
	wls.Error = true
}
func (wls *Wlstruct) SetErrorResset() {
	wls.Error = false
}
func (wls *Wlstruct) GetLevel() int {
	return int(wls.Lvl)
}
func (wls *Wlstruct) SetLevel(lvl int) {
	wls.Lvl = uint16(lvl)
}
