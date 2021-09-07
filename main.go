package main

import (
	"database/sql"
	"flag"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/hostrouter"
	"github.com/stevealexrs/Go-Libra/account"
	"github.com/stevealexrs/Go-Libra/account/accountrouter"
	"github.com/stevealexrs/Go-Libra/database/redisdb"
	"github.com/stevealexrs/Go-Libra/email"
	"github.com/stevealexrs/Go-Libra/session"
)

func main() {
	// flag
	sqlDSN := flag.String("sql", "", "Data source name of the sql relational database")
	rsMaster := flag.String("rs-master", "", "Name of redis sentinel master")
	rsNodes := flag.String("rs-nodes", "", "A list of space-separated host:port addresses of redis sentinel nodes")
	rsPassword := flag.String("rs-p", "", "Password of redis sentinel")
	smtpPlain := flag.String("smtp", "", "A list of space-separated values that consists of email, password, hostname, and server name for plain auth")

	flag.Parse()

	// master router
	r := chi.NewRouter()
	hr := hostrouter.New()

	// A good base middleware stack
	r.Use(
		middleware.RequestID,
		middleware.Logger,
		middleware.Recoverer,

		// Set a timeout value on the request context (ctx), that will signal
		// through ctx.Done() that the request has timed out and further
		// processing should be stopped.
		middleware.Timeout(60*time.Second),
		middleware.Heartbeat("/ping"),
	)

	redisSentinelClient := redis.NewFailoverClient(&redis.FailoverOptions{
		MasterName: *rsMaster,
		SentinelAddrs: strings.Split(*rsNodes, " "),
		Password: *rsPassword,
	})

	sqlDB, err := sql.Open("mysql", *sqlDSN)
	if err != nil {
		panic(err)
	}
	
	redisDB := redisdb.NewRedisHandlerWithClient(redisSentinelClient)

	smtpArgs := strings.Split(*smtpPlain, " ")
	if len(smtpArgs) != 4 {
		panic("invalid smtp flag")
	}
	plainAuth := email.NewSMTPService(smtpArgs[0], smtpArgs[1], smtpArgs[2], smtpArgs[3])

	hr.Map("localhost:1337", defaultRouter())
	hr.Map("api.localhost:1337", apiRouter(sqlDB, redisDB, plainAuth))

	r.Mount("/", hr)

	log.Fatal(http.ListenAndServe(":1337", r))
}

func apiRouter(sqlDB *sql.DB, redisDB *redisdb.Handler, mailService email.Service) chi.Router {
	r := chi.NewRouter()

	userRepo := account.UserRepo{
		DB: sqlDB,
	}

	businessRepo := account.BusinessRepo{
		DB: sqlDB,
	}

	emailClient := email.Client{
		Service: mailService,
	}

	accRouter := accountrouter.New(
		account.UserCreator{
			UserRepo:       &userRepo,
			InvitationRepo: account.NewInvitationEmailVerificationRepo(redisDB, ""),
			EmailRepo:      account.NewRecoveryEmailVerificationRepo(redisDB, ""),
			Ext:            &emailClient,
		},
		session.NewDefSharedProvider(redisDB, ""),
		account.UserAccountRecoveryHelper{
			UserRepo: &userRepo,
			RecoveryRepo: account.NewAccountRecoveryRepo(redisDB, ""),
			Ext: &emailClient,
		},
		account.BusinessCreator{
			BusinessRepo: &businessRepo,
			EmailRepo: account.NewRecoveryEmailVerificationRepo(redisDB, ""),
			Ext: &emailClient,
		},
		session.NewDefSharedProvider(redisDB, ""),
		account.BusinessAccountRecoveryHelper{
			BusinessRepo: &businessRepo,
			RecoveryRepo: account.NewAccountRecoveryRepo(redisDB, ""),
			Ext: &emailClient,
		},
	)

	r.Mount("/users", accRouter.UserHandler())
	r.Mount("/businesses", accRouter.BusinessHandler())
	return r
}

func defaultRouter() chi.Router {
	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Welcome to the main page."))
	})

	return r
}