package game

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os/exec"
	"time"
)

type GameProc struct {
	command  *exec.Cmd
	stdin    io.WriteCloser
	readChan chan string
	started  bool
}

func New(command string) *GameProc {
	cmd := exec.Command("stdbuf", "-oL", command)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Print("Error piping stdin")
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Print("Error piping stdout")
	}
	cmd.Start()
	go cmd.Wait()
	readChan := make(chan string)

	gameProc := GameProc{
		command:  cmd,
		stdin:    stdin,
		readChan: readChan,
		started:  true,
	}

	go gameProc.startRead(stdout)

	return &gameProc
}

func (game *GameProc) startRead(stdout io.ReadCloser) {
	buf := bufio.NewReader(stdout)
	for {
		line, _, err := buf.ReadLine()
		if err != nil {
			log.Print(err)
			game.Stop()
			break
		} else {
			game.readChan <- string(line)
		}
	}
}

func (game *GameProc) Execute(cmd string) string {
	result := ""
	io.WriteString(game.stdin, fmt.Sprintf("%s\n", cmd))
stdoutloop:
	for {
		select {
		case stdout, ok := <-game.readChan:
			if !ok {
				break stdoutloop
			} else {
				result = fmt.Sprintf("%s\n%s", result, stdout)
			}
		case <-time.After(1 * time.Second):
			break stdoutloop
		}
	}
	return result
}

func (game *GameProc) Stop() {
	defer game.stdin.Close()
	defer close(game.readChan)
	err := game.command.Process.Kill()
	if err != nil {
		log.Print("Couldn't kill process")
	}
	game.started = false
	log.Print(game.started)
}
