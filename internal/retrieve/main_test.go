package retrieve_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/Goboolean/fetch-system.IaC/cmd/wire"
	"github.com/Goboolean/fetch-system.IaC/internal/retrieve"
	"github.com/Goboolean/fetch-system.IaC/pkg/db"
	"github.com/stretchr/testify/assert"

	_ "github.com/Goboolean/common/pkg/env"
)



var cleanups []func()

var manager *retrieve.Manager



func SetupManager() *retrieve.Manager {

	polygon, err := wire.InitializePolygonClient()
	if err != nil {
		panic(err)
	}

	db, cleanup, err := wire.InitializePostgreSQLClient()
	if err != nil {
		panic(err)
	}
	cleanups = append(cleanups, cleanup)

	kis, err := wire.InitializeKISReader()
	if err != nil {
		panic(err)
	}

	return retrieve.New(polygon, db, kis)
}

func Teardown() {
	for _, cleanup := range cleanups {
		cleanup()
	}
}



func TestMain(m *testing.M) {
	manager = SetupManager()
	code := m.Run()
	Teardown()
	os.Exit(code)
}


func TestStoreUSAStocks(t *testing.T) {

	database, cleanup, err := wire.InitializePostgreSQLClient()
	if assert.NoError(t, err) {
		t.Cleanup(cleanup)
	}

	t.Run("StoreUSAStocks", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		err := manager.StoreUSAStocks(ctx)
		assert.NoError(t, err)
	})

	t.Run("VerifyResult", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		tickerDetails, err := database.GetProductsByCondition(ctx, db.GetProductsByConditionParams{
			Locale: db.LocaleUsa,
			Market: db.MarketStock,
		})
		assert.NoError(t, err)
		assert.NotEmpty(t, tickerDetails)
		assert.GreaterOrEqual(t, len(tickerDetails), 10000)
	})
}