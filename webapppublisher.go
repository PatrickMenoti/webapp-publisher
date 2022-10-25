package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"

	"github.com/go-git/go-git/v5"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
)

type ProjectSettings struct {
	WorkingDir   *string
	Workspace    string
	BinFormatted string
}

const (
	BINNAME       = "%azioncli"
	BINPATH       = "%s/azioncli"
	PUBLISHCMD    = "%s/azioncli webapp publish"
	WEBDEVENDPATH = "%s/azion/webdev.env"
	AZIONPATH     = "%s/azion"
)

type kv struct {
	Bucket string `json:"bucket"`
	Region string `json:"region"`
	Path   string `json:"path"`
}

func main() {

	configs := &ProjectSettings{}
	wDir, err := getworkingDir()
	if err != nil {
		log.Fatal(err)
	}

	workspace := os.Getenv("GITHUB_WORKSPACE")

	configs.WorkingDir = &wDir
	configs.Workspace = workspace

	err = downloadBin(configs)
	if err != nil {
		log.Fatal(err)
	}

	should, err := shouldInit(configs)
	if err != nil {
		log.Fatal(err)
	}

	if should {
		err = initProject(configs)
		if err != nil {
			log.Fatal(err)
		}
	}

	shouldCommit, err := shouldCommit()
	if err != nil {
		log.Fatal(err)
	}

	if shouldCommit {
		err = commitChanges(configs)
		if err != nil {
			log.Fatal(err)
		}
	}

}

func downloadBin(configs *ProjectSettings) error {

	// Create the file
	// binFormatted := fmt.Sprintf(BINNAME, *&configs.Workspace)
	// configs.BinFormatted = binFormatted
	// out, err := os.Create(configs.BinFormatted)
	// if err != nil {
	// 	return err
	// }
	// defer out.Close()

	// // Get the data
	// resp, err := http.Get("https://downloads.azion.com/linux/x86_64/azioncli")
	// if err != nil {
	// 	return err
	// }
	// defer resp.Body.Close()

	// // Check server response
	// if resp.StatusCode != http.StatusOK {
	// 	return fmt.Errorf("bad status: %s", resp.Status)
	// }

	// // Writer the body to file
	// _, err = io.Copy(out, resp.Body)
	// if err != nil {
	// 	return err
	// }

	// os.Chmod(configs.BinFormatted, 0777)
	// return nil

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

func initProject(configs *ProjectSettings) error {

	projectName := os.Getenv("PROJECT_NAME")
	projectType := os.Getenv("PROJECT_TYPE")

	// cmdString := fmt.Sprintf(BINPATH, *&configs.Workspace)
	ls := exec.Command("ls", configs.Workspace)
	// cmd := exec.Command(cmdString, "webapp", "init", "--name", projectName, "--type", projectType, "-y")

	pathWorkingDir, err := os.Getwd()
	if err != nil {
		return err
	}

	cmdString := fmt.Sprintf(BINPATH, pathWorkingDir)
	cmd := exec.Command(cmdString, "webapp", "init", "--name", projectName, "--type", projectType, "-y")

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	ls.Stdout = &out
	ls.Stderr = &stderr

	err = ls.Run()
	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		return err
	}
	fmt.Println("Result: " + out.String())

	fmt.Println("running command: ")
	fmt.Println(cmd)

	err = cmd.Run()
	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		return err
	}
	fmt.Println("Result: " + out.String())

	switch projectType {

	case "javascript":

		err = publishProject(configs)
		if err != nil {
			return err
		}

	// flareact and nextjs follow the same steps
	default:
		err := updateWebdev()
		if err != nil {
			return err
		}

		err = setupKV(configs)
		if err != nil {
			return err
		}

		err = publishProject(configs)
		if err != nil {
			return err
		}

	}
	return nil
}

