# Lapislazuli

go mod tidy

go build -o .\lapislazuli.exe .\cmd\server\  

.\lapislazuli.exe

# en /actions para probar

go build -o view_pony.exe test.go

# curl para probar lo del pony

curl --location 'http://localhost:8080/index' \
--header 'Content-Type: application/json' \
--data '{"text": "Quiero mirar un caballo pequeño"}'