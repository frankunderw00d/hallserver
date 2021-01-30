module hallserver

go 1.14

require (
	baseservice v1.0.1
	github.com/gomodule/redigo v1.8.3
	jarvis v1.0.1
)

replace (
	baseservice v1.0.1 => ../baseservice
	jarvis v1.0.1 => ../jarvis
)
