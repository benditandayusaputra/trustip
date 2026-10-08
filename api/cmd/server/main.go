package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	authhandler "github.com/benditandayusaputra/trustip/api/internal/modules/auth/handler"
	filehandler "github.com/benditandayusaputra/trustip/api/internal/modules/file/handler"
	notificationhandler "github.com/benditandayusaputra/trustip/api/internal/modules/notification/handler"
	orghandler "github.com/benditandayusaputra/trustip/api/internal/modules/org/handler"
	"github.com/benditandayusaputra/trustip/api/internal/platform/clock"
	"github.com/benditandayusaputra/trustip/api/internal/platform/config"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver"
	"github.com/benditandayusaputra/trustip/api/internal/platform/idgen"
	"github.com/benditandayusaputra/trustip/api/internal/platform/logger"
	"github.com/benditandayusaputra/trustip/api/internal/platform/ratelimit"
	"github.com/benditandayusaputra/trustip/api/internal/platform/scheduler"
	"github.com/benditandayusaputra/trustip/api/internal/platform/security"
)

const requestsPerMinutePerIP = 300

func main() {
	cfg, err := config.Load(os.Environ())
	if err != nil {
		logger.New(os.Stderr, slog.LevelInfo).Error("refusing to start", "error", err)
		os.Exit(1)
	}
	base := logger.New(os.Stdout, cfg.LogLevel)
	appLog := logger.ForStream(base, logger.App)
	if err := serve(cfg, base); err != nil {
		appLog.Error("server stopped", "error", err)
		os.Exit(1)
	}
	appLog.Info("server stopped")
}

func serve(cfg config.Config, base *slog.Logger) error {
	pool, err := database.Open(context.Background(), database.Options{
		URL:               cfg.DatabaseURL,
		Logger:            logger.ForStream(base, logger.App),
		Clock:             clock.System{},
		RefuseSchemaOwner: cfg.Environment == config.Production,
	})
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer pool.Close()
	systemClock := clock.System{}
	components, err := buildAuthComponents(cfg, pool, systemClock)
	if err != nil {
		return err
	}
	trade := buildTradeComponents(pool, systemClock, cfg.SealKeys)
	server := newServer(base, serverDeps{
		ready:     pool.Ping,
		edge:      security.NewEdgeVerifier(cfg.EdgeKeys, systemClock),
		limiter:   ratelimit.New(ratelimit.Rule{Limit: requestsPerMinutePerIP, Window: time.Minute}, systemClock),
		clock:     systemClock,
		auth:      components,
		org:       buildOrgComponents(pool, systemClock),
		files:     buildFileComponents(pool, cfg, systemClock),
		inbox:     buildNotificationComponents(pool, systemClock),
		trade:     trade,
		appOrigin: cfg.PublicAppOrigin,
	})
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	appLog := logger.ForStream(base, logger.App)
	jobs := newScheduler(schedulerParts{
		runners: buildJobRunners(pool, trade), clock: systemClock, waits: scheduler.SystemWaits{}, log: appLog,
	})
	defer startJobs(ctx, jobs, stop)()
	appLog.Info("server starting", "addr", cfg.HTTPAddr, "environment", cfg.Environment)
	if err := server.Run(ctx, cfg.HTTPAddr); err != nil {
		return fmt.Errorf("run http server: %w", err)
	}
	return nil
}

func startJobs(ctx context.Context, jobs *scheduler.Scheduler, stop context.CancelFunc) func() {
	var background sync.WaitGroup
	background.Go(func() { jobs.Run(ctx) })
	return func() {
		stop()
		background.Wait()
	}
}

type serverDeps struct {
	ready     httpserver.ReadinessCheck
	edge      httpserver.EdgeVerifier
	limiter   httpserver.RateLimiter
	clock     serverClock
	auth      authComponents
	org       orgComponents
	files     fileComponents
	inbox     notificationComponents
	trade     tradeComponents
	appOrigin string
}

type serverClock interface {
	Now() time.Time
}

func newServer(base *slog.Logger, deps serverDeps) *httpserver.Server {
	securityLog := logger.ForStream(base, logger.Security)
	server := httpserver.New(httpserver.Options{
		AppLog:      logger.ForStream(base, logger.App),
		AccessLog:   logger.ForStream(base, logger.Access),
		SecurityLog: securityLog,
		IDs:         idgen.UUIDv7{},
		Clock:       deps.clock,
		Build:       httpserver.ReadBuildInfo(),
		Ready:       deps.ready,
		Edge:        deps.edge,
		Limiter:     deps.limiter,
		AppOrigin:   deps.appOrigin,
	})
	authhandler.Mount(server.API(), authDeps(deps.auth, securityLog, deps.clock))
	authenticate := httpserver.Authenticate(newAccessVerifier(deps.auth), deps.auth.cookies.AccessName())
	writeLimiter := newWriteLimiter(deps.clock)
	orghandler.Mount(server.API(), orgDeps(deps.org, orgRoutes{
		authenticate: authenticate,
		filesBaseURL: deps.files.publicBaseURL,
		writeLimiter: writeLimiter,
		securityLog:  securityLog,
	}))
	files := fileDeps(deps.files, fileRoutes{
		authenticate: authenticate, limiterClock: deps.clock, securityLog: securityLog,
	})
	filehandler.Mount(server.API(), files)
	filehandler.MountPublic(server.App(), files)
	notificationhandler.Mount(server.API(), notificationDeps(deps.inbox, notificationRoutes{
		authenticate: authenticate, writeLimiter: writeLimiter, securityLog: securityLog,
	}))
	mountTrade(server.API(), deps.trade, tradeRoutes{
		authenticate: authenticate, filesBaseURL: deps.files.publicBaseURL, writeLimiter: writeLimiter,
		securityLog: securityLog, clock: deps.clock,
	})
	return server
}

func publicRoutes() []httpserver.PublicRoute {
	routes := append(httpserver.PublicRoutes(), authhandler.PublicRoutes()...)
	routes = append(routes, orghandler.PublicRoutes()...)
	return append(routes, filehandler.PublicRoutes()...)
}
