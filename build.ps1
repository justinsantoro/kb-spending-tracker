$env:GOOS='linux'
$env:CGO_ENABLED=0
go build .
$env:GOOS='windows'