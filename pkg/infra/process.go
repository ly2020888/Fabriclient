package infra

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const burst int = 65535

var Temporary string

type Message struct {
	PassWord string   `json:"PassWord"`
	Args     []string `json:"Args"`
}

func Process(configPath string, logger *log.Logger) error {

	config, err := LoadConfig(configPath)
	if err != nil {
		return err
	}
	crypto, err := config.LoadCrypto()
	if err != nil {
		return err
	}
	raw := make(chan *Elements, burst)
	signed := make([]chan *Elements, len(config.Endorsers))
	processed := make(chan *Elements, burst)
	envs := make(chan *Elements, burst)
	blockCh := make(chan *AddressedBlock, burst)
	errorCh := make(chan error, burst)
	assembler := &Assembler{Signer: crypto}
	blockCollector, err := NewBlockCollector(config.CommitThreshold, len(config.Committers))
	if err != nil {
		return errors.Wrap(err, "failed to create block collector")
	}
	for i := 0; i < len(config.Endorsers); i++ {
		signed[i] = make(chan *Elements, burst)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for i := 0; i < 5; i++ {
		go assembler.StartSigner(ctx, raw, signed, errorCh)
		go assembler.StartIntegrator(ctx, processed, envs, errorCh)
	}

	proposers, err := CreateProposers(config.NumOfConn, config.Endorsers, logger)
	if err != nil {
		return err
	}
	proposers.Start(ctx, signed, processed, config)

	broadcaster, err := CreateBroadcasters(ctx, config.NumOfConn, config.Orderer, logger)
	if err != nil {
		return err
	}
	broadcaster.Start(ctx, envs, errorCh)

	observers, err := CreateObservers(ctx, config.Channel, config.Shardings, config.Committers, crypto, logger)
	if err != nil {
		return err
	}

	start := time.Now()

	go blockCollector.Start(ctx, blockCh, time.Now(), true)
	go observers.Start(errorCh, blockCh, start)

	go serve(&config, crypto, raw, errorCh, logger)

	for {
		select {
		case err = <-errorCh:
			logger.Errorf("error:%v", err)
		}
	}
}

func serve(config *Config, crypto *Crypto, raw chan *Elements, errorCh chan error, logger *log.Logger) {
	r := gin.Default()
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = io.Discard
	logger.Infof("已经连接至Fabic网络, 等待指令...")

	r.POST("/invoke", func(ctx *gin.Context) {
		var msg Message
		if err := ctx.ShouldBindJSON(&msg); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Error binding JSON"})
			return
		}

		if Temporary == MD5(msg.PassWord) {
			go StartCreateProposal(msg.Args, config, crypto, raw, errorCh)
			logger.Infof("Fabric调用合约参数为:%v", msg.Args)
			ctx.JSON(http.StatusOK, gin.H{"message": "智能合约调用成功"})
		} else {
			ctx.JSON(http.StatusBadGateway, gin.H{"message": "密码错误，拒绝访问"})
		}

	})
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, "Pong")
	})
	r.Run(":8080")
}

func MD5(str string) string {
	data := []byte(str)
	has := md5.Sum(data)
	md5str := fmt.Sprintf("%x", has) //将[]byte转成16进制
	return md5str
}
