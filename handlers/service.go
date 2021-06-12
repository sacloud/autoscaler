package handlers

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/sacloud/autoscaler/grpcutil"
	"github.com/sacloud/autoscaler/handler"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var _ handler.HandleServiceServer = (*handleService)(nil)

// handleService ハンドラ向けgRPCサーバの実装
type handleService struct {
	handler.UnimplementedHandleServiceServer
	Handler HandlerMeta
}

func (h *handleService) listenAndServe(parentCtx context.Context) error {
	errCh := make(chan error)
	ctx, stop := signal.NotifyContext(parentCtx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	listener, cleanup, err := grpcutil.Listener(&grpcutil.ListenerOption{
		Address: h.Handler.(Listener).ListenAddress(),
	})
	if err != nil {
		h.Handler.GetLogger().Fatal("fatal", err)
		return err // 到達しない
	}

	grpcServer := grpc.NewServer()
	handler.RegisterHandleServiceServer(grpcServer, h)
	reflection.Register(grpcServer)

	defer func() {
		grpcServer.GracefulStop()
		cleanup()
	}()

	go func() {
		if err := h.Handler.GetLogger().Info("message", "started", "address", listener.Addr().String()); err != nil {
			errCh <- err
		}
		if err := grpcServer.Serve(listener); err != nil {
			errCh <- err
		}
	}()

	for {
		select {
		case err := <-errCh:
			h.Handler.GetLogger().Error("error", err) // nolint
		case <-ctx.Done():
			h.Handler.GetLogger().Info("message", "shutting down", "error", ctx.Err()) // nolint
			return ctx.Err()
		}
	}
}

func (h *handleService) PreHandle(req *handler.PreHandleRequest, server handler.HandleService_PreHandleServer) error {
	logger := h.Handler.GetLogger().With(
		"scaling-job-id", req.ScalingJobId,
		"step", "PreHandle",
		"handler", handlerFullName(h.Handler),
	)

	if impl, ok := h.Handler.(PreHandler); ok {
		if err := logger.Info("status", handler.HandleResponse_RECEIVED); err != nil {
			return err
		}
		if err := logger.Debug("request", req.String()); err != nil {
			return err
		}
		return impl.PreHandle(req, server)
	}

	if err := logger.Info("status", handler.HandleResponse_IGNORED); err != nil {
		return err
	}
	return logger.Debug("request", req.String())
}

func (h *handleService) Handle(req *handler.HandleRequest, server handler.HandleService_HandleServer) error {
	logger := h.Handler.GetLogger().With(
		"scaling-job-id", req.ScalingJobId,
		"step", "Handle",
		"handler", handlerFullName(h.Handler),
	)

	if impl, ok := h.Handler.(Handler); ok {
		if err := logger.Info("status", handler.HandleResponse_RECEIVED); err != nil {
			return err
		}
		if err := logger.Debug("request", req.String()); err != nil {
			return err
		}
		return impl.Handle(req, server)
	}

	if err := logger.Info("status", handler.HandleResponse_IGNORED); err != nil {
		return err
	}
	return logger.Debug("request", req.String())
}

func (h *handleService) PostHandle(req *handler.PostHandleRequest, server handler.HandleService_PostHandleServer) error {
	logger := h.Handler.GetLogger().With(
		"scaling-job-id", req.ScalingJobId,
		"step", "PostHandle",
		"handler", handlerFullName(h.Handler),
	)

	if impl, ok := h.Handler.(PostHandler); ok {
		if err := logger.Info("status", handler.HandleResponse_RECEIVED); err != nil {
			return err
		}
		if err := logger.Debug("request", req.String()); err != nil {
			return err
		}
		return impl.PostHandle(req, server)
	}

	if err := logger.Info("status", handler.HandleResponse_IGNORED); err != nil {
		return err
	}
	return logger.Debug("request", req.String())
}
