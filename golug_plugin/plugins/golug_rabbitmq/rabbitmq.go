package golug_rabbitmq

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/pubgo/xprocess"
	"github.com/streadway/amqp"
	"vitess.io/vitess/go/pools"
)

const contentType = "application/json"
const contentEncoding = "UTF-8"

type ExchangeKind string

func (kind ExchangeKind) String() string {
	return string(kind)
}

const (
	ExchangeKindFanout  ExchangeKind = "fanout"
	ExchangeKindDirect  ExchangeKind = "direct"
	ExchangeKindTopic   ExchangeKind = "topic"
	ExchangeKindHeaders ExchangeKind = "headers"

	prefetchCount               = 200
	defaultQueueSuffix          = "_queue"
	defaultExchangeSuffix       = "_exchange"
	defaultRoutingSuffix        = "_routing"
	defaultRetryConsumeDuration = time.Millisecond * 1000
)

var (
	initRabbitPoolMap = sync.Map{}
	initStatusDic     = sync.Map{}
)

type RabbitConfig struct {
	Key string
	URL string
}

const (
	retryCount = 2
)

type RabbitChannel struct {
	*amqp.Channel
	prefix string
	close  bool
	ctx    context.Context
	wg     *sync.WaitGroup
}

func (r *RabbitChannel) GetContext() context.Context {
	return r.ctx
}

func (r *RabbitChannel) GetPrefix() string {
	return r.prefix
}

// 从资源池获取资源
func pickupRabbitPool(rabbitName string) (*ResourcePool, error) {
	lowerName := strings.ToLower(rabbitName)
	result, ok := initRabbitPoolMap.Load(lowerName)
	if ok {
		return result.(*ResourcePool), nil
	}
	log.Errorf("can not get client pool ,rabbitName= %s", rabbitName)
	initRabbitPoolMap.Range(func(key, value interface{}) bool {
		log.Debugf("client pool has client key(%v)", key)
		return true
	})
	return nil, errors.New("can not get client pool ,rabbitName=" + rabbitName)
}

// 判断是否存在
func isRabbitPoolExist(rabbitName string) *ResourcePool {
	lowerName := strings.ToLower(rabbitName)
	result, ok := initRabbitPoolMap.Load(lowerName)
	if ok {
		return result.(*ResourcePool)
	}
	return nil
}

// 组装转换数据，初始化资源池
func initRabbitPool(rabbitName string, rabbitConfig *RabbitConfig) error {
	resourcePool, err := NewResourcePool(rabbitConfig)
	if err != nil {
		log.Errorf("client(%+v) 连接失败, error=%+v", rabbitName, err)
		return fmt.Errorf("client(%+v) 连接失败, error=%+v", rabbitName, err)
	}

	pool := isRabbitPoolExist(rabbitName)

	lowerName := strings.ToLower(rabbitName)
	initRabbitPoolMap.Store(lowerName, resourcePool)
	deferCloseRabbitMQ(pool, rabbitName)
	log.Infof("rebuild client pool done - %v", lowerName)

	return nil
}

// 删除资源池
func deleteRabbitPool(rabbitName string) {
	lowerName := strings.ToLower(rabbitName)
	pool := isRabbitPoolExist(lowerName)
	initRabbitPoolMap.Delete(lowerName)
	log.Infof("delete client pool done - %v", lowerName)
	deferCloseRabbitMQ(pool, lowerName)
}

// 关闭资源池
func deferCloseRabbitMQ(pool *ResourcePool, lowerName string) {
	if pool != nil {
		time.AfterFunc(time.Second*2, func() {
			pool.Close()
			log.Infof("close client client: %v", lowerName)
		})
	}
}

