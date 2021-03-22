# Aspiratv startover

Goal: takes the occasion of writing a webapp for ASPIRATV for refactoring the code using Test Driven Development
See https://quii.gitbook.io/learn-go-with-tests/


## webapp


- Design principles:
    - Use a of very simple CSS frame work
        - avoid grunt / webpack / other contraptions
    - Use of GO for frontend Development
        - "to old for starting js now"
    -
- Deployability
    - prepare a docker image / docker-compose script

- side goals
    - learn HTTP/2, websockets
    - use end to end context
    - Write good GO code
        - Don't reuse existing code directly: be inspired and copy best parts
        - fewer abstractions
        - less packages
        - more modularity
    - Use Test Driven Development
        - make the code testable
        - learn 
            - how to mock up external site dependency
            - how to test web application 



## Command line (should I do that?)
- run some functionalities of web app as command line 


# Design

[](design.svg)

## Web application

The front end is an PWA (progressive web application). It runs into the web browser. It should be fast, mobile or web tv capable. 
It consumes applications server's api to display application's outputs.
## Web application server

This part is in charge of serving the front end web application. It will present APIs 

## Data persistence layer

All application data is stored in to json files. Metadata will be stored into nfo files.
Media are converted if needed and stored as mp4 files.


## Site manager

This layer manage subscriptions, searches using TV site adapters.
