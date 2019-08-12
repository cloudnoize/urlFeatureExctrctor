# intro

For now it queries google for the url, if google knows the host it marks as true.
Checks scheme (http/https) in google results.
Resolve DNS host, for now it doesn't validate and do anything with results.

## build
go build -o urlrep .

## run
./urlrep --query github.com
