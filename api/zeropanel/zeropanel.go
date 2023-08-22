package zeropanel

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/AikoPanel/Aiko-Server/api"
)

// APIClient create a api client to the panel.
type APIClient struct {
	client              *resty.Client
	APIHost             string
	NodeID              int
	Key                 string
	SpeedLimit          float64
	DeviceLimit         int
	DisableCustomConfig bool
	LocalRuleList       []api.DetectRule
	LastReportOnline    map[int]int
	access              sync.Mutex
}

// New creat a api instance
func New(apiConfig *api.Config) *APIClient {

	client := resty.New()
	client.SetRetryCount(3)
	if apiConfig.Timeout > 0 {
		client.SetTimeout(time.Duration(apiConfig.Timeout) * time.Second)
	} else {
		client.SetTimeout(5 * time.Second)
	}
	client.OnError(func(req *resty.Request, err error) {
		if v, ok := err.(*resty.ResponseError); ok {
			// v.Response contains the last response from the server
			// v.Err contains the original error
			log.Print(v.Err)
		}
	})
	client.SetBaseURL(apiConfig.APIHost)
	// Create Key for each requests
	client.SetQueryParam("key", apiConfig.Key)
	// Add support for muKey
	client.SetQueryParam("muKey", apiConfig.Key)
	// Read local rule list
	localRuleList := readLocalRuleList(apiConfig.RuleListPath)

	return &APIClient{
		client:              client,
		NodeID:              apiConfig.NodeID,
		Key:                 apiConfig.Key,
		APIHost:             apiConfig.APIHost,
		SpeedLimit:          apiConfig.SpeedLimit,
		DeviceLimit:         apiConfig.DeviceLimit,
		LocalRuleList:       localRuleList,
		DisableCustomConfig: apiConfig.DisableCustomConfig,
		LastReportOnline:    make(map[int]int),
	}
}

// readLocalRuleList reads the local rule list file
func readLocalRuleList(path string) (LocalRuleList []api.DetectRule) {

	LocalRuleList = make([]api.DetectRule, 0)
	if path != "" {
		// open the file
		file, err := os.Open(path)

		// handle errors while opening
		if err != nil {
			log.Printf("Error when opening file: %s", err)
			return LocalRuleList
		}

		fileScanner := bufio.NewScanner(file)

		// read line by line
		for fileScanner.Scan() {
			LocalRuleList = append(LocalRuleList, api.DetectRule{
				ID:      -1,
				Pattern: regexp.MustCompile(fileScanner.Text()),
			})
		}
		// handle first encountered error while reading
		if err := fileScanner.Err(); err != nil {
			log.Fatalf("Error while reading file: %s", err)
			return
		}

		file.Close()
	}

	return LocalRuleList
}

// Describe return a description of the client
func (c *APIClient) Describe() api.ClientInfo {
	return api.ClientInfo{APIHost: c.APIHost, NodeID: c.NodeID, Key: c.Key}
}

// Debug set the client debug for client
func (c *APIClient) Debug() {
	c.client.SetDebug(true)
}

func (c *APIClient) assembleURL(path string) string {
	return c.APIHost + path
}

func (c *APIClient) parseResponse(res *resty.Response, path string, err error) (*Response, error) {
	if err != nil {
		return nil, fmt.Errorf("request %s failed: %s", c.assembleURL(path), err)
	}

	if res.StatusCode() > 400 {
		body := res.Body()
		return nil, fmt.Errorf("request %s failed: %s, %s", c.assembleURL(path), string(body), err)
	}
	response := res.Result().(*Response)

	if response.Ret != 1 {
		res, _ := json.Marshal(&response)
		return nil, fmt.Errorf("ret %s invalid", string(res))
	}
	return response, nil
}

// GetNodeInfo will pull NodeInfo Config from zeropanel
func (c *APIClient) GetNodeInfo() (nodeInfo *api.NodeInfo, err error) {
	path := fmt.Sprintf("/api/v1/server/nodes/%d/info", c.NodeID)
	res, err := c.client.R().
		SetResult(&Response{}).
		ForceContentType("application/json").
		Get(path)

	response, err := c.parseResponse(res, path, err)
	if err != nil {
		return nil, err
	}

	nodeInfoResponse := new(NodeInfoResponse)

	if err := json.Unmarshal(response.Data, nodeInfoResponse); err != nil {
		return nil, fmt.Errorf("unmarshal %s failed: %s", reflect.TypeOf(nodeInfoResponse), err)
	}

	// New zeropanel API
	disableCustomConfig := c.DisableCustomConfig

	if !disableCustomConfig {
		nodeInfo, err = c.ParseZeroPanelNodeInfo(nodeInfoResponse)
		if err != nil {
			res, _ := json.Marshal(nodeInfoResponse)
			return nil, fmt.Errorf("Parse node info failed: %s, \nError: %s, \nPlease check the doc of custom_config for help: https://zeropanel.github.io/XrayR-doc/dui-jie-sspanel/sspanel/sspanel_custom_config", string(res), err)
		}
	}

	if err != nil {
		res, _ := json.Marshal(nodeInfoResponse)
		return nil, fmt.Errorf("Parse node info failed: %s, \nError: %s", string(res), err)
	}

	return nodeInfo, nil
}

