package auth

import (
	"context"

	ssov1 "github.com/finaptica/protos/gen/go/sso"
	"github.com/finaptica/sso/internal/lib/errs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthService interface {
	Login(ctx context.Context, email string, password string, appId int) (token string, err error)
	Register(ctx context.Context, email, password string) (userId int64, err error)
}

type serverAPI struct {
	ssov1.UnimplementedAuthServer
	authService AuthService
}

func Register(gRPC *grpc.Server, auth AuthService) {
	ssov1.RegisterAuthServer(gRPC, &serverAPI{authService: auth})
}

const (
	emptyValue = 0
)

func (s *serverAPI) Register(ctx context.Context, req *ssov1.RegisterRequest) (*ssov1.RegisterResponse, error) {
	if err := validateRegister(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	id, err := s.authService.Register(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		switch errs.KindOf(err) {
		case errs.AlreadyExists:
			return nil, status.Error(codes.AlreadyExists, "Email already registered")
		default:
			return nil, errs.ToStatus(err)
		}
	}

	return &ssov1.RegisterResponse{UserId: id}, nil
}

func (s *serverAPI) Login(ctx context.Context, req *ssov1.LoginRequest) (*ssov1.LoginResponse, error) {
	if err := validateLogin(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	token, err := s.authService.Login(ctx, req.GetEmail(), req.GetPassword(), int(req.AppId))
	if err != nil {
		if errs.KindOf(err) == errs.Unauthenticated {
			return nil, status.Error(codes.Unauthenticated, "Invalid email or password")
		}
		return nil, errs.ToStatus(err)
	}

	return &ssov1.LoginResponse{
		Token: token,
	}, nil
}

func validateLogin(req *ssov1.LoginRequest) error {
	if req.GetEmail() == "" {
		return status.Error(codes.InvalidArgument, "Email is required")
	}

	if req.GetPassword() == "" {
		return status.Error(codes.InvalidArgument, "Password is required")
	}

	if req.GetAppId() == emptyValue {
		return status.Error(codes.InvalidArgument, "App ID is required")
	}

	return nil
}

func validateRegister(req *ssov1.RegisterRequest) error {
	if req.GetEmail() == "" {
		return status.Error(codes.InvalidArgument, "Email is required")
	}

	if req.GetPassword() == "" {
		return status.Error(codes.InvalidArgument, "Password is required")
	}

	return nil
}
