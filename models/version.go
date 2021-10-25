package models

import "github.com/uc-cdis/cohort-middleware/version"

type Version struct {
	GitCommit  string
	GitVersion string
}

func (h Version) GetVersion() *Version {
	return &Version{GitCommit: version.GitCommit, GitVersion: version.GitVersion}
}