func publishProject(configs *ProjectSettings) error {
	token := os.Getenv("AZION_TOKEN")

	cmdConf := exec.Command(configs.BinFormatted, "configure", "-t", token)
	err := cmdConf.Run()
	if err != nil {
		return err
	}

	cmdPublish := exec.Command(configs.BinFormatted, "webapp", "publish")
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

	key, keyPresent := os.LookupEnv("AWS_ACCESS_KEY_ID")
	secret, secretPresent := os.LookupEnv("AWS_SECRET_ACCESS_KEY")
	if !keyPresent || !secretPresent {
		return errors.New("You must provide AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY environment variables for this Project Type")
	}

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

func shouldInit(configs *ProjectSettings) (bool, error) {

	if configs == nil {
		return false, errors.New("Error creating your Project Settings")
	}

	empty, err := isDirEmpty(*configs.WorkingDir)
	if err != nil {
		return false, err
	}

	force, present := os.LookupEnv("FORCE_INIT")
	if present {
		shouldForce, err := strconv.ParseBool(force)
		if err != nil {
			return false, errors.New("You must provide either true or false for FORCE_INIT environment variable")
		}
		if shouldForce {
			return true, nil
		}
	} else {
		if !empty {
			return true, nil
		} else {
			return false, errors.New("You already have an Azion template initialized. Please, delete the azion folder, or use the FORCE_INIT environment variable, for force initialization of a new azion template!")
		}
	}

	return true, nil
}

func isDirEmpty(dir string) (bool, error) {
	f, err := os.Open(dir)
	if err != nil {
		// Dir does not exist
		if errors.Is(err, os.ErrNotExist) {
			return true, nil
		}
		return false, err
	}
	defer f.Close()

	// read in ONLY one file
	_, err = f.Readdir(1)

	// and if the file is EOF the dir is empty.
	if err == io.EOF {
		return true, nil
	}
	return false, err
}

func getworkingDir() (string, error) {
	pathWorkingDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	return pathWorkingDir, nil
}

func setupKV(configs *ProjectSettings) error {

	bucket, bucketPresent := os.LookupEnv("KV_BUCKET")
	region, regionPresent := os.LookupEnv("KV_REGION")
	path, pathPresent := os.LookupEnv("KV_PATH")
	should, shouldPresent := os.LookupEnv("SETUP_KV")

	if !shouldPresent {
		return nil
	}

	shouldSetup, err := strconv.ParseBool(should)

	if shouldSetup {
		if !bucketPresent || !regionPresent || !pathPresent {
			return errors.New("You must inform KV_BUCKET, KV_REGION and KV_PATH for this PROJECT_TYPE")
		}
	}

	var kVContents kv

	kVContents.Bucket = bucket
	kVContents.Region = region
	kVContents.Path = path

	file, err := json.MarshalIndent(kVContents, "", " ")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(*configs.WorkingDir+"/azion/kv.js", file, 0644)
	if err != nil {
		return err
	}

	return nil
}

func shouldCommit() (bool, error) {

	should, shouldPresent := os.LookupEnv("SHOULD_COMMIT")
	if shouldPresent {
		shouldCommit, err := strconv.ParseBool(should)
		if err != nil {
			return false, errors.New("You must inform either true or false for SHOULD_COMMIT")
		}
		if shouldCommit {
			_, tokenPresent := os.LookupEnv("PUSH_TOKEN")
			_, userPresent := os.LookupEnv("PUSH_USER")
			if !tokenPresent || !userPresent {
				return false, errors.New("You must inform a Github token and an User if you wish to commit changes made by webapp-publisher")
			}
			return true, nil
		}
	}

	return false, nil
}

func commitChanges(configs *ProjectSettings) error {

	r, err := git.PlainOpen(*&configs.Workspace)
	if err != nil {
		return err
	}

	pToken := os.Getenv("PUSH_TOKEN")
	pUser := os.Getenv("PUSH_USER")

	auth := &githttp.BasicAuth{
		Username: pUser,
		Password: pToken,
	}

	w, err := r.Worktree()
	if err != nil {
		return err
	}

	path := fmt.Sprintf(AZIONPATH, *configs.WorkingDir)
	w.Add(path)
	w.Commit("chore: update azion directory", &git.CommitOptions{})

	err = r.Push(&git.PushOptions{
		RemoteName: "origin",
		Auth:       auth,
	})
	//TODO verify which error NoErrAlreadyUpToDate
	if err != nil {
		return err
	}

	return nil
}