// GetUserList will pull user form zeropanel
func (c *APIClient) GetUserList() (UserList *[]api.UserInfo, err error) {
	path := "/api/v1/server/users"
	res, err := c.client.R().
		SetQueryParam("node_id", strconv.Itoa(c.NodeID)).
		SetResult(&Response{}).
		ForceContentType("application/json").
		Get(path)

	response, err := c.parseResponse(res, path, err)
	if err != nil {
		return nil, err
	}

	userListResponse := new([]UserResponse)

	if err := json.Unmarshal(response.Data, userListResponse); err != nil {
		return nil, fmt.Errorf("unmarshal %s failed: %s", reflect.TypeOf(userListResponse), err)
	}
	userList, err := c.ParseUserListResponse(userListResponse)
	if err != nil {
		res, _ := json.Marshal(userListResponse)
		return nil, fmt.Errorf("parse user list failed: %s", string(res))
	}
	return userList, nil
}

// ReportNodeStatus reports the node status to the zeropanel
func (c *APIClient) ReportNodeStatus(nodeStatus *api.NodeStatus) (err error) {
	path := fmt.Sprintf("/api/v1/server/nodes/%d/info", c.NodeID)
	systemload := SystemLoad{
		Uptime: strconv.FormatUint(nodeStatus.Uptime, 10),
		Load:   fmt.Sprintf("%.2f %.2f %.2f", nodeStatus.CPU/100, nodeStatus.Mem/100, nodeStatus.Disk/100),
	}

	res, err := c.client.R().
		SetBody(systemload).
		SetResult(&Response{}).
		ForceContentType("application/json").
		Post(path)

	_, err = c.parseResponse(res, path, err)
	if err != nil {
		return err
	}

	return nil
}

// ReportNodeOnlineUsers reports online user ip
func (c *APIClient) ReportNodeOnlineUsers(onlineUserList *[]api.OnlineUser) error {
	c.access.Lock()
	defer c.access.Unlock()

	reportOnline := make(map[int]int)
	data := make([]OnlineUser, len(*onlineUserList))
	for i, user := range *onlineUserList {
		data[i] = OnlineUser{UID: user.UID, IP: user.IP}
		if _, ok := reportOnline[user.UID]; ok {
			reportOnline[user.UID]++
		} else {
			reportOnline[user.UID] = 1
		}
	}
	c.LastReportOnline = reportOnline // Update LastReportOnline

	postData := &PostData{Data: data}
	path := fmt.Sprintf("/api/v1/server/users/aliveip")
	res, err := c.client.R().
		SetQueryParam("node_id", strconv.Itoa(c.NodeID)).
		SetBody(postData).
		SetResult(&Response{}).
		ForceContentType("application/json").
		Post(path)

	_, err = c.parseResponse(res, path, err)
	if err != nil {
		return err
	}

	return nil
}

// ReportUserTraffic reports the user traffic
func (c *APIClient) ReportUserTraffic(userTraffic *[]api.UserTraffic) error {

	data := make([]UserTraffic, len(*userTraffic))
	for i, traffic := range *userTraffic {
		data[i] = UserTraffic{
			UID:      traffic.UID,
			Upload:   traffic.Upload,
			Download: traffic.Download}
	}
	postData := &PostData{Data: data}
	path := "/api/v1/server/users/traffic"
	res, err := c.client.R().
		SetQueryParam("node_id", strconv.Itoa(c.NodeID)).
		SetBody(postData).
		SetResult(&Response{}).
		ForceContentType("application/json").
		Post(path)
	_, err = c.parseResponse(res, path, err)
	if err != nil {
		return err
	}

	return nil
}

// GetNodeRule will pull the audit rule form zeropanel
func (c *APIClient) GetNodeRule() (*[]api.DetectRule, error) {
	ruleList := c.LocalRuleList
	path := "/api/v1/server/func/detect_rules"
	res, err := c.client.R().
		SetResult(&Response{}).
		ForceContentType("application/json").
		Get(path)

	response, err := c.parseResponse(res, path, err)
	if err != nil {
		return nil, err
	}

	ruleListResponse := new([]RuleItem)

	if err := json.Unmarshal(response.Data, ruleListResponse); err != nil {
		return nil, fmt.Errorf("unmarshal %s failed: %s", reflect.TypeOf(ruleListResponse), err)
	}

	for _, r := range *ruleListResponse {
		ruleList = append(ruleList, api.DetectRule{
			ID:      r.ID,
			Pattern: regexp.MustCompile(r.Content),
		})
	}
	return &ruleList, nil
}

