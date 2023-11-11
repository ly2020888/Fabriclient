package main

import (
	"crypto/md5"
	"fmt"
	"os"

	"github.com/hyperledger-twgc/tape/pkg/infra"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

const (
	loglevel = "TAPE_LOGLEVEL"
)

var (
	app = kingpin.New("tape++", "Efficient TAPE-based client")

	run     = app.Command("run", "Start the tape program").Default()
	con     = run.Flag("config", "Path to config file").Required().Short('c').String()
	pw      = run.Flag("password", "A memory key that needs to be set").Required().Short('p').String()
	version = app.Command("version", "Show version information")
)

func main() {
	var err error

	logger := log.New()
	logger.SetLevel(log.InfoLevel)
	if customerLevel, customerSet := os.LookupEnv(loglevel); customerSet {
		if lvl, err := log.ParseLevel(customerLevel); err == nil {
			logger.SetLevel(lvl)
		}
	}

	fullCmd := kingpin.MustParse(app.Parse(os.Args[1:]))
	fmt.Printf("内建密码为:%s\n", *pw)
	infra.Temporary = MD5(*pw)
	switch fullCmd {
	case version.FullCommand():
	case run.FullCommand():
		err = infra.Process(*con, logger)
		if err != nil {
			logger.Error(err)
		}
	default:
		err = errors.Errorf("invalid command: %s", fullCmd)
	}

	if err != nil {
		logger.Error(err)
		os.Exit(1)
	}

	os.Exit(0)
}

func MD5(str string) string {
	data := []byte(str)
	has := md5.Sum(data)
	md5str := fmt.Sprintf("%x", has) //将[]byte转成16进制
	return md5str
}
