package controllers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/uc-cdis/cohort-middleware/middlewares"
	"github.com/uc-cdis/cohort-middleware/models"
)

type SourceController struct {
	sourceModel      models.SourceI
	teamProjectAuthz middlewares.TeamProjectAuthzI
}

func NewSourceController(sourceModel models.SourceI, teamProjectAuthz middlewares.TeamProjectAuthzI) SourceController {
	return SourceController{
		sourceModel:      sourceModel,
		teamProjectAuthz: teamProjectAuthz,
	}
}

func (u SourceController) RetriveById(c *gin.Context) {
	sourceId := c.Param("id")
	if sourceId != "" {
		sourceId, _ := strconv.Atoi(c.Param("id"))
		source, err := u.sourceModel.GetSourceById(sourceId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Error to retrieve source", "error": err.Error()})
			c.Abort()
			return
		}
		c.JSON(http.StatusOK, gin.H{"source": source})
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"message": "bad request"})
	c.Abort()
}

func (u SourceController) RetriveByName(c *gin.Context) {
	sourceName := c.Param("name")
	if sourceName != "" {
		source, err := u.sourceModel.GetSourceByName(sourceName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Error to retrieve source", "error": err.Error()})
			c.Abort()
			return
		}
		c.JSON(http.StatusOK, gin.H{"source": source})
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"message": "bad request"})
	c.Abort()
}

func (u SourceController) RetriveAll(c *gin.Context) {
	teamProject := c.Query("team-project")
	if teamProject == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error while parsing request", "error": "team-project is a mandatory parameter but was found to be empty!"})
		c.Abort()
		return
	}

	// validate teamproject access permission:
	validAccessRequest := u.teamProjectAuthz.HasAccessToTeamProject(c, teamProject)
	if !validAccessRequest {
		log.Printf("Error: invalid request")
		c.JSON(http.StatusForbidden, gin.H{"message": "access denied"})
		c.Abort()
		return
	}

	source, err := u.sourceModel.GetAllSourcesWithTeamProject(teamProject)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error to retrieve source", "error": err.Error()})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{"sources": source})
}
