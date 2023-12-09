package kafka_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/Goboolean/common/pkg/resolver"
	"github.com/Goboolean/fetch-system.IaC/pkg/model"
	"github.com/Goboolean/fetch-system.IaC/pkg/kafka"
	"github.com/stretchr/testify/assert"
)

func SetupProducer() *kafka.Producer {
	p, err := kafka.NewProducer(&resolver.ConfigMap{
		"BOOTSTRAP_HOST": os.Getenv("KAFKA_BOOTSTRAP_HOST"),
	})
	if err != nil {
		panic(err)
	}
	return p
}

func SetupProducerWithRegistry() *kafka.Producer {
	p, err := kafka.NewProducer(&resolver.ConfigMap{
		"BOOTSTRAP_HOST": os.Getenv("KAFKA_BOOTSTRAP_HOST"),
		"REGISTRY_HOST":  os.Getenv("KAFKA_REGISTRY_HOST"),
	})
	if err != nil {
		panic(err)
	}
	return p
}

func TeardownProducer(p *kafka.Producer) {
	mutex.Lock()
	defer mutex.Unlock()
	p.Close()
}

func Test_Producer(t *testing.T) {

	p := SetupProducer()
	defer TeardownProducer(p)

	t.Run("Ping", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		err := p.Ping(ctx)
		assert.NoError(t, err)
	})
}

func Test_Produce(t *testing.T) {

	p := SetupProducer()
	a := SetupConfigurator()

	t.Cleanup(func() {
		err := a.DeleteAllTopics(context.Background())
		assert.NoError(t, err)

		TeardownProducer(p)
		TeardownConfigurator(a)
	})

	const productId = "test.goboolean.io"
	const productType = "1s"

	var trade = &model.Trade{
		Price:     171.55,
		Size:      100,
		Timestamp: time.Now().Unix(),
	}

	var aggs = &model.Aggregate{
		Open:      170.55,
		Closed:    173.55,
		Min:       170.55,
		Max:       171.55,
		Volume:    100,
		Timestamp: time.Now().Unix(),
	}

	t.Run("ProduceTrade", func(t *testing.T) {
		err := p.ProduceTrade(productId, trade)
		assert.NoError(t, err)
	})

	t.Run("ProduceAggs", func(t *testing.T) {
		err := p.ProduceAggs(productId, productType, aggs)
		assert.NoError(t, err)
	})

	t.Run("Flush", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		count, err := p.Flush(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 0, count)
	})
}
