package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
)

type HttpError struct {
	StatusCode int    `json:"statusCode"`
	Msg        string `json:"msg"`
	Key        string `json:"key"`
}

func (e *HttpError) Error() string {
	return e.Msg
}

func CreateErrorWithMsg(status int, key string, msg string) error {
	return &HttpError{StatusCode: status, Msg: msg, Key: key}
}
func CreateError(status int, key string) error {
	return &HttpError{StatusCode: status, Msg: key, Key: key}
}

type SSHOptions struct {
	Host string `json:"host"`
	Port int    `json:"port"`
	User string `json:"user"`
}

const (
	// KataGoBin the bin file path
	KataGoBin string = "/content/katago"
	// KataGoWeightFile the default weight file
	KataGoWeightFile string = "/content/weight.bin.gz"
	// KataGoConfigFile the default config file
	KataGoConfigFile string = "/content/katago-colab-analyze/config/analysis_example.cfg"
	// KataGoChangeConfigScript changes the config
	KataGoChangeConfigScript string = "/content/katago-colab-analyze/scripts/change_config.sh"
)

func main() {
	args := os.Args[1:]
	if len(args) < 2 {
		log.Printf("ERROR usage: colab-katago SSH_INFO_FILE_PATH USER_PASSWORD")
		return
	}
	fileId := args[0]
	userpassword := args[1]
	var newConfig *string = nil
	if len(args) >= 3 {
		newConfig = &args[2]
	}
	log.Printf("INFO using file ID: %s password: %s\n", fileId, userpassword)
	//sshJSONURL := "https://drive.google.com/uc?id=" + fileId
	//response, err := DoHTTPRequest("GET", sshJSONURL, nil, nil)
	/*if err != nil {
		log.Printf("ERROR error requestting url: %s, err: %+v\n", sshJSONURL, err)
		return
	}*/
	//log.Printf("ssh options\n%s", response)
	sshoptions := SSHOptions{}
	dat, err := os.ReadFile(fileId)
    	if err != nil {
		log.Printf("ERROR failed reading file: %s\n", fileId)
		return
	}
    	log.Printf(string(dat))
	// parse json
	err = json.Unmarshal([]byte(string(dat)), &sshoptions)
	if err != nil {
		log.Printf("ERROR failed parsing json: %s\n", string(dat))
		return
	}

	config := &ssh.ClientConfig{
		Timeout:         30 * time.Second,
		User:            sshoptions.User,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	config.Auth = []ssh.AuthMethod{ssh.Password(userpassword)}

	addr := fmt.Sprintf("%s:%d", sshoptions.Host, sshoptions.Port)
	sshClient, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		log.Fatal("failed to create ssh client", err)
		return
	}
	defer sshClient.Close()

	configFile := KataGoConfigFile
	if newConfig != nil {
		// start the sesssion to do it
		session, err := sshClient.NewSession()
		if err != nil {
			log.Fatal("failed to create ssh session", err)
			return
		}
		defer session.Close()

		cmd := fmt.Sprintf("%s %s", KataGoChangeConfigScript, *newConfig)
		log.Printf("DEBUG running commad:%s\n", cmd)
		configFile = fmt.Sprintf("/content/gtp_colab_%s.cfg", *newConfig)
		session.Run(cmd)

	}

	session, err := sshClient.NewSession()
	if err != nil {
		log.Fatal("failed to create ssh session", err)
		return
	}

	defer session.Close()
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	session.Stdin = os.Stdin

	cmd := fmt.Sprintf("%s analysis -model %s -config %s", KataGoBin, KataGoWeightFile, configFile)
	log.Printf("DEBUG running commad:%s\n", cmd)
	session.Run(cmd)
}
