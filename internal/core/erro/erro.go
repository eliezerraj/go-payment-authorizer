package erro

import (
	"errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrCardTypeInvalid	= errors.New("card type invalid")
	ErrTransactioInvalid = errors.New("transaction is in null")
	ErrNotFound 		= errors.New("item not found")
	ErrUpdate			= errors.New("update unsuccessful")
	ErrServer		 	= errors.New("server identified error")
	ErrHTTPForbiden		= errors.New("forbiden request")
	ErrUnauthorized 	= errors.New("not authorized")
	MissingData 	= status.Errorf(codes.InvalidArgument, "header missing metadata")
)