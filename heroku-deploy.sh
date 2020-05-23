#!/bin/bash

# Build the go binary
echo "Builiding the talky binary"
go build -o bin/talky cmd/main.go

echo "Commiting the binary to git."
git commit -am "Generated new binary."

echo "Pushing the api to heroku."
git push heroku master

# Push the client to heroku
git push heroku-nuxt master
