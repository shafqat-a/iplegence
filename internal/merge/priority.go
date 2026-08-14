package merge

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Priority struct {
	Country      []string `yaml:"country"`
	Continent    []string `yaml:"continent"`
	ASNNumber    []string `yaml:"asn_number"`
	ASNOrg       []string `yaml:"asn_org"`
	ASNDomain    []string `yaml:"asn_domain"`
	City         []string `yaml:"city"`
	Location     []string `yaml:"location"`
	Subdivisions []string `yaml:"subdivisions"`
	Postal       []string `yaml:"postal"`
	Traits       string   `yaml:"traits"`
}

func LoadPriority(path string) (Priority, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Priority{}, err
	}
	var p Priority
	if err := yaml.Unmarshal(b, &p); err != nil {
		return Priority{}, err
	}
	if p.Traits != "" && p.Traits != "or" {
		return Priority{}, fmt.Errorf("priority traits must be \"or\", got %q", p.Traits)
	}
	if p.Traits == "" {
		p.Traits = "or"
	}
	return p, nil
}

func rank(list []string, id string) (int, bool) {
	for i, v := range list {
		if v == id {
			return i, true
		}
	}
	return 0, false
}

func better(list []string, srcID, winnerID string) bool {
	srcRank, srcOK := rank(list, srcID)
	if !srcOK {
		return false
	}
	if winnerID == "" {
		return true
	}
	winRank, winOK := rank(list, winnerID)
	if !winOK {
		return true
	}
	return srcRank < winRank
}