/*
	获取rabbit客户端
	ctx
	rabbitName 配置中的信息
*/
// Caution!!! This function return a channel, you should close it after use.
func PickupRabbitClient(ctx context.Context, rabbitName string) (*RabbitChannel, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context nil")
	}
	for i := 0; i < retryCount; i++ {
		resourcePool, err := pickupRabbitPool(rabbitName)
		if err != nil {
			return nil, fmt.Errorf("convert RabbitMQ fail, InitRbmqMap.Load(%s) = %+v", rabbitName, resourcePool)
		}

		resource, err := resourcePool.Get()
		if err != nil {
			return nil, fmt.Errorf("get RabbitMQ from the pool failed. %+v ", err)
		}

		ch, err := resource.Channel()
		// 获取ch异常
		if err != nil {
			// ch如果是关闭的异常,则通过对应的factory,创建新的resource
			if err == amqp.ErrClosed {
				val, ok := factories.Load(rabbitName)
				// 如果获取factory失败
				if !ok {
					resourcePool.Put(resource)
					return nil, fmt.Errorf("not found new resource factory")
				}
				factory := val.(pools.Factory)
				resourceNew, err := factory(ctx)
				// 无法创建新的resource
				if err != nil {
					// 放回旧的resource,这里不放回去会导致goroutine泄漏
					resourcePool.Put(resource)
					// factory 新的resource失败，则重启pool
					log.Errorf("factory new resource error %v,to init rabbit pool", err)
					// 重启resource池子
					if err := initRabbitPool(rabbitName, resource.config); err != nil {
						log.Errorf("init Rabbit pool error %v,rabbit(%s)", err, rabbitName)
						break
					}
				} else {
					// 如果已经factory成功了，则把老的resource关闭，释放对象
					_ = xprocess.Go(func(ctx context.Context) error {
						resource.Close()
						return nil
					})
					log.Debugf("put new resource to pool")
					// 放入新的resource
					resourcePool.Put(resourceNew)
				}
			} else {
				// 如果是其他异常
				log.Errorf("channel get error %v rabbitName=%s", err, rabbitName)
				// 这里不放回去会导致goroutine泄漏
				resourcePool.Put(resource)
			}

			continue
		} else {
			// 如果没有异常
			// Put back if the connection is ok
			resourcePool.Put(resource)
		}

		// We Should Set Qos to [100-300] due to https://www.rabbitmq.com/confirms.html#channel-qos-prefetch
		if err := ch.Qos(prefetchCount, 0, false); err != nil {
			return nil, err
		}

		return &RabbitChannel{
			Channel: ch,
			prefix:  rabbitName,
			ctx:     ctx,
			wg:      &sync.WaitGroup{},
		}, nil
	}

	return nil, fmt.Errorf("client(%+v) failed when retry %+v times. ", rabbitName, retryCount)
}

/*重新获取channel
 */
func (r *RabbitChannel) rePickupChannel() error {
	err := r.Channel.Close()
	if err != nil && err != amqp.ErrClosed {
		return err
	}

	newCh, err := PickupRabbitClient(r.ctx, r.prefix)
	if err != nil {
		return err
	}

	r.Channel = newCh.Channel
	r.close = false
	return nil
}

/*
 定义exchange
 参数：
	name exchange名称
	kind exchange种类
	durable 是否持久化
	autoDelete 当所有绑定的队列都与交换器解绑后，交换器会自动删除
 返回值：
	error 操作期间产生的错误
*/
func (r *RabbitChannel) DeclareExchange(name string, kind string, durable, autoDelete bool) error {
	//TODO tracing
	switch kind {
	case ExchangeKindDirect.String(), ExchangeKindFanout.String(), ExchangeKindTopic.String(), ExchangeKindHeaders.String():
	default:
		return fmt.Errorf("kind must fanout,direct,topic,headers ")
	}
	decErr := r.ExchangeDeclare(name, kind, durable, autoDelete, false, false, nil)
	if decErr != nil {
		return decErr
	}
	return nil
}

/*
  删除 exchange
  参数：
	name exchange名称
  返回值：
	error 操作期间产生的错误
*/
func (r *RabbitChannel) DeleteExchange(name string) error {
	//TODO tracing
	delErr := r.ExchangeDelete(name, false, true)
	if delErr != nil {
		return delErr
	}
	return nil
}

/*
 定义队列
 参数：
	name 队列的名称
	durable 是否持久化
	autoDelete 当所有消费者都断开时，队列会自动删除
 返回值：
	error 操作期间产生的错误
*/
func (r *RabbitChannel) DeclareQueue(name string, durable, autoDelete bool) (amqp.Queue, error) {
	//TODO tracing
	q, queErr := r.QueueDeclare(name, durable, autoDelete, false, false, nil)
	if queErr != nil {
		return q, queErr
	}
	return q, nil
}

/*
  删除队列
  参数：
	name 队列名称
  返回值：
	int 清除的消息数
	error 操作期间产生的错误
*/
func (r *RabbitChannel) DeleteQueue(name string) (int, error) {
	//TODO tracing
	count, delErr := r.QueueDelete(name, false, false, false)
	if delErr != nil {
		return count, delErr
	}
	return count, nil
}

