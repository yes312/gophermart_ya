package app

import (
	"context"

	"gophermart/internal/config"
	db "gophermart/internal/database"
	"gophermart/internal/services"
	transport "gophermart/internal/transport/handlers"
	"net/http"
	"sync"

	jwtpackage "gophermart/pkg/jwt"

	"github.com/go-chi/chi"
	"go.uber.org/zap"
)

type Server struct {
	ctx     context.Context
	server  *http.Server
	config  *config.Config
	mux     *chi.Mux
	storage Storager
	logger  *zap.SugaredLogger
}

var _ Storager = &db.Storage{}

type Storager interface {
	db.StoragerDB
}

func New(ctx context.Context, config *config.Config) *Server {

	return &Server{
		ctx:    ctx,
		config: config,
	}
}

func (s *Server) Start(ctx context.Context, logger *zap.SugaredLogger, wg *sync.WaitGroup) error {

	s.logger = logger
	storage, err := db.New(ctx, s.config.DatabaseURI, s.config.MigrationsPath, logger)
	if err != nil {
		return err
	}
	s.storage = storage

	s.mux = s.ConfigureMux()

	s.server = &http.Server{
		Addr:    s.config.RunAdress,
		Handler: s.mux,
	}
	s.logger.Info("адрес сервера: " + s.config.RunAdress)
	a := services.NewAccrual(s.config.AccrualSysremAdress, s.config.AccrualRequestInterval, s.config.AccuralPuttingDBInterval, s.storage, s.logger, s.config.NumberOfWorkers)

	wg.Add(1)
	go a.RunAccrualRequester(ctx, wg)

	return s.server.ListenAndServe()
}

func (s *Server) Close() error {
	s.logger.Info("===Завершение работы сервера===")
	err := s.server.Shutdown(s.ctx)
	if err != nil {
		return err
	}
	err = s.storage.Close()
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) ConfigureMux() *chi.Mux {

	router := chi.NewRouter()
	handler := transport.New(s.ctx, s.storage, s.logger)
	handler.AuthToken = *jwtpackage.NewToken(s.config.TokenExp, s.config.Key)

	router.Route("/", func(r chi.Router) {

		r.Post("/api/user/register", handler.Registration)
		r.Post("/api/user/login", handler.Login)

		r.Post("/api/user/orders", handler.AuthMiddleware(handler.UploadOrders))     //загрузка пользователем номера заказа для расчёта;
		r.Get("/api/user/orders", handler.AuthMiddleware(handler.GetUploadedOrders)) //получение списка загруженных пользователем номеров заказов, статусов их обработки и информации о начислениях

		r.Get("/api/user/balance", handler.AuthMiddleware(handler.GetBalance))                //получение текущего баланса счёта баллов лояльности пользователя
		r.Post("/api/user/balance/withdraw", handler.AuthMiddleware(handler.WithdrawBalance)) //Запрос на списание средств

		r.Get("/api/user/withdrawals", handler.AuthMiddleware(handler.GetWithdrawals)) //Получение информации о выводе средств

	})

	return router
}
