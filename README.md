# githit
Find trending git repositories.

This is in an very early stage, at the moment the core of the API works, but it 
is not stable and there is no Frontend yet.

## Run the server
1) Install golang
2) Run the following:
```
go build
TWITTER_API_KEY=XXXX \
TWITTER_API_SECRET=XXXX \
TWITTER_ACCESS_TOKEN=XXXX \
TWITTER_ACCESS_TOKEN_SECRET=XXXX \
./githit
```
Of course you have to replace the tokens with the ones twitter provided you.
3) Open [localhost:3000/api/projects](localhost:3000/api/projects) in your browser