package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type RequestPayLoad struct {
	Term    string  `json:"term"`
	Signal  Signal  `json:"signal"`
	Options Options `json:"options"`
}
type Signal struct {
}
type Options struct {
	CollectionsLimit int    `json:"collectionsLimit"`
	Types            string `json:"types"`
}

func main() {
	// Request URL:https://www.france.tv/recherche/lancer/
	// Request Method:POST
	// Remote Address:23.198.76.190:443
	// Host: www.france.tv
	// User-Agent: Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:77.0) Gecko/20100101 Firefox/77.0
	// Accept: */*
	// Accept-Language: en-US,en;q=0.5
	// Accept-Encoding: gzip, deflate, br
	// Referer: https://www.france.tv/recherche/
	// Content-Type: text/plain;charset=UTF-8
	// Origin: https://www.france.tv
	// Content-Length: 90
	// Connection: keep-alive
	// Cookie: utag_main=v_id:016cdc127b34000ba63e5f9608880104c001600900bd0$_sn:86$_ss:0$_st:1574362349698$ses_id:1574359765526%3Bexp-session$_pn:7%3Bexp-session; gig_hasGmid=ver2; gig_bootstrap_3_1UVV83m_9BIzYs1z_EzYR-OdACvYSmuOET8okPHXG_ixlVCPlUpwRWK1rvN_kINm=_gigya_ver3; gig_canary=false; gig_canary_ver=11005-5-26524275
	// TE: Trailers

	rq := RequestPayLoad{
		Term: "plus belle la vie",
		Options: Options{
			CollectionsLimit: 20,
			Types:            "collection",
		},
	}
	enc, err := json.Marshal(rq)
	if err != nil {
		panic(err)
	}

	resp, err := http.Post("https://www.france.tv/recherche/lancer/", "text/plain;charset=UTF-8", bytes.NewBuffer(enc))
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if len(body) < 2 {
		panic("Almost empty response")
	}

	var s string
	err = json.Unmarshal(body, &s)
	if err != nil {
		panic(err)
	}
	fmt.Println(s)

}
