package bll

import (
	"context"
	"fmt"
	"github.com/cloudslit/cloudslit/provider/internal/config"
	"github.com/cloudslit/cloudslit/provider/internal/schema"
	"github.com/cloudslit/cloudslit/provider/pkg/errors"
	"github.com/cloudslit/cloudslit/provider/pkg/logger"
	"github.com/cloudslit/cloudslit/provider/pkg/p2p"
	"github.com/cloudslit/cloudslit/provider/pkg/util/json"
	"github.com/cloudslit/cloudslit/provider/pkg/web3/w3s"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// Pubsub
type Pubsub struct {
}

func NewPubsub() *Pubsub {
	return &Pubsub{}
}

// eventhandle
func (a *Pubsub) StartPubsubHandler(ctx context.Context, ps *p2p.PubSub, p *p2p.P2P) {
	server := a.NewServerInfo()
	refreshticker := time.NewTicker(10 * time.Second)
	defer refreshticker.Stop()
	for {
		select {
		case msg := <-ps.Inbound:
			// Print the recieved messages to the message box
			err := a.ReceiveHandle(ctx, ps, msg.Message)
			if err != nil {
				logger.Errorf("Receive Msg Handle Err:%s", msg)
			}

		case <-refreshticker.C:
			// Timing publish
			a.nodePublish(ps, server)
			//msg := `{"type":"order","data":{"server_cid":"bafybeia67xlj2w56ps7x5youglzyisqbb2syymyqditout6qfy77rrcxbq","wallet":"0x1B4b827703dc3545089fcee70F0e6e732BFF4413","uuid":"cf636cbe-cfd4-44c0-8a9d-bec110382e6a","port":0}}`
			//err := a.ReceiveHandle(ctx, ps, msg)
			//if err != nil {
			//	logger.Errorf("Receive Msg Handle Err:%s", err)
			//}

		case log := <-ps.Logs:
			// Add the log to the message box
			logger.Infof("PubSub Log:%s", log)
		}
	}
}

func (a *Pubsub) ReceiveHandle(ctx context.Context, ps *p2p.PubSub, msg string) error {
	logger.Infof("Received Msg:%s", msg)
	var pss schema.PsMessage
	err := json.Unmarshal([]byte(msg), &pss)
	if err != nil {
		return errors.WithStack(err)
	}
	if pss.Type == "order" {
		err := a.orderReceive(ctx, ps, &pss)
		if err != nil {
			return err
		}
	}
	return nil
}

// orderReceive 接收订单信息
func (a *Pubsub) orderReceive(ctx context.Context, ps *p2p.PubSub, pss *schema.PsMessage) error {
	order, err := pss.ToNodeOrder()
	if err != nil {
		return err
	}
	if order.Wallet != config.C.Web3.Account {
		return fmt.Errorf("wallet 异常，expect:%s, get:%s", config.C.Web3.Account, order.Wallet)
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	data, err := w3s.Get(ctx, order.ServerCid)
	if err != nil {
		return errors.WithStack(err)
	}
	pc, err := schema.ToProviderContent(data)
	if err != nil {
		return errors.WithStack(err)
	}
	// 预制端口
	port, err := verifyPort(config.C.App.LocalPort + 1)
	if err != nil {
		return errors.WithStack(err)
	}
	// 监听端口
	p := NewProvider()
	l, err := p.Listen(ctx, port, pc)
	if err != nil {
		return errors.WithStack(err)
	}
	go p.Handle(ctx, l)
	order.Port = port
	pub := &schema.PsMessage{
		Type: "order",
		Data: order.String(),
	}
	str := pub.String()
	ps.Outbound <- str
	return nil
}

// nodePublish 节点发布自身信息
func (a *Pubsub) nodePublish(ps *p2p.PubSub, server *schema.NodeInfo) {
	logger.Infof("I'm:%s ", ps.Host.Host.ID().String())
	pub := &schema.PsMessage{
		Type: "node",
		Data: server,
	}
	str := pub.String()
	ps.Outbound <- str
}

func (a *Pubsub) NewServerInfo() *schema.NodeInfo {
	result := schema.NodeInfo{
		PeerId: config.C.Web3.Account,
		Addr:   config.C.App.LocalAddr,
		Port:   config.C.App.LocalPort,
		Price:  config.C.Web3.Price,
		Type:   schema.NodeTypeProvider,
	}
	trace, err := a.GetCftrace()
	if err != nil {
		logger.Warnf("Request Cfssl CDN Trace Error:%s", err)
	} else {
		result.MetaData = schema.MetaData{
			Ip:   trace.Ip,
			Loc:  trace.Loc,
			Colo: trace.Colo,
		}
	}
	return &result
}

type CfTrace struct {
	Ip   string `json:"ip"`
	Loc  string `json:"loc"`
	Colo string `json:"colo"`
}

func (a *Pubsub) GetCftrace() (*CfTrace, error) {
	resp, err := http.Get("https://www.cloudflare.com/cdn-cgi/trace")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	result := new(CfTrace)
	sb := strings.Split(string(b), "\n")
	for _, item := range sb {
		is := strings.Split(item, "=")
		if is[0] == "ip" {
			result.Ip = is[1]
		}
		if is[0] == "loc" {
			result.Loc = is[1]
		}
		if is[0] == "colo" {
			result.Colo = is[1]
		}
	}
	return result, err
}