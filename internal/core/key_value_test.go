package core_test

import (
	"testing"

	"github.com/jersonsatoru/cnb/internal/core"
	"github.com/jersonsatoru/cnb/internal/logger"
	"github.com/stretchr/testify/require"
)

var tl, _ core.TransactionLogger
var kv *core.KeyValueStore

func init() {
	tl, _ = logger.NewTransactionLogger("mock")
	kv = core.NewKeyValueStore(tl)
}

func TestApplicationGet(t *testing.T) {

	kv.Put("batata", "123")
	value, err := kv.Get("batata")
	require.Nil(t, err)
	require.Equal(t, value, "123")

	value, err = kv.Get("bebe")
	require.ErrorIs(t, err, core.ErrNoSUchKey)
	require.Equal(t, value, "")
}

func TestApplicationPut(t *testing.T) {
	err := kv.Put("batata", "123")
	require.Nil(t, err)
}

func TestApplicationDelete(t *testing.T) {
	err := kv.Put("batata", "123")
	require.Nil(t, err)

	err = kv.Delete("batata")
	require.Nil(t, err)

	value, err := kv.Get("batata")
	require.ErrorIs(t, err, core.ErrNoSUchKey)
	require.Equal(t, value, "")
}
