package api

import (
	db "github.com/Saker233/simplebank/db/sqlc"
	"github.com/gin-gonic/gin"
)

// serves HTTP resquests for banking service
type Server struct {
	store *db.Store
	router *gin.Engine
}

// creates a server and setup routes
func NewServer(store *db.Store) *Server {
	server := &Server{store: store}
	router := gin.Default()

	// add routes
	router.POST("/accounts", server.createAccount)
	router.GET("/accounts/:id", server.getAccount)
	router.GET("/accounts", server.listAccounts)


	server.router = router
	return server
}

func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}