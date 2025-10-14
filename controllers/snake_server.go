package controllers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/secomp2025/localsnake/database"
	"github.com/secomp2025/localsnake/game"
)

type SnakeServer struct {
	Addr    string
	path    string
	port    int
	command *exec.Cmd
	logFile *os.File
}

type SnakeServerController struct {
	lock          sync.RWMutex
	httpClient    *http.Client
	servers       map[int64]SnakeServer
	reservedPorts map[int]bool
	basePort      int

	pyServerPath string
	jsServerPath string
	cServerPath  string
}

type globalSnakeServerController struct {
	lock       sync.Mutex
	controller *SnakeServerController
}

var serverManager globalSnakeServerController

const (
	BASE_PORT = 8000
	MAX_PORTS = 300
)

func InitSnakeServerManager(staticFS fs.FS) {
	serverManager.lock.Lock()
	defer serverManager.lock.Unlock()
	if serverManager.controller != nil {
		return
	}

	pyServerPath := filepath.Join("static", "code-templates", "py", "server.py")
	pyServerFile, err := staticFS.Open(pyServerPath)
	if err != nil {
		log.Println("Error opening server file for snake", err)
		return
	}
	defer pyServerFile.Close()

	copyFile(pyServerFile, pyServerPath)

	serverManager.controller = &SnakeServerController{httpClient: &http.Client{}, basePort: BASE_PORT, servers: make(map[int64]SnakeServer), reservedPorts: make(map[int]bool), pyServerPath: pyServerPath}
}

func GetServerManager() *SnakeServerController {
	return serverManager.controller
}

func DestroySnakeServerManager() {
	serverManager.lock.Lock()
	defer serverManager.lock.Unlock()
	if serverManager.controller == nil {
		return
	}

	for _, server := range serverManager.controller.servers {
		if server.command != nil {
			server.command.Process.Kill()
			server.command.Wait()
		}
	}

	serverManager.controller = nil
}

func (c *SnakeServerController) GetSnakeStatus(ctx context.Context, snake *database.Snake) (game.Status, error) {

	snakeID := snake.ID
	server := c.getServer(snakeID)
	if server == nil {
		return game.StatusOffline, nil
	}

	// log.Println("Checking status for snake", snake)
	resp, err := c.httpClient.Get(server.Addr)
	if err != nil {
		log.Println("Error checking status for snake", snake, err)
		c.stopAndRemoveServer(snakeID)
		return game.StatusOffline, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Println("Error checking status for snake", snake, resp.StatusCode)
		c.stopAndRemoveServer(snakeID)
		return game.StatusOffline, nil
	}

	return game.StatusOnline, nil
}

func (c *SnakeServerController) ManageSnake(ctx context.Context, snake *database.Snake) error {
	snakeID := snake.ID
	if c.serverExists(snakeID) {
		log.Println("Snake already managed", snake)
		return nil
	}

	log.Println("Finding empty port for snake", snake)
	port, err := c.getEmptyPort()
	if err != nil {
		log.Println("Error getting empty port for snake", snake, err)
		return err
	}

	log.Println("Starting snake server for snake", snake, "on port", port)

	logFile, err := os.OpenFile("snake-"+strconv.FormatInt(snakeID, 10)+".log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println("Error creating log file for snake", snake, err)
		return err
	}

	var serverCommand *exec.Cmd

	if strings.HasSuffix(snake.Path, ".py") {
		serverCommand = exec.Command("python", c.pyServerPath, snake.Path, strconv.Itoa(port))
		serverCommand.Env = append(os.Environ(), "PYTHONUNBUFFERED=1")
	} else if strings.HasSuffix(snake.Path, ".js") {
		serverCommand = exec.Command("npm", "run", "serve", "--", snake.Path, strconv.Itoa(port))
		serverCommand.Dir = c.jsServerPath
	} else if strings.HasSuffix(snake.Path, ".c") {
		// compile and run shared object
		sharedObjectPath := snake.Path[:len(snake.Path)-2] + ".so"
		compileCmd := exec.Command("gcc", "-shared", "-o", sharedObjectPath, snake.Path)
		if err := compileCmd.Run(); err != nil {
			return fmt.Errorf("error compiling snake %d: %w", snake.ID, err)
		}

		serverCommand = exec.Command("LD_PRELOAD="+sharedObjectPath, c.cServerPath, snake.Path, strconv.Itoa(port))
	} else {
		return fmt.Errorf("invalid snake file extension: %s", snake.Path)
	}

	serverCommand.Stdout = logFile
	serverCommand.Stderr = logFile

	serverCommand.Start()

	log.Println("Snake server for snake", snake, "started on port", port)

	c.addServer(snakeID, SnakeServer{
		Addr:    "http://localhost:" + strconv.Itoa(port),
		path:    snake.Path,
		port:    port,
		command: serverCommand,
		logFile: logFile,
	})
	return nil
}

func (c *SnakeServerController) StopSnake(snakeID int64) {
	c.stopServer(snakeID)
}

func (c *SnakeServerController) StopAndRemoveSnake(snakeID int64) {
	c.stopAndRemoveServer(snakeID)
}

func (c *SnakeServerController) GetServer(snakeID int64) *SnakeServer {
	return c.getServer(snakeID)
}

func (c *SnakeServerController) getServer(snakeID int64) *SnakeServer {
	c.lock.RLock()
	defer c.lock.RUnlock()
	server, ok := c.servers[snakeID]
	if !ok {
		return nil
	}
	return &server
}

func (c *SnakeServerController) serverExists(snakeID int64) bool {
	c.lock.RLock()
	defer c.lock.RUnlock()
	_, ok := c.servers[snakeID]
	return ok
}

func (c *SnakeServerController) getEmptyPort() (int, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	for port := range MAX_PORTS {
		port += c.basePort
		if v, ok := c.reservedPorts[port]; !ok || !v {
			c.reservedPorts[port] = true
			return port, nil
		}
	}
	return 0, errors.New("no available ports")
}

func (c *SnakeServerController) addServer(snakeID int64, server SnakeServer) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.servers[snakeID] = server
}

func (c *SnakeServerController) stopServer(snakeID int64) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if _, ok := c.servers[snakeID]; !ok {
		log.Println("Snake server for snake", snakeID, "not found")
		return
	}

	server := c.servers[snakeID]
	if server.command != nil {
		server.command.Process.Kill()
		server.command.Wait()
	}

	log.Println("Snake server for snake", snakeID, "stopped")
}

func (c *SnakeServerController) stopAndRemoveServer(snakeID int64) {
	c.stopServer(snakeID)
	c.removeServer(snakeID)
}

func (c *SnakeServerController) removeServer(snakeID int64) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if _, ok := c.servers[snakeID]; !ok {
		log.Println("Snake server for snake", snakeID, "not found")
		return
	}

	log.Println("Removing snake server for snake", c.servers[snakeID])

	c.reservedPorts[c.servers[snakeID].port] = false
	delete(c.servers, snakeID)
	c.servers[snakeID].logFile.Close()
}

func copyFile(src fs.File, dst string) error {
	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, src); err != nil {
		return err
	}
	return nil
}
