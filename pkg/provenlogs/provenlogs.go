/*
 * @Author: guiguan
 * @Date:   2019-05-20T14:33:09+10:00
 * @Last modified by:   guiguan
 * @Last modified time: 2019-05-20T14:42:11+10:00
 */

package provenlogs

import (
	"sync"

	"go.mongodb.org/mongo-driver/mongo"
)

var (
	serv     *Server
	servOnce = new(sync.Once)
)

// Server represents a Provenlogs service instance
type Server struct {
	provendbURI string

	logCol *mongo.Collection
}

// NewServer creates a new singleton Provenlogs service instance
func NewServer(
	provendbURI string,
) *Server {
	servOnce.Do(func() {
		serv = &Server{
			provendbURI: provendbURI,
		}
	})

	return serv
}
