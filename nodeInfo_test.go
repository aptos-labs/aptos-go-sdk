package aptos

import (
	"bytes"
	"context"
	"log/slog"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

type levelCounts struct {
	lock   sync.Mutex
	counts map[slog.Level]int
}

func (lc *levelCounts) inc(level slog.Level) {
	lc.lock.Lock()
	defer lc.lock.Unlock()
	lc.counts[level]++
}

func (lc *levelCounts) get(level slog.Level) int {
	lc.lock.Lock()
	defer lc.lock.Unlock()
	return lc.counts[level]
}

func newLevelCounts() *levelCounts {
	out := new(levelCounts)
	out.counts = make(map[slog.Level]int, 5)
	return out
}

type CountingHandlerWrapper struct {
	inner  slog.Handler
	counts *levelCounts
}

func NewCountingHandlerWrapper(inner slog.Handler) *CountingHandlerWrapper {
	return &CountingHandlerWrapper{
		inner:  inner,
		counts: newLevelCounts(),
	}
}

func (chw *CountingHandlerWrapper) Enabled(ctx context.Context, level slog.Level) bool {
	return chw.inner.Enabled(ctx, level)
}

func (chw *CountingHandlerWrapper) Handle(ctx context.Context, rec slog.Record) error {
	chw.counts.inc(rec.Level)
	return chw.inner.Handle(ctx, rec)
}

func (chw *CountingHandlerWrapper) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &CountingHandlerWrapper{
		inner:  chw.inner.WithAttrs(attrs),
		counts: chw.counts,
	}
}

func (chw *CountingHandlerWrapper) WithGroup(name string) slog.Handler {
	return &CountingHandlerWrapper{
		inner:  chw.inner.WithGroup(name),
		counts: chw.counts,
	}
}

type testSlogContext struct {
	logbuf          bytes.Buffer
	jsonHandler     *slog.JSONHandler
	countingHandler *CountingHandlerWrapper
	logger          *slog.Logger
	oldDefault      *slog.Logger
}

func setupTestLogging() *testSlogContext {
	logContext := &testSlogContext{}
	logContext.jsonHandler = slog.NewJSONHandler(
		&logContext.logbuf,
		&slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	)
	logContext.countingHandler = NewCountingHandlerWrapper(logContext.jsonHandler)
	logContext.logger = slog.New(logContext.countingHandler)
	logContext.oldDefault = slog.Default()
	slog.SetDefault(logContext.logger)
	return logContext
}

func restoreNormalLogging(t *testing.T, logContext *testSlogContext) {
	if t.Failed() {
		t.Log(logContext.logbuf.String())
	}
	slog.SetDefault(logContext.oldDefault)
}

func TestEpoch(t *testing.T) {
	lc := setupTestLogging()
	defer restoreNormalLogging(t, lc)

	info := NodeInfo{
		EpochStr: "123",
	}
	assert.Equal(t, uint64(123), info.Epoch())
	info.EpochStr = "garbage"
	assert.Equal(t, uint64(0), info.Epoch())
	assert.Equal(t, 1, lc.countingHandler.counts.get(slog.LevelError))
}

func TestLedgerVersion(t *testing.T) {
	lc := setupTestLogging()
	defer restoreNormalLogging(t, lc)

	info := NodeInfo{
		LedgerVersionStr: "123",
	}
	assert.Equal(t, uint64(123), info.LedgerVersion())
	info.LedgerVersionStr = "garbage"
	assert.Equal(t, uint64(0), info.LedgerVersion())
	assert.Equal(t, 1, lc.countingHandler.counts.get(slog.LevelError))
}

func TestOldestLedgerVersion(t *testing.T) {
	lc := setupTestLogging()
	defer restoreNormalLogging(t, lc)

	info := NodeInfo{
		OldestLedgerVersionStr: "123",
	}
	assert.Equal(t, uint64(123), info.OldestLedgerVersion())
	info.OldestLedgerVersionStr = "garbage"
	assert.Equal(t, uint64(0), info.OldestLedgerVersion())
	assert.Equal(t, 1, lc.countingHandler.counts.get(slog.LevelError))
}

func TestBlockHeight(t *testing.T) {
	lc := setupTestLogging()
	defer restoreNormalLogging(t, lc)

	info := NodeInfo{
		BlockHeightStr: "123",
	}
	assert.Equal(t, uint64(123), info.BlockHeight())
	info.BlockHeightStr = "garbage"
	assert.Equal(t, uint64(0), info.BlockHeight())
	assert.Equal(t, 1, lc.countingHandler.counts.get(slog.LevelError))
}

func TestOldestBlockHeight(t *testing.T) {
	lc := setupTestLogging()
	defer restoreNormalLogging(t, lc)

	info := NodeInfo{
		OldestBlockHeightStr: "123",
	}
	assert.Equal(t, uint64(123), info.OldestBlockHeight())
	info.OldestBlockHeightStr = "garbage"
	assert.Equal(t, uint64(0), info.OldestBlockHeight())
	assert.Equal(t, 1, lc.countingHandler.counts.get(slog.LevelError))
}
