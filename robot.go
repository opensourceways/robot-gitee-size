package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/opensourceways/community-robot-lib/giteeclient"
	"github.com/sirupsen/logrus"

	sdk "gitee.com/openeuler/go-gitee/gitee"
	libconfig "github.com/opensourceways/community-robot-lib/config"
	libplugin "github.com/opensourceways/community-robot-lib/giteeplugin"
)

const (
	botName     = "size"
	labelPrefix = "size/"
	labelXS     = "size/XS"
	labelS      = "size/S"
	labelM      = "size/M"
	labelL      = "size/L"
	labelXL     = "size/XL"
	labelXXL    = "size/XXL"
)

type iClient interface {
	AddPRLabel(org, repo string, number int32, label string) error
	RemovePRLabel(org, repo string, number int32, label string) error
	GetPullRequestChanges(org, repo string, number int32) ([]sdk.PullRequestFiles, error)
}

func newRobot(cli iClient) *robot {
	return &robot{cli: cli}
}

type robot struct {
	cli iClient
}

func (bot *robot) NewPluginConfig() libconfig.PluginConfig {
	return &configuration{}
}

func (bot *robot) getConfig(cfg libconfig.PluginConfig, org, repo string) (*botConfig, error) {
	c, ok := cfg.(*configuration)
	if !ok {
		return nil, fmt.Errorf("can't convert to configuration")
	}

	if bc := c.configFor(org, repo); bc != nil {
		return bc, nil
	}

	return nil, fmt.Errorf("no config for this repo:%s/%s", org, repo)
}

func (bot *robot) RegisterEventHandler(p libplugin.HandlerRegitster) {
	p.RegisterPullRequestHandler(bot.handlePREvent)
}

func (bot *robot) handlePREvent(e *sdk.PullRequestEvent, pc libconfig.PluginConfig, log *logrus.Entry) error {
	if !isPRChanged(e) {
		log.Info("pull request is not opened or source_branch_changed, skipping")

		return nil
	}

	pr := giteeclient.GetPRInfoByPREvent(e)
	org, repo := pr.Org, pr.Repo

	cfg, err := bot.getConfig(pc, org, repo)
	if err != nil {
		return err
	}

	return bot.handlePR(pr, cfg)
}

func (bot *robot) handlePR(pr giteeclient.PRInfo, cfg *botConfig) error {
	changeFiles, err := bot.cli.GetPullRequestChanges(pr.Org, pr.Repo, pr.Number)
	if err != nil {
		return err
	}

	var changeCount int

	for _, file := range changeFiles {
		addCount, _ := strconv.Atoi(file.Additions)
		deleteCount, _ := strconv.Atoi(file.Deletions)
		changeCount += addCount + deleteCount
	}

	label := bot.compareAndGetLabel(changeCount, cfg.Sizes)

	if pr.HasLabel(label) {
		return nil
	}

	for l := range pr.Labels {
		if strings.HasPrefix(l, labelPrefix) {

			if err := bot.cli.RemovePRLabel(pr.Org, pr.Repo, pr.Number, l); err != nil {
				return err
			}
		}
	}

	return bot.cli.AddPRLabel(pr.Org, pr.Repo, pr.Number, label)
}

func (bot *robot) compareAndGetLabel(totalCount int, size Size) (label string) {
	if totalCount < size.S {
		return labelXS
	} else if totalCount < size.M {
		return labelS
	} else if totalCount < size.L {
		return labelM
	} else if totalCount < size.Xl {
		return labelL
	} else if totalCount < size.Xxl {
		return labelXL
	}

	return labelXXL
}

func isPRChanged(e *sdk.PullRequestEvent) bool {
	switch e.GetActionDesc() {
	case "open":
		return true
	case giteeclient.PRActionChangedSourceBranch:
		return true
	default:
		return false
	}
}
