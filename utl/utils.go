// utl
package utl

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"golang.org/x/net/html"
)

var Domain string
var Path string

var prtch = make(chan string, 500)
var ChALinks = make(chan []Wlstruct, 5000000)
var ChLinks = make(chan Wlstruct, 50)

func SPrtErr(format string, a ...interface{}) error {
	return fmt.Errorf("Error: "+format, a...)
}

func PrintQueue() {
	for {
		st := <-prtch
		fmt.Print(st)
	}
}

func Prtf(format string, a ...interface{}) {
	st := fmt.Sprintf(format, a...)
	prtch <- st
}
func Prtln(st string) {
	st += "\n"
	prtch <- st
}

func Mydomain(url string, dmn string) bool {
	pos := strings.Index(url, dmn)
	if pos > -1 {
		return true
	} else {
		return false
	}
}

func Replurl(wkslnk *Wlstruct, pathdir string, contentType string) (string, error) {
	var retval string
	u, err := url.Parse(wkslnk.GetLink())
	if err != nil {
		return "", SPrtErr("utlils: Replurl - err=%q", err)
	}
	relpath := u.Path
	if len(relpath) == 0 {
		relpath = "/index"
	}
	relpath = strings.Replace(relpath, ".", "_", -1)
	if string(relpath[0]) != "/" {
		relpath = "/" + relpath
	}
	if wkslnk.IsContentTypeHTML() {
		if strings.Index(relpath, ".htm") > -1 {
			retval = pathdir + relpath
		} else {
			retval = pathdir + relpath + ".html"
		}
	} else {
		retval = pathdir + relpath
	}
	/*
		if strings.Index(retval, "/home/sh/tmp/progress/21241.html/amp") != -1 { //dbg
			log.Fatalln("Error: dbg")
		}
	*/
	return retval, nil
}

func GetContent(wls *Wlstruct, pathdir string) error {
	if wls.GetLink() == "" {
		wls.SetError()
		return SPrtErr("utils: GetContent - wls.Link is null\n")
	}
	if wls.IsError() {
		return SPrtErr("utils: GetContent - wls.Link=%q is error\n", wls.GetLink())
	}

	var found bool
	var resp *http.Response

	if err := SavedFiles.RLlock(wls.GetLink()); err == nil { //Rlock element of map if  record present
		found = true
	}
	var err error

	if found {
		var el *FileAttr
		if el, err = SavedFiles.GetUnsafeElement(wls.GetLink()); err != nil {
			log.Fatalf("!!! Fatal error: utils: GetContent,GetUnsafeElement- record is absent after it was- err=%q", err)
		}
		wls.SetContentType(el.ContentType)
		wls.SetSavepath(el.SavePath)

		if err := SavedFiles.RUnlock(wls.GetLink()); err != nil { //RUnlock element of map
			log.Fatalf("!!! Fatal error: utils: GetContent - record was locked, but after is absent - err=%q\n", err)
		}
		if wls.GetDocaddr() != nil {
			*(wls.GetDocaddr()) = wls.GetSavepath()
		} else {
			log.Fatal("Internal error: utils: GetContent - Docaddr is nil\n")
		}
	} else { //need get content
		resp, err = http.Get(wls.GetLink())
		if err != nil {
			if resp != nil {
				resp.Body.Close()
			}
			wls.SetError()
			return SPrtErr("utils: GetContent - http.Get - error   err=%q", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			wls.SetError()
			return SPrtErr("utils: GetContent - getting %s: %s", wls.GetLink(), resp.Status)
		}
		wls.ContentType = resp.Header.Get("Content-Type")
		Prtf("---contentType=%q  url=%q\n", wls.GetContentType(), wls.GetLink())
		savepath, err := Replurl(wls, pathdir, wls.GetContentType())
		if err != nil {
			return SPrtErr("utils: GetContent - Replurl   err=%q", err)
		}
		wls.SetSavepath(savepath)
		SavedFiles.Add(wls.GetLink(), wls.GetContentType(), wls.GetSavepath())
		wls.SetNeedSave(true)
	} //end need get content
	if wls.IsNeedSave() {
		if wls.IsContentTypeHTML() {
			doc, err := html.Parse(resp.Body)
			resp.Body.Close()
			if err != nil {
				wls.SetError()
				return SPrtErr("utils: GetContent parsing %s as HTML: err=%q\n", wls.GetLink(), err)
			}
			wls.SetUdoc(doc)
		} else {
			body, err := ioutil.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				wls.SetError()
				return SPrtErr("utils: GetContent ioutil.ReadAll err=%q\n", err)
			}
			wls.SetUdoc(&body)
		}
	}

	//--------------------------------------------
	wls.SetResponse(resp)
	wls.SetErrorResset()

	return nil
}

func Savedoc(wls Wlstruct, pathfile string, contentType string) error {
	// check conditions at the begin
	saved, err := SavedFiles.IsSaved(wls.GetLink())
	if err != nil {
		log.Fatalf("Internal error: utils - call Savedoc before SavedFiles rec created  err=%q\n", err)
	}
	if saved {
		return nil //allready saved
	}
	if !wls.IsNeedSave() {
		return nil // nothing to do
	}
	if wls.IsError() {
		SPrtErr("utils: Savedoc bad link=%q\n", wls.GetLink())
	}
	// checked - need to save

	lastndx := strings.LastIndex(pathfile, "/")
	if lastndx != -1 {
		pathdir := pathfile[:lastndx+1]
		err := os.MkdirAll(pathdir, 0750)
		if err != nil {
			log.Fatalf("!!! Fatal error: utils: error in  os.MkdirAll - err=%q\n", err)
		}
	}
	file, err := os.OpenFile(pathfile, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0750)
	if err != nil {
		return SPrtErr("utils: Savedoc os.OpenFile - unavalible to create file - %s  err=%q\n", pathfile, err)
	}
	defer file.Close()
	if wls.IsContentTypeHTML() {
		if doc, ok := wls.IsUdocHTMLNode(); ok {
			err = html.Render(file, doc)
			if err != nil {
				return SPrtErr("utils: Savedoc html.Render  err%q\n", err)
			}
			doc = nil
		} else {
			return SPrtErr("utils: Savedoc difference between IsContentTypeHTML and IsUdocHTMLNode Link=%q\n", wls.GetLink())
		}
	} else {
		if pdump, err := wls.GetUdocPBytes(); err == nil {
			if err := ioutil.WriteFile(pathfile, *pdump, 0750); err == nil {
				pdump = nil
			} else {
				return SPrtErr("utils: Savedoc ioutil.WriteFile error - pathfile=%q\n", pathfile)
			}

		} else {
			return SPrtErr("utils: Savedoc - uncnown udoc type... - pathfile=%q  link=%q\n", pathfile, wls.GetLink())
		}
	}

	//------------------------------------------------------
	if err := SavedFiles.SetSaved(wls.GetLink()); err != nil {
		log.Fatalf("!!! Fatal error: utils: Savedoc SetSaved - link=%q, err=%q", wls.GetLink(), err)
	}
	Prtf("*** SaveDoc - pathfile%q\n", pathfile)
	return nil
}
