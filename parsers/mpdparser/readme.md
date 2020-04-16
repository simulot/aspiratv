# MPD parser

I have encountered few problems with ffmpeg when playing FRANCE2 content delivered with DASH protocol.

- manifest.mpd: Invalid data found when processing input (maybe subtitle stream)
- wrong video size selection
- HTTP 302 redirection not followed



The idea is to get the MPD from the server and edit it for:
- follow the redirection to get the licensed content
- removing the subtitles
- removing all but best video quality
- serve this edited mpd to ffmpeg
