package main

import (
	"fmt"

	libconfig "github.com/opensourceways/community-robot-lib/config"
)

type configuration struct {
	ConfigItems []botConfig `json:"config_items,omitempty"`
}

func (c *configuration) configFor(org, repo string) *botConfig {
	if c == nil {
		return nil
	}

	items := c.ConfigItems
	v := make([]libconfig.IPluginForRepo, len(items))
	for i := range items {
		v[i] = &items[i]
	}

	if i := libconfig.FindConfig(org, repo, v); i >= 0 {
		return &items[i]
	}
	return nil
}

func (c *configuration) Validate() error {
	if c == nil {
		return nil
	}

	items := c.ConfigItems
	for i := range items {
		if err := items[i].validate(); err != nil {
			return err
		}
	}
	return nil
}

func (c *configuration) SetDefault() {
	if c == nil {
		return
	}

	Items := c.ConfigItems
	for i := range Items {
		Items[i].setDefault()
	}
}

type botConfig struct {
	libconfig.PluginForRepo
	//Sizes []Size `json:"sizes" required:"true"`
	Sizes Size `json:"sizes" required:"true"`
}

type Size struct {
	S   int `json:"s" required:"true"`
	M   int `json:"m" required:"true"`
	L   int `json:"l" required:"true"`
	Xl  int `json:"xl" required:"true"`
	Xxl int `json:"xxl" required:"true"`
}

func (c *botConfig) setDefault() {
}

func (c *botConfig) validate() error {
	if err := c.PluginForRepo.Validate(); err != nil {
		return err
	}

	err := c.Sizes.validate()
	if err != nil {
		return err
	}

	return nil
}

func (s *Size) validate() error {
	if s.S <= 0 || s.M <= 0 || s.L <= 0 || s.Xl <= 0 || s.Xxl <= 0 {
		return fmt.Errorf("invalid value in config file")
	}

	if !(s.S <= s.M) || !(s.M <= s.L) || !(s.L <= s.Xl) || !(s.Xl <= s.Xxl) {
		return fmt.Errorf("has set invalid values in config, wrong size relationship")
	}

	return nil
}
