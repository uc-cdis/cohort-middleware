package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/uc-cdis/cohort-middleware/models"
)

type VersionController struct{}

var versionModel = new(models.Version)

func (u VersionController) Retrieve(c *gin.Context) {
	version := versionModel.GetVersion()
	c.JSON(http.StatusOK, gin.H{"version": version})
}

func (u VersionController) RetrieveSchemaVersion(c *gin.Context) {
	version := versionModel.GetSchemaVersion()
	c.JSON(http.StatusOK, gin.H{"version": version})
}
