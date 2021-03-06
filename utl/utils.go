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
	"sync"

	"golang.org/x/net/html"
)

var Domain string
var Path string

var prtch = make(chan string, 500)
var ChALinks = make(chan []Wlstruct, 5000000)
var ChLinks = make(chan Wlstruct, 50)

var Wg sync.WaitGroup

func SPrtErr(format string, a ...interface{}) error {
	return fmt.Errorf("Error: "+format, a...)
}
func SPrtInfo(format string, a ...interface{}) error {
	return fmt.Errorf("Info: "+format, a...)
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

	if string(relpath[0]) != "/" {
		relpath = "/" + relpath
	}
	var part1, part2 string

	lastslash := strings.LastIndex(relpath, "/")
	if lastslash == -1 {
		if wkslnk.IsContentTypeHTML() {
			retval = pathdir + "/index.html"
		} else {
			retval = pathdir + relpath
		}
	} else {
		part1 = relpath[0 : lastslash+1]
		part1 = strings.Replace(part1, ".", "_", -1)
		part2 = relpath[lastslash+1:]
		if wkslnk.IsContentTypeHTML() {
			lastpoint := strings.LastIndex(part2, ".")
			lasthtml := strings.LastIndex(part2, ".htm")
			l := len(part2)

			switch {
			case u.RawQuery != "":
				p1p2 := part2[:lastpoint]
				p2p2hashs := html.EscapeString(u.RawQuery)
				retval = pathdir + part1 + p1p2 + "___" + p2p2hashs + ".html"

			case l > 4 && lastpoint > -1 && lastpoint == lasthtml:
				retval = pathdir + part1 + part2
			case l == 4 && lastpoint > -1 && lastpoint == lasthtml:
				retval = pathdir + part1 + "/index.html"
			case l > 1 && lastpoint == -1 && lasthtml == -1:
				retval = pathdir + part1 + part2 + "/index.html"
			case l == 1:
				retval = pathdir + part1 + part2 + "index.html"
			case l == 0:
				retval = pathdir + part1 + "index.html"
			}
		} else {
			retval = pathdir + part1 + part2
		}
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

	var resp *http.Response

	if el, err := SavedFiles.GetElementAndRLock(wls.GetLink()); err == nil { // if found
		wls.SetContentType(el.ContentType)
		wls.SetSavepath(el.SavePath)

		SavedFiles.RUnlock()
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
			if err := SavedFiles.Add(wls.GetLink(), "", "", true); err != nil { // mark as error if already exist
				SavedFiles.SetError(wls.GetLink()) // mark as error
			}
			return SPrtErr("utils: GetContent - http.Get - error   err=%q", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			wls.SetError()
			if err := SavedFiles.Add(wls.GetLink(), "", "", true); err != nil { // mark as error if already exist
				SavedFiles.SetError(wls.GetLink()) // mark as error
			}
			return SPrtErr("utils: GetContent - getting %s: %s", wls.GetLink(), resp.Status)
		}
		wls.ContentType = resp.Header.Get("Content-Type")
		Prtf("---contentType=%q  url=%q\n", wls.GetContentType(), wls.GetLink())
		savepath, err := Replurl(wls, pathdir, wls.GetContentType())
		if err != nil {
			return SPrtErr("utils: GetContent - Replurl   err=%q", err)
		}
		wls.SetSavepath(savepath)
		if !SavedFiles.IsExist(wls.GetLink()) {
			SavedFiles.Add(wls.GetLink(), wls.GetContentType(), wls.GetSavepath(), false)
		} else {
			if SavedFiles.IsError(wls.GetLink()) {
				return SPrtInfo("utils: GetContent, IsError - bad link=%q\n", wls.GetLink())
			}
		}
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

func ChanIsFree() bool {

	select {
	case wls := <-ChLinks:
		ChLinks <- wls
		return false
	default:
		return true
	}

}
