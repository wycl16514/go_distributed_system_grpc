module server

go 1.17

replace order_management => ../proto/service/order_management

require (
	github.com/golang/protobuf v1.5.2
	google.golang.org/grpc v1.43.0
	order_management v0.0.0-00010101000000-000000000000
)

require (
	golang.org/x/net v0.0.0-20200822124328-c89045814202 // indirect
	golang.org/x/sys v0.0.0-20200323222414-85ca7c5b95cd // indirect
	golang.org/x/text v0.3.0 // indirect
	google.golang.org/genproto v0.0.0-20200526211855-cb27e3aa2013 // indirect
	google.golang.org/protobuf v1.26.0 // indirect
)
