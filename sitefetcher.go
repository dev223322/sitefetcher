// sitefetcher
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/debug"
	"runtime/pprof"
	"strings"

	"github.com/dev223322/sitefetcher/utl"
)

func crawl() {
	var cnt int
	for { //main loop
		wls := <-utl.ChLinks
		utl.Prtf("link=%q    level=%d\n", wls.Link, wls.Lvl)
		pwls, list, err := utl.Extract(&wls)
		if err != nil {
			utl.Prtf("error: crawl Extract ignore bad link=%q err=%q\n", wls.GetLink(), err)
			continue
		}
		if pwls == nil {
			utl.Prtf("info: crawl link=%q  already was saved", wls.GetLink())
			continue
		}
		if list == nil { // nil array links
			err := utl.Savedoc(*pwls, pwls.GetSavepath(), pwls.GetContentType())
			if err != nil {
				utl.Prtf("error: not saved link=%q err=%q\n", wls.GetLink(), err)
			}
			continue
		}
		// temp ----------------------------------
		cnt++

		if cnt == 200 {
			memprofile := "/home/sh/tmp1/memprofile"
			f, _ := os.Create(memprofile)
			runtime.GC()
			pprof.WriteHeapProfile(f)
			f.Close()
		}
		if cnt%100 == 0 {
			debug.FreeOSMemory()
		}
		utl.Prtf("******* ckl ****** i=%d\n", cnt)
		// temp ------------------------------

		err = utl.Savedoc(*pwls, pwls.GetSavepath(), pwls.GetContentType())
		if err != nil {
			utl.Prtf("error: not saved2 link=%q err=%q\n", wls.GetLink(), err)
		}
		level := wls.Lvl + 1
		var wlm = make([]utl.Wlstruct, 1)

		for _, el := range list {
			if !el.IsError() {
				if el.IsNeedSave() {
					if el.IsContentTypeHTML() {
						var wlel = new(utl.Wlstruct)
						wlel.SetLink(el.GetLink())
						wlel.SetLevel(int(level))
						wlel.SetUdoc(el.GetUdocHTMLNode())
						if wlel.IsUdocNil() {
							log.Fatalf("Internal Error: crawl range list IsUdocNil  is nil ")
						}
						wlel.SetContentType(el.GetContentType())
						wlel.SetSavepath(el.GetSavepath())
						wlel.SetResponse(el.GetResponse())
						wlel.SetNeedSave(el.IsNeedSave())
						wlm = append(wlm, *wlel)
						wlel = nil
					} else { // to save
						err = utl.Savedoc(el, el.GetSavepath(), el.GetContentType())
						if err != nil {
							utl.Prtf("error: not saved3 link=%q err=%q\n", wls.GetLink(), err)
						}
					}
				}
			}
		}
		list = nil
		pwls.SetDocaddr(nil)
		pwls.SetResponse(nil)
		pwls.SetUdoc(nil)
		pwls = nil

		utl.ChALinks <- wlm

		//----------------------------------------------------
	} // end of main loop
}

func main() {
	debug.SetGCPercent(3)
	go utl.PrintQueue() // start print queue

	var n int           // number of pending sends to worklist
	var maxlevel uint16 // max level of links from url
	var wklvl int
	var t int
	flag.IntVar(&wklvl, "level", 1, "max level of links from url")
	flag.IntVar(&t, "t", 10, "number of threads")
	flag.StringVar(&utl.Path, "out", "/home/sh/tmp", "path to catalog for create files")
	flag.Parse()
	maxlevel = uint16(wklvl)
	// Start with the command-line arguments.
	n++
	startlink := flag.Arg(0)
	if startlink == "" || t > 10 {
		fmt.Println("pgm [-level=NN] [-out=PATH] [-t=1..10]  <url>")

		return
	}
	var err error
	utl.Domain, err = getdmn(startlink)
	if err != nil {
		log.Fatalf("Error: err=%q", err)
	}
	go func() { utl.ChALinks <- newwlm(startlink, 1) }()
	for i := 0; i < t; i++ {
		go func() { crawl() }()
	}
	// Crawl  concurrently.
	seen := make(map[string]bool)
	for ; n > 0; n-- {
		list := <-utl.ChALinks
		for _, link := range list {
			if link.Lvl < maxlevel && !seen[link.GetLink()] && utl.Mydomain(link.Link, utl.Domain) {
				seen[link.Link] = true
				n++
				utl.ChLinks <- link
			}
		}
		for _, link := range list {
			link.Resp = nil
			link.Udoc = nil
		}
		list = nil
	}

}

func getdmn(url string) (string, error) {
	if len(url) == 0 {
		return "", utl.SPrtErr("getdmn - url=%q is not url")
	}
	b := strings.Index(url, "//")
	if b == -1 {
		return "", utl.SPrtErr("getdmn2 - url=%q is not url")
	}
	b += 2
	dmn := url[b:]
	e := strings.Index(dmn, "/")
	if e > 0 {
		dmn = dmn[0:e]
	}
	return dmn, nil
}

func newwlm(link string, level int) []utl.Wlstruct {
	wlm := make([]utl.Wlstruct, 1)
	wlm[0].SetLink(link)
	wlm[0].SetLevel(level)
	wlm[0].SetNeedSave(true)

	return wlm
}
