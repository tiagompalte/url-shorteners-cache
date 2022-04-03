# URL Shorteners Cache

Golang + Redis

## Routes

- POST /

curl --location --request POST 'http://localhost:8000' \
  --form 'url="http://www.google.com"'

- GET /

curl --location --request GET 'http://localhost:8000/m-9JiXynR'

- GET /{url_short}/count

curl --location --request GET 'http://localhost:8000/m-9JiXynR/count'