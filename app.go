package main

import (
	"fmt"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"os"
	"strings"
	"bytes"
	"github.com/Sirupsen/logrus"
	"path"
)

func main() {
	message := os.Args[1:]
	if len(message) == 0 {
		logrus.Info("no message to send")
		return
	}
	msg := strings.Join(message, "\n")
	configObject := readConfigFile()
	token := getToken(configObject.Cropid, configObject.Secret)
	sendMsg(token, msg, configObject)
}

func getTokenUrl(corpid, corpsect string) string {
	return fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=%v&corpsecret=%v", corpid, corpsect);
}

func getToken(corpid, corpsect string) string {
	resp, err := http.Get(getTokenUrl(corpid, corpsect))
	if err != nil {
		return "";
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	var tokenObject TokenObject
	err = json.Unmarshal(body, &tokenObject)
	if err != nil {
		return ""
	}
	return tokenObject.AccessToken
}

type TokenObject struct {
	Errcode     int    `json:"errcode"`
	Errmsg      string `json:"errmsg"`
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

type ConfigObject struct {
	Cropid  string   `json:"cropid"`
	Secret  string   `json:"secret"`
	Users   []string `json:"users"`
	Toparty string   `json:"toparty"`
	Totag   string   `json:"totag"`
	AgentId int      `json:"agentid"`
}

type TextMsgObject struct {
	Touser  string      `json:"touser"`
	Toparty string      `json:"toparty"`
	Totag   string      `json:"totag"`
	Msgtype string      `json:"msgtype"`
	Agentid int         `json:"agentid"`
	Text    *TextObject `json:"text"`
	Safe    int         `json:"safe"`
}

type TextObject struct {
	Content string `json:"content"`
}

func readConfigFile() ConfigObject {
	appPath, e := os.Executable()
	if e != nil {
		logrus.Errorf("File error: %v\n", e)
		os.Exit(1)
	}

	file, e := ioutil.ReadFile(path.Dir(appPath) + "/config.json")
	if e != nil {
		logrus.Errorf("File error: %v\n", e)
		os.Exit(1)
	}
	var configObject ConfigObject
	e = json.Unmarshal(file, &configObject)
	if e != nil {
		logrus.Errorf("invalid config json data error: %v\n", e)
		os.Exit(1)
	}
	return configObject
}

func sendMsg(token, message string, configObject ConfigObject) bool {
	toUsers := strings.Join(configObject.Users, "|")

	text := &TextObject{
		Content: message,
	}
	textMsg := &TextMsgObject{
		Touser:  toUsers,
		Toparty: configObject.Toparty,
		Totag:   configObject.Totag,
		Msgtype: "text",
		Agentid: configObject.AgentId,
		Safe:    0,
		Text:    text,
	}
	jsonStr, err := json.Marshal(&textMsg)
	if err != nil {
		logrus.Errorf(err.Error())
		return false
	}
	req, err := http.NewRequest("POST", getSendMsgUrl(token), bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logrus.Errorf(err.Error())
		return false
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	logrus.Infof("response body : %v", string(body))
	return true
}

func getSendMsgUrl(token string) string {
	return fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/message/send?access_token=%v", token)
}