// ReportIllegal reports the user illegal behaviors
func (c *APIClient) ReportIllegal(detectResultList *[]api.DetectResult) error {

	data := make([]IllegalItem, len(*detectResultList))
	for i, r := range *detectResultList {
		data[i] = IllegalItem{
			ID:  r.RuleID,
			UID: r.UID,
		}
	}
	postData := &PostData{Data: data}
	path := "/api/v1/server/users/detectlog"
	res, err := c.client.R().
		SetQueryParam("node_id", strconv.Itoa(c.NodeID)).
		SetBody(postData).
		SetResult(&Response{}).
		ForceContentType("application/json").
		Post(path)
	_, err = c.parseResponse(res, path, err)
	if err != nil {
		return err
	}
	return nil
}

// ParseUserListResponse parse the response for the given nodeinfo format
func (c *APIClient) ParseUserListResponse(userInfoResponse *[]UserResponse) (*[]api.UserInfo, error) {
	c.access.Lock()
	// Clear Last report log
	defer func() {
		c.LastReportOnline = make(map[int]int)
		c.access.Unlock()
	}()

	var deviceLimit, localDeviceLimit int = 0, 0
	var speedlimit uint64 = 0
	var userList []api.UserInfo
	for _, user := range *userInfoResponse {
		if c.DeviceLimit > 0 {
			deviceLimit = c.DeviceLimit
		} else {
			deviceLimit = user.DeviceLimit
		}

		// If there is still device available, add the user
		if deviceLimit > 0 && user.AliveIP > 0 {
			lastOnline := 0
			if v, ok := c.LastReportOnline[user.ID]; ok {
				lastOnline = v
			}
			// If there are any available device.
			if localDeviceLimit = deviceLimit - user.AliveIP + lastOnline; localDeviceLimit > 0 {
				deviceLimit = localDeviceLimit
				// If this backend server has reported any user in the last reporting period.
			} else if lastOnline > 0 {
				deviceLimit = lastOnline
				// Remove this user.
			} else {
				continue
			}
		}

		if c.SpeedLimit > 0 {
			speedlimit = uint64((c.SpeedLimit * 1000000) / 8)
		} else {
			speedlimit = uint64((user.SpeedLimit * 1000000) / 8)
		}
		userList = append(userList, api.UserInfo{
			UID:         user.ID,
			Email:       user.Email,
			UUID:        user.UUID,
			Passwd:      user.Passwd,
			SpeedLimit:  speedlimit,
			DeviceLimit: deviceLimit,
			Port:        user.Port,
		})
	}

	return &userList, nil
}

// ParsezeropanelNodeInfo parse the response for the given nodeinfor format
func (c *APIClient) ParseZeroPanelNodeInfo(nodeInfoResponse *NodeInfoResponse) (*api.NodeInfo, error) {

	var speedlimit uint64 = 0
	var transportProtocol string
	var nodetype string
	nodetype = nodeInfoResponse.NodeType

	nodeConfig := new(CustomConfig)
	json.Unmarshal(nodeInfoResponse.CustomConfig, nodeConfig)
	if c.SpeedLimit > 0 {
		speedlimit = uint64((c.SpeedLimit * 1000000) / 8)
	} else {
		speedlimit = uint64((nodeInfoResponse.SpeedLimit * 1000000) / 8)
	}

	parsedPort, err := strconv.ParseInt(nodeConfig.OffsetPortNode, 10, 32)
	if err != nil {
		return nil, err
	}
	port := uint32(parsedPort)

	if nodetype == "Shadowsocks" {
		transportProtocol = "tcp"
	}

	if nodetype == "Vmess" {
		transportProtocol = nodeConfig.Network
	}

	if nodetype == "Vless" {
		transportProtocol = nodeConfig.Network
	}

	if nodetype == "Trojan" {
		transportProtocol = nodeConfig.Network
	}

	// Create GeneralNodeInfo
	nodeinfo := &api.NodeInfo{
		NodeType:          nodetype,
		NodeID:            c.NodeID,
		Port:              port,
		SpeedLimit:        speedlimit,
		TransportProtocol: transportProtocol,
		Host:              nodeConfig.Host,
		Path:              nodeConfig.Path,
		Security:          nodeConfig.Security,
		VlessFlow:         nodeConfig.Flow,
		CypherMethod:      nodeConfig.SSEncryption,
		ServerKey:         nodeConfig.ServerPsk,
		ServiceName:       nodeConfig.ServiceName,
		Header:            nodeConfig.Header,
		RealityConfig:     nodeConfig.RealityConfig,
	}

	return nodeinfo, nil
}
