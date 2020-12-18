package golug_rabbitmq

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/streadway/amqp"
	"vitess.io/vitess/go/pools"
)

const (
	DefaultMQTimeout = time.Second * 2
	DefaultIdleTime  = 0
	DefaultHeartbeat = time.Second * 2
	DefaultCapacity  = 10
	MaxCapacity      = 30
)

var (
	ErrorOutOfCapacity = errors.New("rabbitMQ resource pool has got the Max Capacity. ")
)

var factories = sync.Map{}

type ResourcePool struct {
	*pools.ResourcePool
}

// NewResourcePool create a resource pool
func NewResourcePool(config *RabbitConfig) (*ResourcePool, error) {

	// Set the client pool by DefaultCapacity
	resourcePool := pools.NewResourcePool(newResource(config), DefaultCapacity, MaxCapacity, DefaultIdleTime, 0, func(t time.Time) {})

	// Check the connect is ok
	ctx, cancel := context.WithTimeout(context.TODO(), DefaultMQTimeout)
	defer cancel()

	conn, err := resourcePool.Get(ctx)
	if err != nil {
		return nil, err
	}
	resourcePool.Put(conn)

	return &ResourcePool{
		ResourcePool: resourcePool,
	}, nil
}

func (rp *ResourcePool) Get() (*Resource, error) {
	ctx, cancel := context.WithTimeout(context.TODO(), DefaultMQTimeout)
	defer cancel()

	r, err := rp.ResourcePool.Get(ctx)
	if err != nil {
		if err == pools.ErrTimeout {
			cp := rp.Capacity()
			if cp == MaxCapacity {
				return nil, ErrorOutOfCapacity
			}

			cp += DefaultCapacity
			if cp >= MaxCapacity {
				cp = MaxCapacity
			}
			if err := rp.ResourcePool.SetCapacity(int(cp)); err != nil {
				return nil, err
			}

			ctx, cancel = context.WithTimeout(context.TODO(), DefaultMQTimeout)
			defer cancel()
			r, err := rp.ResourcePool.Get(ctx)
			if err != nil {
				return nil, err
			}
			return r.(*Resource), nil
		}
		return nil, err
	}

	return r.(*Resource), nil
}

// Resource adapts a client connection to a Vitess Resource.
type Resource struct {
	*amqp.Connection
	config *RabbitConfig
}

// newResource return an closure for create a client connection
func newResource(config *RabbitConfig) (factory pools.Factory) {
	factory = func(ctx context.Context) (pools.Resource, error) {
		c := amqp.Config{
			Heartbeat: DefaultHeartbeat,
		}
		conn, err := amqp.DialConfig(config.URL, c)

		if err != nil {
			if conn != nil {
				if err := conn.Close(); err != nil {
					return nil, err
				}
			}
			return nil, err
		}

		return &Resource{conn, config}, nil
	}
	factories.Store(config.Key, factory)
	return factory
}

// Close is put the conn to the poll
func (r *Resource) Close() {
	if err := r.Connection.Close(); err != nil {
		log.Errorf("conn close error %v", err)
	}
}
