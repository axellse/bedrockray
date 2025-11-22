package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
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

	if _, err := os.Stat(mcdir); errors.Is(err, os.ErrNotExist) {
  		fmt.Println(mcdir, "does not exist, creating...")
		err := os.MkdirAll(mcdir, 0755)
		if err != nil {
			fmt.Println("FATAL: failed creating mcdir:", err)
			os.Exit(1)
		}
	}

	fmt.Println("mcdir alright, now checking for server", mcid)
	serverPath := filepath.Join(mcdir, "bdray-server-" + mcid)
	if _, err := os.Stat(serverPath); errors.Is(err, os.ErrNotExist) {
		fmt.Println("server does not exist!")
		fmt.Println("preparing to create server", mcid)
		os.Mkdir(serverPath, 0755)

		clean := func ()  {
			err := os.Remove(serverPath)
			if err != nil {
				fmt.Println("ion ever care anymore this servers completly fucked💀🥀")
			}
			os.Exit(1)
		}

		fmt.Println("now reading server zip...")
		ba, err :=  os.ReadFile("./server.zip")
		if err != nil {
			fmt.Println("FATAL: could not open server zip!")
			clean()
		}

		serverZipPath := filepath.Join(serverPath, "server.zip")
		err = os.WriteFile(serverZipPath, ba, 0755)
		if err != nil {
			fmt.Println("FATAL: could not write server zip!")
			clean()
		}

		cmd := exec.Command("tar", "-xf", "server.zip") //tar is also on windows
		cmd.Path = serverPath

		err = cmd.Run()
		if err !=  nil {
			fmt.Println("FATAL: failed unzipping server")
			clean()
		}

		err = os.Remove(serverZipPath)
		if err !=  nil {
			fmt.Println("failed removing server zip file, fine!")
		}
	}

	fmt.Println("pulled server ok, now writing server.properties")
	ba, err := os.ReadFile("./server.properties")
	if err != nil {
		fmt.Println("FATAL: could not read server.properties")
	}

	err = os.WriteFile(filepath.Join(serverPath, "server.properties"), ba, 0775)
	if err != nil {
		fmt.Println("FATAL: could not write server.properties")
	}

	fmt.Println("server alright, now starting server.")
	serverBinPath := filepath.Join(mcdir, "bedrock-server")

	cmd := exec.Command(serverBinPath)
	err = cmd.Start()
	if err != nil {
		fmt.Println("FATAL: server exited.")
		os.Exit(1)
	}
}