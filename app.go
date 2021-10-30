package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

// App struct exposes references to the router and the database that the application uses.
type App struct {
	Router *gin.Engine
	DB     *sql.DB
}

// Initializes DB, gin engine and set up routers.
func (a *App) Initialize(dbPath string) {

	// Here can be added middleware for stuff like authentication. Since it's just a simple REST API example, we don't really care about auth.

	var err error
	a.DB, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Panic(err)
	}

	// This code tries to create a DB scheme if it's missing.
	// It was only made to be able to run this simple app. In real production code this is not acceptable.
	if _, err := a.DB.Exec(DomainTableCreationQuery); err != nil {
		log.Panic(err)
	}

	a.Router = gin.Default()
	a.initializeRoutes()
	log.Println("The app routes were initialized.")
}

// Attaches the router to a http.Server and starts listening and serving HTTP requests.
func (a *App) Run(addr string) {
	log.Panic(a.Router.Run(addr))
}

// Cleans up all used resources.
func (a *App) Dispose() {
	a.DB.Close()
	log.Println("All holding resources were disposed.")
}

// The place where all app's routes and handlers are defined and initialzed.
func (a *App) initializeRoutes() {
	a.Router.POST("/domains", a.postCounterRoute)
	a.Router.GET("/domains/statistics/last-minute", a.getDomainsStatisticsLastMinuteRoute)
	a.Router.GET("/domains/statistics/last-hour", a.getDomainsStatisticsLastHourRoute)
}

// The handler for GET route that returns top 10 most requested domains within last round minute.
func (a *App) getDomainsStatisticsLastMinuteRoute(c *gin.Context) {
	// getting current time.
	t := time.Now()
	// finding how much seconds passed since the last round minute
	s := int64(t.Second())
	// getting unix epoch time and substract the seconds to get last round minute.
	ut := (t.Unix() - s)

	domains, err := getTopTenDomains(a.DB, ut-60, ut)
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"domains": domains,
		"success": "true",
	})
}

// The handler for GET route that returns top 10 most requested domains within last round hour.
func (a *App) getDomainsStatisticsLastHourRoute(c *gin.Context) {
	// getting current time.
	t := time.Now()
	// finding how much seconds passed since the last round hour
	s := int64(t.Second() + t.Minute()*60)
	// getting unix epoch time and substract the seconds to get last round hour.
	ut := (t.Unix() - s)

	domains, err := getTopTenDomains(a.DB, ut-(60*60), ut)
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"domains": domains,
		"success": "true",
	})
}

// POST domains route handler.
func (a *App) postCounterRoute(c *gin.Context) {
	if c.ContentType() != "application/json" {
		respondWithError(c, http.StatusForbidden, "Unsupported content type.")
		return
	}

	var data map[string]interface{}

	buf, e := ioutil.ReadAll(c.Request.Body)
	if e != nil {
		respondWithError(c, http.StatusInternalServerError, e.Error())
		return
	}

	c.Request.Body = ioutil.NopCloser(bytes.NewReader(buf))
	e = json.Unmarshal(buf, &data)
	if e != nil {
		respondWithError(c, http.StatusInternalServerError, e.Error())
		return
	}

	for k, v := range data {
		// Making sure we're not processing 'timestamp' key here.
		if k == "timestamp" {
			continue
		}

		var d domain = domain{
			Timestamp: int(data["timestamp"].(float64)),
			Name:      k,
			Requests:  int(v.(float64)),
		}

		e = d.upsertDomain(a.DB)
		if e != nil {
			log.Print(e.Error())
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": "true",
	})
}

// Used to standardize all API error responses.
// The func returns a typical JSON response (Content-type: application/json)
// with provided HTTP Status Code and error message.
// In those responses the `success` key always has `false` value
func respondWithError(c *gin.Context, code int, msg string) {
	log.Printf("Error: %s. URL: %s", msg, c.Request.RequestURI)
	c.AbortWithStatusJSON(code, gin.H{
		"success": "false",
		"error":   msg,
	})
}
