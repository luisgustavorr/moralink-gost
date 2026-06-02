package apimanagers

import (
	pb "MoraLinkGOst/modules/proto/agentpb"
	"MoraLinkGOst/modules/utils"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type requestInfo struct {
	url     string
	method  string
	payload string
	token   string
}
type tokenReturn struct {
	Token string
}

var ClientToken string
var API_TokenGetter *pb.APITokenGetter

func Request(infos requestInfo, customKeys []string, customValues []string) (bodyResp []byte, reasonFailed error) {

	url := infos.url
	method := infos.method

	payload := strings.NewReader(infos.payload)
	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("User-Agent", "insomnia/2023.5.8")

	req.Header.Add("Authorization", infos.token)
	for i, v := range customKeys {
		req.Header.Add(v, customValues[i])
	}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return body, err
	}
	return body, nil
}
func DecideWhoActs(ao pb.APIOwner, c *pb.APITokenGetter) (*utils.DbInfos, error) {
	var db = &utils.DbInfos{}
	switch ao {
	case pb.APIOwner_FRONTSYS:
		connectFrontsys(c, db)
	case pb.APIOwner_TRAY:
		connectTray(c, db)
	}
	return db, nil
}
