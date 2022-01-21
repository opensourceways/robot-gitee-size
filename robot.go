package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/opensourceways/community-robot-lib/config"
	framework "github.com/opensourceways/community-robot-lib/robot-gitee-framework"
	sdk "github.com/opensourceways/go-gitee/gitee"
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

func (bot *robot) NewConfig() config.Config {
	return &configuration{}
}

func (bot *robot) getConfig(cfg config.Config, org, repo string) (*botConfig, error) {
	c, ok := cfg.(*configuration)
	if !ok {
		return nil, fmt.Errorf("can't convert to configuration")
	}

	if bc := c.configFor(org, repo); bc != nil {
		return bc, nil
	}

	return nil, fmt.Errorf("no config for this repo:%s/%s", org, repo)
}

func (bot *robot) RegisterEventHandler(f framework.HandlerRegitster) {
	f.RegisterPullRequestHandler(bot.handlePREvent)
}

func (bot *robot) handlePREvent(e *sdk.PullRequestEvent, pc config.Config, log *logrus.Entry) error {
	if !isPRChanged(e) {
		log.Info("pull request is not opened or source_branch_changed, skipping")

		return nil
	}

	org, repo := e.GetOrgRepo()
	pr := e.GetPullRequest()

	cfg, err := bot.getConfig(pc, org, repo)
	if err != nil {
		return err
	}

	return bot.handlePR(org, repo, pr, cfg)
}

func (bot *robot) handlePR(org, repo string, pr *sdk.PullRequestHook, cfg *botConfig) error {
	number := pr.GetNumber()
	changeFiles, err := bot.cli.GetPullRequestChanges(org, repo, number)
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

	prLabels := pr.LabelsToSet()
	if prLabels.Has(label) {
		return nil
	}

	for l := range prLabels {
		if strings.HasPrefix(l, labelPrefix) {

			if err := bot.cli.RemovePRLabel(org, repo, number, l); err != nil {
				return err
			}
		}
	}

	return bot.cli.AddPRLabel(org, repo, number, label)
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
	case sdk.ActionOpen:
		return true
	case sdk.PRActionChangedSourceBranch:
		return true
	default:
		return false
	}
}
