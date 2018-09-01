
# Build

1) Install latest chrome

2) Install go/golang and godep

2) `godep go install`

3) `go build main.go`

4) generates distributable binary (main.exe or main).



# Run

`main.exe -u user@email.com -p password -c course-1-name -c course-2-name`

OR

`main.exe --username user@email.com --password password --course course-1-name --course course-2-name`

You can have one or more courses. just repeat the --course/-c argument.
chrome browser will be automatically opened/closed while scraping.

### To extract course name 

visit the page of the course and extract the course attribute.

https://app.pluralsight.com/player?***course=course-1-name***&author=author&name=course-m0&clip=0&mode=live

*course=course-1-name*
