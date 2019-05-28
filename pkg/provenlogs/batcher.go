/*
 * @Author: guiguan
 * @Date:   2019-05-20T15:53:05+10:00
 * @Last modified by:   guiguan
 * @Last modified time: 2019-05-27T13:20:57+10:00
 */

package provenlogs

import (
	"context"
	"crypto/rsa"
	"errors"
	"sync"
	"time"

	"github.com/SouthbankSoftware/provenlogs/pkg/provendb"
	"github.com/SouthbankSoftware/provenlogs/pkg/ticker"
	"go.mongodb.org/mongo-driver/mongo"
)

type batcher struct {
	sync.RWMutex
	batchTime time.Duration
	batchSize int
	prvKey    *rsa.PrivateKey
	logCol    *mongo.Collection
	pdb       *provendb.ProvenDB

	ctx       context.Context
	ctxCancel context.CancelFunc
	ticker    *ticker.Ticker
	entries   []LogEntry
	entryCH   chan LogEntry
}

func newBatcher(
	batchTime time.Duration,
	batchSize int,
	prvKey *rsa.PrivateKey,
	logCol *mongo.Collection,
) *batcher {
	return &batcher{
		batchTime: batchTime,
		batchSize: batchSize,
		prvKey:    prvKey,
		logCol:    logCol,
		pdb:       provendb.NewProvenDB(logCol.Database()),
	}
}

func (b *batcher) batch(
	entry LogEntry,
) {
	b.RLock()
	defer b.RUnlock()

	if b.ctx == nil {
		return
	}

	select {
	case <-b.ctx.Done():
		return
	case b.entryCH <- entry:
	}
}

func (b *batcher) run(
	ctx context.Context,
) error {
	// init and destroy should be paired
	init := func(ctx context.Context) error {
		b.Lock()
		defer b.Unlock()

		if b.ctx != nil {
			return errors.New("the batcher is already running")
		}

		b.ctx, b.ctxCancel = context.WithCancel(ctx)
		b.ticker = ticker.NewTicker(b.batchTime)
		b.entries = make([]LogEntry, 0, b.batchSize)
		b.entryCH = make(chan LogEntry, b.batchSize/10)

		return nil
	}

	destroy := func() {
		b.Lock()
		defer b.Unlock()

		if b.ctx == nil {
			return
		}

		b.ctxCancel()
		b.ctx, b.ctxCancel = nil, nil
		b.ticker.Stop()
		b.ticker = nil
		b.entries = nil
		close(b.entryCH)
		b.entryCH = nil
	}

	sign := func() error {
		es := LogEntries(b.entries)

		es.Sort()

		sig, err := es.Sign(b.prvKey)
		if err != nil {
			return err
		}

		es.AttachSig(sig)

		return nil
	}

	// should be non-interruptible and not be canceled when the batcher context is canceled, this is
	// to avoid data lose
	submit := func() error {
		es := LogEntries(b.entries)
		ctx := context.Background()

		_, err := b.logCol.InsertMany(ctx, es.AnyArray())
		if err != nil {
			return err
		}

		gR, err := b.pdb.GetVersion(ctx)
		if err != nil {
			return err
		}

		_, err = b.pdb.SubmitProof(ctx, gR.Version)
		return err
	}

	// should only be called within the same goroutine to avoid unnecessary concurrency control
	flush := func() error {
		if len(b.entries) == 0 {
			return nil
		}

		serv.l.Debug("finalizing batch")

		serv.l.Debug("sign batch")
		err := sign()
		if err != nil {
			return err
		}

		serv.l.Debug("submit to ProvenDB")
		err = submit()
		if err != nil {
			return err
		}

		b.entries = b.entries[:0]
		serv.l.Debug("finalized batch")

		return nil
	}

	err := init(ctx)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			err := flush()
			destroy()
			return err
		case entry := <-b.entryCH:
			if len(b.entries) >= b.batchSize {
				b.ticker.Reset()
				err := flush()
				if err != nil {
					return err
				}
			}

			b.entries = append(b.entries, entry)
		case <-b.ticker.C:
			err := flush()
			if err != nil {
				return err
			}
		}
	}
}
