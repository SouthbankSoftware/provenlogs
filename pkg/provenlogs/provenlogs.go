/*
 * @Author: guiguan
 * @Date:   2019-05-20T14:33:09+10:00
 * @Last modified by:   guiguan
 * @Last modified time: 2019-05-26T23:25:40+10:00
 */

package provenlogs

import (
	"bufio"
	"context"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"github.com/SouthbankSoftware/provenlogs/pkg/provendb"
	"github.com/SouthbankSoftware/provenlogs/pkg/rsakey"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/network/connstring"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

var (
	serv     *Server
	servOnce = new(sync.Once)
)

// Server represents a ProvenLogs service instance
type Server struct {
	provenDBURI     string
	provenDBColName string
	l               *zap.Logger
	prvKeyPath      string
	batchTime       time.Duration
	batchSize       int
}

// NewServer creates a new singleton ProvenLogs service instance
func NewServer(
	provenDBURI string,
	provenDBColName string,
	logger *zap.Logger,
	prvKeyPath string,
	batchTime time.Duration,
	batchSize int,
) *Server {
	servOnce.Do(func() {
		serv = &Server{
			provenDBURI:     provenDBURI,
			provenDBColName: provenDBColName,
			l:               logger,
			prvKeyPath:      prvKeyPath,
			batchTime:       batchTime,
			batchSize:       batchSize,
		}
	})

	return serv
}

// Run runs the ProvenLogs service
func (s *Server) Run(ctx context.Context) error {
	prvPEM, err := ioutil.ReadFile(s.prvKeyPath)
	if err != nil {
		return err
	}

	prv, err := rsakey.ImportPrivateKeyFromPEM(prvPEM)
	if err != nil {
		return err
	}

	var parser Parser = NewZapProductionParser()

	logCol, err := getLogCol(ctx, s.provenDBURI, s.provenDBColName)
	if err != nil {
		return err
	}

	batcher := newBatcher(
		s.batchTime,
		s.batchSize,
		prv,
		logCol,
	)

	pipeR, pipeW := io.Pipe()
	tee := io.TeeReader(os.Stdin, pipeW)

	eg, egCtx := errgroup.WithContext(ctx)
	bcCtx, bcCancel := context.WithCancel(egCtx)

	eg.Go(func() error {
		return batcher.run(bcCtx)
	})

	eg.Go(func() error {
		scanner := bufio.NewScanner(pipeR)

		for scanner.Scan() {
			line := scanner.Text()

			batcher.batch(parser.Parse(line))
		}

		return scanner.Err()
	})

	eg.Go(func() error {
		_, err := io.Copy(os.Stdout, tee)

		// just give some grace time to the pipe
		time.Sleep(time.Second)

		pipeW.Close()
		bcCancel()

		return err
	})

	return eg.Wait()
}

func getLogCol(
	ctx context.Context,
	provenDBURI,
	provenDBColName string,
) (*mongo.Collection, error) {
	cs, err := connstring.Parse(provenDBURI)
	if err != nil {
		return nil, err
	}

	if cs.Database == "" {
		return nil, errors.New("database name must be provided in the ProvenDB URI")
	}

	client, err := mongo.NewClient(options.
		Client().
		ApplyURI(provenDBURI).
		SetRegistry(DefaultBSONRegistry()),
	)
	if err != nil {
		return nil, err
	}

	err = client.Connect(ctx)
	if err != nil {
		return nil, err
	}

	db := client.Database(cs.Database)
	err = provendb.NewProvenDB(db).ShowMetaData(ctx, true)
	if err != nil {
		return nil, err
	}

	return db.Collection(provenDBColName), nil
}
