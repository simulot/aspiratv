# version 0.16.0
## 🛠️ Major code refactoring 🛠️
This work is done to prepare a web interface for Aspiratv.  

## ⚠️ BREAKING CHANGE: ⚠️
  Command line has changed. Read documentation page

## Changelog
- decouple user interface from pulling and download tasks. This will ease the realization of a web front end.
- fix clunky code 
    - concurency management
    - logs
- Change command line flag management. (breaking change)
    - now use pflag library to provide a set of options for each sub-commands
        - command `run` (default) for dowloading show according config.json file
        - command `download` for downloading a show according options on the command line
- Remove over complicated code
    - http client
    - html parsing
- Remove unused code
- Update dependencies
    - multiple progress bars
- Fix go lint warnings
- Fix FFMPEG errors not sent to the log file
- Improve ^C handling
    - remove dirs created by interupted download
    - Faster exit



# version 0.15.0

- Fix case in Show path
- Rewriting Download functions to avoid path error for nfo
- Change path for show specials
    -francetv:
        - fix missing thumbnail
        - skip non available episodes
        - download extras, teasers and bonnuses
- Fix #71 francetv regression on episode and season
- Fix a error in config.json file conaintned in readme.
- francetv: don't grab aired time on detail page

# version 0.14.0

## New features
- add flags
    - --name-template to give the media's file name template
    - --season-template go give the season part template of the path
- add fields to config.json
    - 	SeasonPathTemplate
        > Template for season path, can be empty to skip season in path. When missing uses default naming
	-   ShowNameTemplate    
        >Template for the name of mp4 file, can't be empty. When missing, uses default naming
	-   TitleFilter         
        >ShowTitle or Episode title must match this regexp to be downloaded
	-   TitleExclude        
        >ShowTitle and Episode title must not match this regexp to be downloaded
- Name template implementation
    - for movies
        > `{{.Title}}.mp4`
    - for shows
        > `{{.Showtitle}} - {{.Aired.Time.Format "2006-01-02"}}.mp4`
    - for series
        > `{{.Showtitle}} - s{{.Season | printf "%02d" }}e{{.Episode | printf "%02d" }} - {{.Title}}.mp4`
- Season path template implementation
    - for movies
        >   (empty)
    - for shows:
        > `Season {{.Aired.Time.Year | printf "%04d" }}`
    - for series
        > `Season {{.Season | printf "%02d" }}`

## Fixes 
- artetv
    - Fix #67 [artetv] Can't visit URL "Too Many Requests"
        - Implementation of query throttling
        - Reduce query to search API 

## Other changes
- franctev
    - Get Aired and ID form "Toutes les pages" used implement template


