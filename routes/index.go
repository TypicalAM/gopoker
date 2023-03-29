package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Index is the index route
func (controller Controller) Index(c *gin.Context) {
	pd := controller.DefaultPageData(c)
	pd.Title = "Home"
	c.HTML(http.StatusOK, "index.html", pd)
}
