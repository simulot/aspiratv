package ttml

import (
	"bytes"
	"encoding/xml"
	"strings"
	"testing"
)

func TestTTML_ToSrt(t *testing.T) {

	tests := []struct {
		name    string
		text    string
		wantDst string
		wantErr bool
	}{
		{
			name: "Case 1 Doctor Who",
			text: `<?xml version="1.0" encoding="utf-8"?>
			<tt xmlns="http://www.w3.org/ns/ttml"
				xmlns:ttm="http://www.w3.org/ns/ttml#metadata"
				xmlns:ttp="http://www.w3.org/ns/ttml#parameter"
				xmlns:tts="http://www.w3.org/ns/ttml#styling" xml:lang="fr">
				<head>
					<metadata>
						<ttm:title></ttm:title>
						<ttm:desc></ttm:desc>
						<ttm:copyright></ttm:copyright>
					</metadata>
					<styling>
						<style xml:id="captionStyle" tts:backgroundColor="transparent" tts:displayAlign="center" tts:extent="100% 20%" tts:fontFamily="proportionalSansSerif" tts:fontSize="30px" tts:origin="0% 75%" tts:textAlign="center" tts:textOutline="black 1px 0px" />
					</styling>
					<layout>
						<region style="captionStyle" xml:id="region2" />
					</layout>
				</head>
				<body>
					<div region="region2">
						<p begin="00:04:31.766" end="00:04:34.100" region="region2" xml:id="caption75" ttm:role="caption">
							<span tts:textAlign="center" tts:color="white">Tertio, ne paniquez pas.</span>
							<br></br>
							<span tts:textAlign="center" tts:color="white">Surtout Graham.</span>
						</p>
						<p begin="00:04:34.300" end="00:04:35.600" region="region2" xml:id="caption76" ttm:role="caption">
							<span tts:textAlign="center" tts:color="white">Je panique pas !</span>
						</p>
						<p begin="00:04:36.000" end="00:04:38.533" region="region2" xml:id="caption77" ttm:role="caption">
							<span tts:textAlign="center" tts:color="white">Si. Et j&apos;avais dit</span>
							<br></br>
							<span tts:textAlign="center" tts:color="white">de ne pas répondre.</span>
						</p>
						<p begin="00:04:38.733" end="00:04:39.466" region="region2" xml:id="caption78" ttm:role="caption">
							<span tts:textAlign="center" tts:color="white">Hein ?</span>
						</p>
						<p begin="00:04:39.666" end="00:04:42.266" region="region2" xml:id="caption79" ttm:role="caption">
							<span tts:textAlign="center" tts:color="white">Je fais vite.</span>
							<br></br>
							<span tts:textAlign="center" tts:color="white">La bombe a coupé le lien</span>
						</p>
					</div>
				</body>
			</tt>`,
			wantDst: `75
00:04:31,766 --> 00:04:34,100
Tertio, ne paniquez pas.
Surtout Graham.

76
00:04:34,300 --> 00:04:35,600
Je panique pas !

77
00:04:36,000 --> 00:04:38,533
Si. Et j'avais dit
de ne pas répondre.

78
00:04:38,733 --> 00:04:39,466
Hein ?

79
00:04:39,666 --> 00:04:42,266
Je fais vite.
La bombe a coupé le lien

`,
			wantErr: false,
		},
		{
			name: "Case 2 Nijago",
			text: `<?xml version="1.0" encoding="utf-8"?>
			<tt xmlns="http://www.w3.org/ns/ttml"
				xmlns:ttm="http://www.w3.org/ns/ttml#metadata"
				xmlns:ttp="http://www.w3.org/ns/ttml#parameter"
				xmlns:tts="http://www.w3.org/ns/ttml#styling" xml:lang="fr">
				<head>
					<metadata>
						<ttm:title></ttm:title>
						<ttm:desc></ttm:desc>
						<ttm:copyright></ttm:copyright>
					</metadata>
					<styling>
						<style xml:id="captionStyle" tts:backgroundColor="transparent" tts:displayAlign="center" tts:extent="100% 20%" tts:fontFamily="proportionalSansSerif" tts:fontSize="30px" tts:origin="0% 75%" tts:textAlign="center" tts:textOutline="black 1px 0px" />
					</styling>
					<layout>
						<region style="captionStyle" xml:id="region2" />
					</layout>
				</head>
				<body>
					<div region="region2">
						<p begin="00:00:00.000" end="00:00:01.266" region="region2" xml:id="caption1" ttm:role="caption">
							<span tts:textAlign="center" tts:color="Cyan">-Précédemment dans &quot;Ninjago&quot;.</span>
						</p>
						<p begin="00:00:01.533" end="00:00:03.133" region="region2" xml:id="caption2" ttm:role="caption">
							<span tts:textAlign="center" tts:color="white">-Nous avons des amis.</span>
						</p>
						<p begin="00:00:03.400" end="00:00:06.133" region="region2" xml:id="caption3" ttm:role="caption">
							<span tts:textAlign="center" tts:color="white">-Il semblerait qu&apos;on ne soit pas</span>
							<br></br>
							<span tts:textAlign="center" tts:color="white">les seuls dans ce royaume.</span>
						</p>
						<p begin="00:00:06.400" end="00:00:08.533" region="region2" xml:id="caption4" ttm:role="caption">
							<span tts:textAlign="center" tts:color="white">-Ils sont peut-être en vie.</span>
						</p>
						<p begin="00:00:08.800" end="00:00:10.200" region="region2" xml:id="caption5" ttm:role="caption">
							<span tts:textAlign="center" tts:color="yellow">-Alors ?</span>
						</p>
					</div>
				</body>
			</tt>`,
			wantErr: false,
			wantDst: `1
00:00:00,000 --> 00:00:01,266
<font color="Cyan">-Précédemment dans "Ninjago".</font>

2
00:00:01,533 --> 00:00:03,133
-Nous avons des amis.

3
00:00:03,400 --> 00:00:06,133
-Il semblerait qu'on ne soit pas
les seuls dans ce royaume.

4
00:00:06,400 --> 00:00:08,533
-Ils sont peut-être en vie.

5
00:00:08,800 --> 00:00:10,200
<font color="yellow">-Alors ?</font>

`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			xTT := &TTML{}
			err := xml.NewDecoder(strings.NewReader(tt.text)).Decode(&xTT)
			if err != nil {
				t.Errorf("Can't parse XML: %w", err)
				return
			}
			dst := &bytes.Buffer{}
			if err := xTT.ToSrt(dst); (err != nil) != tt.wantErr {
				t.Errorf("TTML.ToSrt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotDst := dst.String(); gotDst != tt.wantDst {
				t.Errorf("TTML.ToSrt() = %v, want %v", gotDst, tt.wantDst)
			}
		})
	}
}
