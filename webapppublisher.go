package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
)

const (
	BINNAME       = "azioncli"
	BINPATH       = "%s/azioncli"
	PUBLISHCMD    = "%s/azioncli webapp publish"
	WEBDEVENDPATH = "%s/azion/webdev.env"
)

func main() {

	err := downloadBin()
	if err != nil {
		fmt.Println(err)
	}
	err = initProject()
	if err != nil {
		fmt.Println(err)
	}
}

func downloadBin() error {

	// Create the file
	out, err := os.Create(BINNAME)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get("https://downloads.azion.com/linux/x86_64/azioncli")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	os.Chmod(BINNAME, 0777)
	return nil
}

func initProject() error {

	projectName := os.Getenv("PROJECT_NAME")
	projectType := os.Getenv("PROJECT_TYPE")

	pathWorkingDir, err := os.Getwd()
	if err != nil {
		return err
	}

	cmdString := fmt.Sprintf(BINPATH, pathWorkingDir)
	cmd := exec.Command(cmdString, "webapp", "init", "--name", projectName, "--type", projectType, "-y")

	fmt.Println("Running command: ")
	fmt.Println(cmd)

	err = cmd.Run()
	if err != nil {
		return err
	}

	switch projectType {
	case "nextjs":
		err := updateWebdev()
		if err != nil {
			fmt.Println(err)
		}

		err = publishProject()
		if err != nil {
			fmt.Println(err)
		}

	case "flareact":
		err := updateWebdev()
		if err != nil {
			fmt.Println(err)
		}

		err = publishProject()
		if err != nil {
			fmt.Println(err)
		}

	default:
		err = publishProject()
		if err != nil {
			fmt.Println(err)
		}
	}

	return nil
}

func publishProject() error {
	token := os.Getenv("AZION_TOKEN")

	pathWorkingDir, err := os.Getwd()
	if err != nil {
		return err
	}

	cmdString := fmt.Sprintf(BINPATH, pathWorkingDir)
	cmdConf := exec.Command(cmdString, "configure", "-t", token)
	err = cmdConf.Run()
	if err != nil {
		return err
	}

	cmdPublish := exec.Command(cmdString, "webapp", "publish")
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmdPublish.Stdout = &out
	cmdPublish.Stderr = &stderr

	fmt.Println("running command: ")
	fmt.Println(cmdPublish)

	err = cmdPublish.Run()
	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		return err
	}
	fmt.Println("Result: " + out.String())

	return nil
}

func updateWebdev() error {

	key := os.Getenv("AWS_ACCESS_KEY_ID")
	secret := os.Getenv("AWS_SECRET_ACCESS_KEY")

	fileContent := ""
	fileContent += "AWS_ACCESS_KEY_ID=" + key + "\n" + "AWS_SECRET_ACCESS_KEY=" + secret

	pathWorkingDir, err := os.Getwd()
	if err != nil {
		return err
	}

	path := fmt.Sprintf(WEBDEVENDPATH, pathWorkingDir)

	err = ioutil.WriteFile(path, []byte(fileContent), 0644)
	if err != nil {
		return err
	}

	return nil
}