/*
 exchange和queue绑定
 参数：
	queue 队列名称
	bindkey 绑定的key
	exchange 交换器名称
 返回值：
	error 操作期间产生的错误
*/
func (r *RabbitChannel) Bind(queue, bindKey, exchange string) error {
	//TODO tracing
	bindErr := r.QueueBind(queue, bindKey, exchange, false, nil)
	if bindErr != nil {
		return bindErr
	}
	return nil
}

/*
 发送消息到指定的exchange
 参数：
	exchange 制定交换机
	routingKey 路由key
 	msg 消息内容
    deliveryMode 持久化 Transient (0 or 1) or Persistent (2)
 返回值：
	error 操作期间产生的错误
*/
func (r *RabbitChannel) Publish(exchange, routingKey string, msg []byte, deliveryMode uint8) error {
	// TODO tracing
	if deliveryMode > 2 {
		return errors.New("deliveryMode is Transient (0 or 1) or Persistent (2)")
	}
	headers := amqp.Table{}
	pubErr := r.Channel.Publish(exchange, routingKey, false, false, amqp.Publishing{
		Headers:         headers,
		ContentType:     contentType,
		ContentEncoding: contentEncoding,
		DeliveryMode:    deliveryMode, // 持久化
		Body:            msg,
	})
	if pubErr != nil {
		return pubErr
	}
	return nil
}

/*
 消费消息队列
 参数：
	queue 队列名称
	autoAck	是否自动回复ack
 返回值：
	Delivery 传递消息的单向通道，可以通过读取该通道获取接收到的消息
	error 操作期间产生的错误
*/
func (r *RabbitChannel) Consume(queue string, autoAck bool) (<-chan amqp.Delivery, error) {
	// TODO tracing
	ch, subErr := r.Channel.Consume(queue, "", autoAck, false, false, false, nil)
	if subErr != nil {
		return ch, subErr
	}
	return ch, nil
}

/*
  关闭channel
*/
func (r *RabbitChannel) Close() error {
	r.wg.Wait()
	if r.Channel != nil {
		r.close = true
		return r.Channel.Close()
	}
	return nil
}

/*
  简单发布消息
  参数：
	topic 发布消息的topic
 	msg 消息内容
    opts 预留opts操作
*/
func (r *RabbitChannel) DioPublish(topic string, msg []byte, opts ...PublishOption) error {
	if len(msg) == 0 {
		return fmt.Errorf("PublishError:MsgNil")
	}
	if topic == "" {
		return fmt.Errorf("PublishError:TopicNil")
	}
	/*
		var opt PublishOptions
		for _, o := range opts {
			if o == nil {
				return  fmt.Errorf("PublishMsgError:OptsNil")
			}
			o(&opt)
	}*/
	if err := r.declareTopic(topic); err != nil {
		return err
	}
	routing := defaultRouting(topic)
	exchange := defaultExchange(topic)
	if err := r.Publish(exchange, routing, msg, 2); err != nil {
		return err
	}
	return nil
}

/*
  简单消费消息
  参数：
	topic 消费消息的topic
    opts 预留opts操作
*/
func (r *RabbitChannel) DioSubscribe(topic string, autoAck bool, handler Handler, opts ...SubscribeOption) error {
	if topic == "" {
		return fmt.Errorf("SubscribeError:TopicNil")
	}
	/*var opt SubscribeOptions
	for _, o := range opts {
		if o == nil {
			return  fmt.Errorf("ConsumeMsgError:OptsNil")
		}
		o(&opt)
	}*/
	var err error

	if err := r.declareTopic(topic); err != nil {
		return err
	}

	if err := r.declareQueue(topic); err != nil {
		return err
	}

	queue := defaultQueue(topic)
	ch, err := r.Consume(queue, autoAck)

	if err != nil {
		return err
	}

	task := func() {
		for {
			for msg := range ch {
				r.wg.Add(1) // task pair
				msgCopy := msg

				handlerClosure := func() {
					defer func() {
						if r := recover(); r != nil {
							log.Errorf("[Task] handler panic:[%#v] msg:[%#v] err:[%s] ", handler, msgCopy, r)
						}
						r.wg.Done() // task pair
					}()

					if err := handler(msg); err != nil {
						log.Errorf("[Task]handler msg error %v", err)
					}
				}
				handlerClosure()
			}
			if !r.close {
				var retry int
				for {
					retry++
					log.Infof("[Task]chan关闭，重新获取chan,retry count %v", retry)

					if err := r.rePickupChannel(); err != nil {
						log.Errorf("[Task]rePickupChannel error: %v ", err)
						time.Sleep(defaultRetryConsumeDuration)
						continue
					}
					if err := r.reDeclareTopic(topic); err != nil {
						log.Errorf("[Task]reDeclareTopic error: %v", err)
						time.Sleep(defaultRetryConsumeDuration)
						continue
					}
					if err := r.reDeclareQueue(topic); err != nil {
						log.Errorf("[Task]reDeclareQueue error: %v", err)
						time.Sleep(defaultRetryConsumeDuration)
						continue
					}
					ch, err = r.Consume(queue, autoAck)
					if err != nil {
						log.Errorf("[Task]Consume error %v", err)
						time.Sleep(defaultRetryConsumeDuration)
						continue
					}
					log.Debugf("获取新的chan成功")
					break
				}
			} else {
				return
			}
		}

	}

	_ = xprocess.Go(func(ctx context.Context) error {
		task()
		return nil
	})

	return nil
}

