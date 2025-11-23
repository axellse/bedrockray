package main

import (
	"bufio"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

type ServerController struct {
	ServerBinaryPath string
	ServerProcessLogs [][]byte
	ServerProcessStdin io.WriteCloser
	ServerProcess *exec.Cmd
	Active bool
}

//Starts and runs the server. Blocks until the server is stopped or encounters a catastrophic error
func (s *ServerController) Start() {
	if s.Active {
		fmt.Println("already active, cant start.")
		return
	}
	fmt.Println("now starting server...")
	cmd := exec.Command(s.ServerBinaryPath)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		fmt.Println("FATAL: could not create server stdin pipe:", err)
		os.Exit(1)
	}

	s.ServerProcessStdin = stdin
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println("FATAL: could not create server stdout pipe:", err)
		os.Exit(1)
	}

	go func () {
		rd := bufio.NewReader(stdout)
		for {
			ba, err := rd.ReadBytes('\n')
			if err != nil {
				break
			}
			s.ServerProcessLogs = append(s.ServerProcessLogs, ba)
		}
	}()
	s.ServerProcess = cmd
	err = cmd.Start()
	if err != nil {
		fmt.Println("server start fail:", err)
		return
	}

	s.Active = true
	err = cmd.Wait()
	if err != nil {
		fmt.Println("server exited with error:", err)
		return
	}
	s.Active = false
}

//Stops the server
func (s *ServerController) Stop() {
	if !s.Active {
		fmt.Println("server not active, cant stop.")
		return
	}

	fmt.Println("stopping server...")
	s.ServerProcessStdin.Write([]byte("stop\n"))
	time.Sleep(5 * time.Second)
	if s.Active {
		fmt.Println("server not stopped, giving it another 15 sec...")
		time.Sleep(15 * time.Second)
		if s.Active {
			fmt.Println("alright, now forcefully killing...")
			s.ServerProcess.Process.Kill()
		}
	}
}

func main() {
	fmt.Println("\033[32m _                  _                         _                           ")
	fmt.Println("| |__     ___    __| |  _ __    ___     ___  | | __  _ __    __ _   _   _ ")
	fmt.Println("| '_ \\   / _ \\  / _` | | '__|  / _ \\   / __| | |/ / | '__|  / _` | | | | |")
	fmt.Println("| |_) | |  __/ | (_| | | |    | (_) | | (__  |   <  | |    | (_| | | |_| |")
	fmt.Println("|_.__/   \\___|  \\__,_| |_|     \\___/   \\___| |_|\\_\\ |_|     \\__,_|  \\__, |")
	fmt.Println("                                                                    |___/ \033[0m")
	fmt.Println("Started bedrockray!")
	mcdir := os.Getenv("MCDIR")
	mcid := os.Getenv("MCID")

	if mcdir == "" {
		fmt.Println("FATAL: no mcdir specified")
		os.Exit(1)
	}
	if mcid == "" {
		fmt.Println("FATAL: no mcid specified")
		os.Exit(1)
	}

	_, err := os.Stat(mcdir)
	if errors.Is(err, os.ErrNotExist) {
  		fmt.Println(mcdir, "does not exist, creating...")
		err := os.MkdirAll(mcdir, 0777)
		if err != nil {
			fmt.Println("FATAL: failed creating mcdir:", err)
			os.Exit(1)
		}
	} else if err != nil {
		fmt.Println("FATAL: other mcdir error", err)
		os.Exit(1)
	}

	fmt.Println("mcdir alright, now checking for server", mcid)
	serverPath := filepath.Join(mcdir, "bdray-server-" + mcid)
	if _, err := os.Stat(serverPath); errors.Is(err, os.ErrNotExist) {
		fmt.Println("server does not exist!")
		fmt.Println("creating server", serverPath)
		
		if os.Mkdir(serverPath, 0777) != nil {
			fmt.Println("FATAL: failed creating dir:", err)
			os.Exit(1)
		}		

		mcserverdl := os.Getenv("MCSERVERDL")
		if mcserverdl == "" && runtime.GOOS == "windows" {
			mcserverdl = "https://www.minecraft.net/bedrockdedicatedserver/bin-win/bedrock-server-1.21.124.2.zip"
		} else {
			mcserverdl = "https://www.minecraft.net/bedrockdedicatedserver/bin-linux/bedrock-server-1.21.124.2.zip"
		}
		fmt.Println("now pulling server from", mcserverdl)

		clean := func ()  {
			err := os.RemoveAll(serverPath)
			if err != nil {
				fmt.Println("remove server error:", err)
			}
			os.Exit(1)
		}

		client := &http.Client{
			Transport: &http.Transport{
				TLSNextProto: map[string]func(string, *tls.Conn) http.RoundTripper{}, // disable HTTP/2
			},
		}
		req, err := http.NewRequest("GET", mcserverdl, nil)
		if err != nil {
			fmt.Println("FATAL: could not prepare server download: ", err)
			clean()
		}
		req.Header.Set("User-Agent", "Mozilla/5.0")
		resp, err := client.Do(req)

		if err != nil {
			fmt.Println("FATAL: could not start server download: ", err)
			if resp != nil {
				bd, _ := io.ReadAll(resp.Body)
				fmt.Println(string(bd))
			}
			clean()
		}

		fmt.Println("download start ok, now writing to file...")

		serverZipPath := filepath.Join(serverPath, "server.zip")
		serverZip, err := os.Create(serverZipPath)
		if err != nil {
			fmt.Println("FATAL: could not open server zip:", err)
			clean()
		}

		_, err = io.Copy(serverZip, resp.Body)
		if err != nil {
			fmt.Println("FATAL: could not unzip server!")
			clean()
		}
		err = serverZip.Close()
		if err != nil {
			fmt.Println("close serverzip error, maybe fine?")
		}

		fmt.Println("download success, now unzipping")
		cmd := exec.Command("tar", "-xf", "server.zip") //tar is also on windows
		cmd.Dir = serverPath

		ba, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Println("FATAL: failed unzipping server:", err)
			fmt.Println(string(ba))
			clean()
		}

		err = os.Remove(serverZipPath)
		if err !=  nil {
			fmt.Println("failed removing server zip file, fine!")
		}

		fmt.Println("server now created!")
	}
	fmt.Println("server ready.")

	serverBinPath := filepath.Join(serverPath, "bedrock_server")
	if runtime.GOOS == "windows" {
		serverBinPath = filepath.Join(serverPath, "bedrock_server.exe")
	}

	sc := ServerController{
		ServerBinaryPath: serverBinPath,
	}
	go StartDashboard(os.Getenv("RAY_PORT"), sc)
	sc.Start()
}