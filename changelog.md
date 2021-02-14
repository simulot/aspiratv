

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


