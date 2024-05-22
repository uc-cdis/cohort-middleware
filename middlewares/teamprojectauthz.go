package middlewares

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/uc-cdis/cohort-middleware/config"
	"github.com/uc-cdis/cohort-middleware/models"
	"github.com/uc-cdis/cohort-middleware/utils"
)

type TeamProjectAuthzI interface {
	TeamProjectValidationForCohort(ctx *gin.Context, cohortDefinitionId int) bool
	TeamProjectValidation(ctx *gin.Context, cohortDefinitionIds []int, filterCohortPairs []utils.CustomDichotomousVariableDef) bool
	TeamProjectValidationForCohortIdsList(ctx *gin.Context, uniqueCohortDefinitionIdsList []int) bool
	HasAccessToTeamProject(ctx *gin.Context, teamProject string) bool
}

type HttpClientI interface {
	Do(req *http.Request) (*http.Response, error)
}

type TeamProjectAuthz struct {
	cohortDefinitionModel models.CohortDefinitionI
	httpClient            HttpClientI
}

func NewTeamProjectAuthz(cohortDefinitionModel models.CohortDefinitionI, httpClient HttpClientI) TeamProjectAuthz {
	return TeamProjectAuthz{
		cohortDefinitionModel: cohortDefinitionModel,
		httpClient:            httpClient,
	}
}

func (u TeamProjectAuthz) HasAccessToTeamProject(ctx *gin.Context, teamProject string) bool {
	teamProjectAsResourcePath := teamProject
	teamProjectAccessService := "atlas-argo-wrapper-and-cohort-middleware"

	req, err := PrepareNewArboristRequestForResourceAndService(ctx, teamProjectAsResourcePath, teamProjectAccessService)
	if err != nil {
		ctx.AbortWithStatus(500)
		panic("Error while preparing Arborist request")
	}
	// send the request to Arborist:
	resp, _ := u.httpClient.Do(req)
	log.Printf("Got response status %d from Arborist...", resp.StatusCode)

	// arborist will return with 200 if the user has been granted access to the cohort-middleware URL in ctx:
	if resp.StatusCode == 200 {
		return true
	} else {
		// unauthorized or otherwise:
		log.Printf("Authorization check for team project failed with status %d ...", resp.StatusCode)
		return false
	}
}

func (u TeamProjectAuthz) hasAccessToAtLeastOne(ctx *gin.Context, teamProjects []string) bool {
	for _, teamProject := range teamProjects {
		if u.HasAccessToTeamProject(ctx, teamProject) {
			return true
		} else {
			// unauthorized:
			log.Printf("NO access to team project...checking next one (if any)...")
		}
	}
	log.Printf("NO access to any of the team projects queried...")
	return false
}

func (u TeamProjectAuthz) TeamProjectValidationForCohort(ctx *gin.Context, cohortDefinitionId int) bool {
	filterCohortPairs := []utils.CustomDichotomousVariableDef{}
	return u.TeamProjectValidation(ctx, []int{cohortDefinitionId}, filterCohortPairs)
}

func (u TeamProjectAuthz) TeamProjectValidation(ctx *gin.Context, cohortDefinitionIds []int, filterCohortPairs []utils.CustomDichotomousVariableDef) bool {

	uniqueCohortDefinitionIdsList := utils.GetUniqueCohortDefinitionIdsList(cohortDefinitionIds, filterCohortPairs)
	return u.TeamProjectValidationForCohortIdsList(ctx, uniqueCohortDefinitionIdsList)
}

// "team project" related checks:
// (1) check if any cohorts belong to the "global reader role"
//
//		(1.1) if yes, check if ALL cohorts belong to the "global reader role".
//	       If so, return true.
//
// (2) check if all remaining cohorts belong to the same "team project"
// (3) check if the user has permission in the "team project"
// Returns true if all checks above pass, false otherwise.
func (u TeamProjectAuthz) TeamProjectValidationForCohortIdsList(ctx *gin.Context, uniqueCohortDefinitionIdsList []int) bool {

	// validate input:
	if len(uniqueCohortDefinitionIdsList) == 0 {
		log.Printf("Invalid request error: NO cohort ids in list to check")
		return false
	}
	conf := config.GetConfig()
	globalReaderRole := conf.GetString("global_reader_role")
	globalCohortDefinitionIds, _ := u.cohortDefinitionModel.GetCohortDefinitionIdsForTeamProject(globalReaderRole)
	// check overlap:
	overlapWithGlobal := utils.Intersect(uniqueCohortDefinitionIdsList, globalCohortDefinitionIds)
	// and for the following checks, filter out the cohorts associated with 'global reader role':
	cohortDefinitionIdsToCheck := utils.Subtract(uniqueCohortDefinitionIdsList, globalCohortDefinitionIds)
	if len(overlapWithGlobal) > 0 {
		// one or more cohortDefinitionIds are part of globalReaderRole. Check if all of them are global:
		if len(cohortDefinitionIdsToCheck) == 0 {
			// all cohortDefinitionIds are global,
			// so return true:
			return true
		}
	}
	// proceed with the checks on the remaining list of cohortDefinitionIds:
	teamProjects, _ := u.cohortDefinitionModel.GetTeamProjectsThatMatchAllCohortDefinitionIds(cohortDefinitionIdsToCheck)
	if len(teamProjects) == 0 {
		log.Printf("Invalid request error: could not find a 'team project' that is associated to ALL the cohorts present in this request")
		return false
	}
	if !u.hasAccessToAtLeastOne(ctx, teamProjects) {
		log.Printf("Invalid request error: user does not have access to any of the 'team projects' associated with the cohorts in this request")
		return false
	}
	// passed both tests:
	return true
}
