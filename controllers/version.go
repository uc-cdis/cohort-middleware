package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/uc-cdis/cohort-middleware/models"
	"net/http"
)

type VersionController struct{}

var versionModel = new(models.Version)

func (u VersionController) Retrieve(c *gin.Context) {
	version := versionModel.GetVersion()
	c.JSON(http.StatusOK, gin.H{"version": version})
	return
}