func defaultExchange(topic string) string {
	return topic + defaultExchangeSuffix
}

func defaultQueue(topic string) string {
	return topic + defaultQueueSuffix
}

func defaultRouting(topic string) string {
	return topic + defaultRoutingSuffix
}

/* 创建exchange
 */
func (r *RabbitChannel) declareTopic(topic string) error {
	if topic == "" {
		return fmt.Errorf("TopicError:Empty topic:%s ", topic)
	}

	topic = defaultExchange(topic)

	var onceErr error
	oncePtr, _ := initStatusDic.LoadOrStore(topic, &sync.Once{})
	// Declare once
	oncePtr.(*sync.Once).Do(func() {
		if err := r.DeclareExchange(topic, ExchangeKindDirect.String(), true, false); err != nil {
			onceErr = fmt.Errorf("DeclareExchangeError: %v", err.Error())
		}
	})
	if onceErr != nil {
		initStatusDic.Delete(topic)
	}

	return onceErr
}

/*重新创建exchange
 */
func (r *RabbitChannel) reDeclareTopic(topic string) error {
	if topic == "" {
		return fmt.Errorf("TopicError:Empty topic:%s ", topic)
	}

	topic = defaultExchange(topic)

	if err := r.DeclareExchange(topic, ExchangeKindDirect.String(), true, false); err != nil {
		return fmt.Errorf("ReDeclareExchangeError: %v", err.Error())
	}

	return nil
}

/* 创建exchange和queue的bind关系
 */
func (r *RabbitChannel) declareQueue(topic string) error {
	if topic == "" {
		return fmt.Errorf("TopicError:Empty topic:%s ", topic)
	}
	exchange := defaultExchange(topic)
	queue := defaultQueue(topic)
	routing := defaultRouting(topic)

	dictKey := exchange + "." + queue

	var onceErr error
	oncePtr, _ := initStatusDic.LoadOrStore(dictKey, &sync.Once{})
	// Declare once
	oncePtr.(*sync.Once).Do(func() {
		// declare queue name with priority, such as: channelName.Low
		_, err := r.DeclareQueue(queue, true, false)
		if err != nil {
			onceErr = fmt.Errorf("DeclareQueueError:%s QueueName:%s", err.Error(), queue)
			return
		}
		// bind queues with exchange
		if err := r.Bind(queue, routing, exchange); err != nil {
			onceErr = fmt.Errorf("BindQueueError:%s QueueName:%s,Routing:%s,Exchange:%s", err.Error(), queue, routing, exchange)
			return
		}
	})
	if onceErr != nil {
		initStatusDic.Delete(dictKey)
	}
	return onceErr
}

/* 重新创建exchange和queue的bind关系
 */
func (r *RabbitChannel) reDeclareQueue(topic string) error {
	if topic == "" {
		return fmt.Errorf("TopicError:Empty topic:%s ", topic)
	}
	exchange := defaultExchange(topic)
	queue := defaultQueue(topic)
	routing := defaultRouting(topic)

	// declare queue name with priority, such as: channelName.Low
	_, err := r.DeclareQueue(queue, true, false)
	if err != nil {
		return fmt.Errorf("DeclareQueueError:%s QueueName:%s", err.Error(), queue)

	}
	// bind queues with exchange
	if err := r.Bind(queue, routing, exchange); err != nil {
		return fmt.Errorf("BindQueueError:%s QueueName:%s,Routing:%s,Exchange:%s", err.Error(), queue, routing, exchange)

	}
	return nil
}
