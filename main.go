package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/secomp2025/localsnake/controllers"
	"github.com/secomp2025/localsnake/database"

	// "github.com/secomp2025/localsnake/game"
	"github.com/secomp2025/localsnake/handlers"
)

//go:embed static
var staticFiles embed.FS

var isShuttingDown atomic.Bool

const (
	shutdownPeriod      = 15 * time.Second
	shutdownHardPeriod  = 3 * time.Second
	readinessDrainDelay = 5 * time.Second
)

func main() {
	devMode := os.Getenv("DEV_MODE") == "1"

	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// component := pages.Hello("World")

	// Initialize database connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := database.Init(ctx, "a.db"); err != nil {
		panic(err)
	}
	defer database.Close()

	var staticFS fs.FS
	if devMode {
		staticFS = staticFiles
		http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	} else {
		var err error
		staticFS, err = fs.Sub(staticFiles, "static")
		if err != nil {
			panic(err)
		}

		http.Handle(
			"/static/",
			http.StripPrefix("/static/",
				http.FileServer(http.FS(staticFS))),
		)
	}

	var destroyOnce sync.Once
	controllers.InitSnakeServerManager(staticFS)
	defer destroyOnce.Do(controllers.DestroySnakeServerManager)

	// Routes
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if isShuttingDown.Load() {
			http.Error(w, "Shutting down", http.StatusServiceUnavailable)
			return
		}

		fmt.Fprintln(w, "Ok")
	})

	http.HandleFunc("/", handlers.HomePage)
	http.HandleFunc("/login", handlers.LoginHandler)
	http.HandleFunc("/register", handlers.Register)
	http.HandleFunc("/upload-snake", handlers.UploadSnake)
	http.HandleFunc("/logout", handlers.Logout)
	http.HandleFunc("/status", handlers.StatusHandler)

	http.HandleFunc("/create-game", handlers.CreateTeamGameHandler)
	// handle /game/<game_id>
	http.HandleFunc("/game/", handlers.GameHandler)

	http.HandleFunc("/login-adm", handlers.PostLoginAdm)
	http.HandleFunc("/adm", handlers.AdminHandler)

	http.HandleFunc("/battle", handlers.HandleBattle)
	http.HandleFunc("/rerun", handlers.RerunHandler)

	ongoingCtx, stopOngoingGracefully := context.WithCancel(context.Background())
	server := &http.Server{
		Addr: ":3000",
		BaseContext: func(_ net.Listener) context.Context {
			return ongoingCtx
		},
	}
	go func() {
		log.Printf("Starting server at %s\n", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	<-rootCtx.Done()
	stop()

	isShuttingDown.Store(true)
	log.Println("Shutting down")

	destroyOnce.Do(controllers.DestroySnakeServerManager)

	if !devMode {
		time.Sleep(readinessDrainDelay)
		log.Println("Waiting for ongoing requests to finish")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownPeriod)
		defer cancel()

		err := server.Shutdown(shutdownCtx)
		stopOngoingGracefully()
		if err != nil {
			log.Println("Failed to wait for ongoing requests to finish")
			time.Sleep(shutdownHardPeriod)
		}
	}
	log.Println("Server shut down")
}
