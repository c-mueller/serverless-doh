module github.com/c-mueller/serverless-doh/lambda

go 1.12

require (
	github.com/akrylysov/algnhsa v0.0.0-20190319020909-05b3d192e9a7
	github.com/c-mueller/serverless-doh v0.0.0-20190718184545-54f7558bfd7b
	golang.org/x/sync v0.0.0-20190423024810-112230192c58 // indirect
)

replace github.com/c-mueller/serverless-doh => ../
