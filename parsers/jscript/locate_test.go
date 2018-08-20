package jscript

import (
	"fmt"
	"io/ioutil"
	"path"
	"reflect"
	"regexp"
	"testing"
)

func TestAnchorIndex(t *testing.T) {
	type args struct {
		b      []byte
		anchor *regexp.Regexp
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			"nil",
			args{
				nil,
				regexp.MustCompile(`exports=\{`),
			},
			-1,
		},
		{
			"empty",
			args{
				[]byte(""),
				regexp.MustCompile(`exports=\{`),
			},
			-1,
		},
		{
			"no object",
			args{
				[]byte("something:llll"),
				regexp.MustCompile(`exports=\{`),
			},
			-1,
		},
		{
			"tooshort",
			args{
				[]byte("1234"),
				regexp.MustCompile(`exports=\{`),
			},
			-1,
		},
		{
			"easy",
			args{
				[]byte("exports={obj:12}"),
				regexp.MustCompile(`exports=\{`),
			},
			8,
		},
		{
			"long",
			args{
				[]byte(`void 0===(a="function"==typeof(r=function(e){"use strict";e.exports={obj:12};}`),
				regexp.MustCompile(`exports=\{`),
			},
			68,
		},
		{
			"repeated",
			args{
				[]byte(`void 0===(a="function"==typeof(r=function(e){"use strict";e.exports={obj:12};}void 0===(a="function"==typeof(r=function(e){"use strict";e.exports={obj:12};}`),
				regexp.MustCompile(`exports=\{`),
			},
			68,
		},
		{
			"anchor",
			args{
				[]byte(`void 0===(a="function"==typeof(r=function(e){"use strict";e.exports={obj:12};};void 0===(a="function"==typeof(r=function(e){"use strict";e.exports={fr:{Club:{home:"Accéder mon ARTE",profile:"Modifier mon profil",logout:"Me déconnecter",pseudo:"Mon ARTE"},LogoNavigation:{label:"Accueil",href:"https://www.arte.tv/fr/"},DesktopNavigation:{ariaLabel:"Menu secondaire",links:[{label:"Guide +7"`),
				regexp.MustCompile(`exports=\{[a-z]{2}:\{Club`),
			},
			147,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AnchorIndex(tt.args.b, tt.args.anchor); got != tt.want {
				t.Errorf("AnchorIndex() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindObjectEnd(t *testing.T) {
	type args struct {
		b           []byte
		objectStart int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			"simple",
			args{
				[]byte(`a={b:455};another stuff`),
				2,
			},
			9,
		},
		{
			"unclosed object",
			args{
				[]byte(`a={b:'455';another stuff`),
				2,
			},
			-1,
		},
		{
			"simple and double quote string",
			args{
				[]byte(`a={b:"455"};another stuff`),
				2,
			},
			11,
		},
		{
			"simple and simple quote string",
			args{
				[]byte(`a={b:'455'};another stuff`),
				2,
			},
			11,
		},
		{
			"simple quote in double quote string",
			args{
				[]byte(`a={b:"o'connor"}`),
				2,
			},
			16,
		},
		{
			"double quote in simple quote string",
			args{
				[]byte(`a={b:'o"connor'}`),
				2,
			},
			16,
		},
		{
			"escaped simple quote in simple quote string",
			args{
				[]byte(`a={b:'o\'connor'}`),
				2,
			},
			17,
		},
		{
			"{ in double quote string",
			args{
				[]byte(`a={b:"o{connor"}`),
				2,
			},
			16,
		},
		{
			"} in double quote string",
			args{
				[]byte(`a={b:"o}connor"}`),
				2,
			},
			16,
		},
		{
			"ログインメールアドレス in double quote string",
			args{
				[]byte(`a={b:"ログインメールアドレス"}`),
				2,
			},
			41,
		},
		{
			"\\U7B7C in double quote string",
			args{
				[]byte(`a={b:"\U7B7CABLE"}`),
				2,
			},
			18,
		},

		{
			"nested object",
			args{
				[]byte(`a={b:"o'connor",c:{d:52},d:522}`),
				2,
			},
			31,
		},
		{
			"nested object with extra }",
			args{
				[]byte(`a={b:"o'connor",c:{d:52},d:522}}`),
				2,
			},
			31,
		},
		{
			"nested object unclosed",
			args{
				[]byte(`a={b:"o'connor",c:{d:52,d:522}`),
				2,
			},
			-1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FindObjectEnd(tt.args.b, tt.args.objectStart); got != tt.want {
				t.Errorf("FindObjectEnd() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRealFile(t *testing.T) {
	fileName := path.Join("testdata", "player.js")
	f, err := ioutil.ReadFile(fileName)
	if err != nil {
		t.Error(err)
		return
	}

	obj := ObjectAtAnchor(f, regexp.MustCompile(`exports=\{[a-z]{2}:\{Club`))
	fmt.Println(string(obj))

}

func TestObjectAtAnchor(t *testing.T) {
	type args struct {
		b      []byte
		anchor *regexp.Regexp
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			"test1",
			args{
				[]byte(`this crap before object={home:"Home page",profile:"Change my profile",logout:"Logout",login:"unknown"} and crap after `),
				regexp.MustCompile(`object=\{`),
			},
			[]byte(`{home:"Home page",profile:"Change my profile",logout:"Logout",login:"unknown"}`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ObjectAtAnchor(tt.args.b, tt.args.anchor); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ObjectAtAnchor() = %v, want %v", string(got), string(tt.want))
			}
		})
	}
}
