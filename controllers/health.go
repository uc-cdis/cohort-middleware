package controllers

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type HealthController struct{}

// Status godoc
// @Summary Show healthcheck for cohort-middleware.
// @Description
// @Tags root
// @Accept */*
// @Produce plain
// @Success 200 {string} string "ok!"
// @Router /_health [get]
func (h HealthController) Status(c *gin.Context) {
	c.String(http.StatusOK, "ok!")
}
