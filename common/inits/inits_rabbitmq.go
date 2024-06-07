package inits

import (
	"Bourse/config"
	"fmt"
	"gitee.com/phper95/pkg/logger"
	"github.com/streadway/amqp"
	"strconv"
	"time"
)

type RabbitMQ struct {
	//连接
	Conn *amqp.Connection
	//管道
	Channel *amqp.Channel
	//队列名称
	QueueName string
	//交换机
	Exchange string
	//key Simple模式 几乎用不到
	Key string
	//连接信息
	Mqurl string
}

// NewRabbitMQ 创建RabbitMQ结构体实例
func NewRabbitMQ(queuename string, exchange string, key string) *RabbitMQ {
	rabbitmq := &RabbitMQ{QueueName: queuename, Exchange: exchange, Key: key, Mqurl: config.Cfg.RabbitMQ.MQUrl}
	var err error
	//创建rabbitmq连接
	rabbitmq.Conn, err = amqp.Dial(rabbitmq.Mqurl)
	rabbitmq.failOnErr(err, "创建连接错误！")
	rabbitmq.Channel, err = rabbitmq.Conn.Channel()
	rabbitmq.failOnErr(err, "获取channel失败")
	return rabbitmq
}

// Destory 断开channel和connection
func (r *RabbitMQ) Destory() {
	r.Channel.Close()
	r.Conn.Close()
}

// 错误处理函数
func (r *RabbitMQ) failOnErr(err error, message string) {
	if err != nil {
		logger.Warn(fmt.Sprintf("%s:%s", message, err))
		panic(fmt.Sprintf("%s:%s", message, err))
	}
}

// NewRabbitMQSimple 简单模式step：1。创建简单模式下RabbitMQ实例
func NewRabbitMQSimple(queueName string) *RabbitMQ {
	return NewRabbitMQ(queueName, "", "")
}

// NewRabbitMQPubSub 订阅模式创建rabbitmq实例
func NewRabbitMQPubSub(exchangeName string) *RabbitMQ {
	//创建rabbitmq实例
	rabbitmq := NewRabbitMQ("", exchangeName, "")
	var err error
	//获取connection
	rabbitmq.Conn, err = amqp.Dial(rabbitmq.Mqurl)
	rabbitmq.failOnErr(err, "failed to connecct rabbitmq!")
	//获取channel
	rabbitmq.Channel, err = rabbitmq.Conn.Channel()
	rabbitmq.failOnErr(err, "failed to open a channel!")
	return rabbitmq
}

// PublishPub 订阅模式生成
func (r *RabbitMQ) PublishPub(message string) {
	//尝试创建交换机，不存在创建
	err := r.Channel.ExchangeDeclare(
		//交换机名称
		r.Exchange,
		//交换机类型 广播类型
		"fanout",
		//是否持久化
		true,
		//是否字段删除
		false,
		//true表示这个exchange不可以被client用来推送消息，仅用来进行exchange和exchange之间的绑定
		false,
		//是否阻塞 true表示要等待服务器的响应
		false,
		nil,
	)
	r.failOnErr(err, "failed to declare an excha"+"nge")

	//2 发送消息
	err = r.Channel.Publish(
		r.Exchange,
		"",
		false,
		false,
		amqp.Publishing{
			//类型
			ContentType: "text/plain",
			//消息
			Body: []byte(message),
		})
}

// RecieveSub 订阅模式消费端代码
func (r *RabbitMQ) RecieveSub() {
	//尝试创建交换机，不存在创建
	err := r.Channel.ExchangeDeclare(
		//交换机名称
		r.Exchange,
		//交换机类型 广播类型
		"fanout",
		//是否持久化
		true,
		//是否字段删除
		false,
		//true表示这个exchange不可以被client用来推送消息，仅用来进行exchange和exchange之间的绑定
		false,
		//是否阻塞 true表示要等待服务器的响应
		false,
		nil,
	)
	r.failOnErr(err, "failed to declare an excha"+"nge")
	//2试探性创建队列，创建队列
	q, err := r.Channel.QueueDeclare(
		"", //随机生产队列名称
		false,
		false,
		true,
		false,
		nil,
	)
	r.failOnErr(err, "Failed to declare a queue")
	//绑定队列到exchange中
	err = r.Channel.QueueBind(
		q.Name,
		//在pub/sub模式下，这里的key要为空
		"",
		r.Exchange,
		false,
		nil,
	)
	//消费消息
	message, err := r.Channel.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	forever := make(chan bool)
	go func() {
		for d := range message {
			logger.Warn(fmt.Sprintf("Received a message:%s,", d.Body))
		}
	}()
	fmt.Println("退出请按 Ctrl+C")
	<-forever
}

// NewRabbitMQTopic 话题模式 创建RabbitMQ实例
func NewRabbitMQTopic(exchagne string, routingKey string) *RabbitMQ {
	//创建rabbitmq实例
	rabbitmq := NewRabbitMQ("", exchagne, routingKey)
	var err error
	rabbitmq.Conn, err = amqp.Dial(rabbitmq.Mqurl)
	rabbitmq.failOnErr(err, "failed     to connect rabbingmq!")
	rabbitmq.Channel, err = rabbitmq.Conn.Channel()
	rabbitmq.failOnErr(err, "failed to open a channel")
	return rabbitmq
}

// PublishTopic 话题模式发送信息
func (r *RabbitMQ) PublishTopic(message string) {
	//尝试创建交换机，不存在创建
	err := r.Channel.ExchangeDeclare(
		//交换机名称
		r.Exchange,
		//交换机类型 话题模式
		"topic",
		//是否持久化
		true,
		//是否字段删除
		false,
		//true表示这个exchange不可以被client用来推送消息，仅用来进行exchange和exchange之间的绑定
		false,
		//是否阻塞 true表示要等待服务器的响应
		false,
		nil,
	)
	r.failOnErr(err, "topic failed to declare an excha"+"nge")
	//2发送信息
	err = r.Channel.Publish(
		r.Exchange,
		//要设置
		r.Key,
		false,
		false,
		amqp.Publishing{
			//类型
			ContentType: "text/plain",
			//消息
			Body: []byte(message),
		})
}

// RecieveTopic 话题模式接收信息
// 要注意key
// 其中* 用于匹配一个单词，#用于匹配多个单词（可以是零个）
// 匹配 表示匹配imooc.* 表示匹配imooc.hello,但是imooc.hello.one需要用imooc.#才能匹配到
func (r *RabbitMQ) RecieveTopic() {
	//尝试创建交换机，不存在创建
	err := r.Channel.ExchangeDeclare(
		//交换机名称
		r.Exchange,
		//交换机类型 话题模式
		"topic",
		//是否持久化
		true,
		//是否字段删除
		false,
		//true表示这个exchange不可以被client用来推送消息，仅用来进行exchange和exchange之间的绑定
		false,
		//是否阻塞 true表示要等待服务器的响应
		false,
		nil,
	)
	r.failOnErr(err, "failed to declare an excha"+"nge")
	//2试探性创建队列，创建队列
	q, err := r.Channel.QueueDeclare(
		"", //随机生产队列名称
		false,
		false,
		true,
		false,
		nil,
	)
	r.failOnErr(err, "Failed to declare a queue")
	//绑定队列到exchange中
	err = r.Channel.QueueBind(
		q.Name,
		//在pub/sub模式下，这里的key要为空
		r.Key,
		r.Exchange,
		false,
		nil,
	)
	//消费消息
	message, err := r.Channel.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	forever := make(chan bool)
	go func() {
		for d := range message {
			logger.Warn(fmt.Sprintf("Received a message:%s,", d.Body))
		}
	}()
	fmt.Println("退出请按 Ctrl+C")
	<-forever
}

// NewRabbitMQRouting 路由模式 创建RabbitMQ实例
func NewRabbitMQRouting(exchagne string, routingKey string) *RabbitMQ {
	//创建rabbitmq实例
	rabbitmq := NewRabbitMQ("", exchagne, routingKey)
	var err error
	rabbitmq.Conn, err = amqp.Dial(rabbitmq.Mqurl)
	rabbitmq.failOnErr(err, "failed     to connect rabbingmq!")
	rabbitmq.Channel, err = rabbitmq.Conn.Channel()
	rabbitmq.failOnErr(err, "failed to open a channel")
	return rabbitmq
}

// PublishRouting 路由模式发送信息
func (r *RabbitMQ) PublishRouting(message string) {
	//尝试创建交换机，不存在创建
	err := r.Channel.ExchangeDeclare(
		//交换机名称
		r.Exchange,
		//交换机类型 广播类型
		"direct",
		//是否持久化
		true,
		//是否字段删除
		false,
		//true表示这个exchange不可以被client用来推送消息，仅用来进行exchange和exchange之间的绑定
		false,
		//是否阻塞 true表示要等待服务器的响应
		false,
		nil,
	)
	r.failOnErr(err, "failed to declare an excha"+"nge")
	//发送信息
	err = r.Channel.Publish(
		r.Exchange,
		//要设置
		r.Key,
		false,
		false,
		amqp.Publishing{
			//类型
			ContentType: "text/plain",
			//消息
			Body: []byte(message),
		})
}

// RecieveRouting 路由模式接收信息
func (r *RabbitMQ) RecieveRouting() {
	//尝试创建交换机，不存在创建
	err := r.Channel.ExchangeDeclare(
		//交换机名称
		r.Exchange,
		//交换机类型 广播类型
		"direct",
		//是否持久化
		true,
		//是否字段删除
		false,
		//true表示这个exchange不可以被client用来推送消息，仅用来进行exchange和exchange之间的绑定
		false,
		//是否阻塞 true表示要等待服务器的响应
		false,
		nil,
	)
	r.failOnErr(err, "failed to declare an excha"+"nge")
	//2试探性创建队列，创建队列
	q, err := r.Channel.QueueDeclare(
		"", //随机生产队列名称
		false,
		false,
		true,
		false,
		nil,
	)
	r.failOnErr(err, "Failed to declare a queue")
	//绑定队列到exchange中
	err = r.Channel.QueueBind(
		q.Name,
		//在pub/sub模式下，这里的key要为空
		r.Key,
		r.Exchange,
		false,
		nil,
	)
	//消费消息
	message, err := r.Channel.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	forever := make(chan bool)
	go func() {
		for d := range message {
			logger.Warn(fmt.Sprintf("Received a message:%s,", d.Body))
		}
	}()
	fmt.Println("退出请按 Ctrl+C")
	<-forever
}

// PublishSimple 简单模式Step:2、简单模式下生产代码
func (r *RabbitMQ) PublishSimple(message string) {
	//1、申请队列，如果队列存在就跳过，不存在创建
	//优点：保证队列存在，消息能发送到队列中
	_, err := r.Channel.QueueDeclare(
		//队列名称
		r.QueueName,
		//是否持久化
		true,
		//是否为自动删除 当最后一个消费者断开连接之后，是否把消息从队列中删除
		false,
		//是否具有排他性 true表示自己可见 其他用户不能访问
		false,
		//是否阻塞 true表示要等待服务器的响应
		false,
		//额外数学系
		nil,
	)
	if err != nil {
		fmt.Println(err)
	}

	// exchange 绑定 queue
	//r.Channel.QueueBind(r.QueueName, routing_key, r.Exchange, false, nil)

	//2.发送消息到队列中
	r.Channel.Publish(
		//默认的Exchange交换机是default,类型是direct直接类型
		r.Exchange,
		//要赋值的队列名称
		r.QueueName,
		//如果为true，根据exchange类型和routkey规则，如果无法找到符合条件的队列那么会把发送的消息返回给发送者
		false,
		//如果为true,当exchange发送消息到队列后发现队列上没有绑定消费者，则会把消息还给发送者
		false,
		//消息
		amqp.Publishing{
			//类型
			ContentType: "text/plain",
			//消息
			Body: []byte(message),
		})
}

func (r *RabbitMQ) ConsumeSimple() {
	//1、申请队列，如果队列存在就跳过，不存在创建
	//优点：保证队列存在，消息能发送到队列中
	_, err := r.Channel.QueueDeclare(
		//队列名称
		r.QueueName,
		//是否持久化
		false,
		//是否为自动删除 当最后一个消费者断开连接之后，是否把消息从队列中删除
		false,
		//是否具有排他性
		false,
		//是否阻塞
		false,
		//额外数学系
		nil,
	)
	if err != nil {
		fmt.Println(err)
	}
	//接收消息
	msgs, err := r.Channel.Consume(
		r.QueueName,
		//用来区分多个消费者
		"",
		//是否自动应答
		true,
		//是否具有排他性
		false,
		//如果设置为true,表示不能同一个connection中发送的消息传递给这个connection中的消费者
		false,
		//队列是否阻塞
		false,
		nil,
	)
	if err != nil {
		fmt.Println(err)
	}
	forever := make(chan bool)

	//启用协程处理
	go func() {
		for d := range msgs {
			//实现我们要处理的逻辑函数
			logger.Warn(fmt.Sprintf("Received a message:%s", d.Body))
			fmt.Println(d.Body)
		}
	}()

	logger.Warn(fmt.Sprintf("【*】warting for messages, To exit press CCTRAL+C"))
	<-forever
}

// 使用示例
func example() {
	// =========== Simple模式 发送者 ===========
	rabbitmq := NewRabbitMQSimple("imoocSimple")
	rabbitmq.PublishSimple("hello imooc!")
	//接收者
	rabbitmq = NewRabbitMQSimple("imoocSimple")
	rabbitmq.ConsumeSimple()

	// =========== 订阅模式发送者 ===========
	rabbitmq = NewRabbitMQPubSub("" + "newProduct")
	for i := 0; i <= 100; i++ {
		rabbitmq.PublishPub("订阅模式生产第" + strconv.Itoa(i) + "条数据")
		fmt.Println(i)
		time.Sleep(1 * time.Second)
	}
	//接收者
	rabbitmq = NewRabbitMQPubSub("" + "newProduct")
	rabbitmq.RecieveSub()

	// =========== 路由模式发送者 ===========
	imoocOne := NewRabbitMQRouting("exImooc", "imooc_one")
	imoocTwo := NewRabbitMQRouting("exImooc", "imooc_two")

	for i := 0; i <= 10; i++ {
		imoocOne.PublishRouting("hello imooc one!" + strconv.Itoa(i))
		imoocTwo.PublishRouting("hello imooc two!" + strconv.Itoa(i))
		time.Sleep(1 * time.Second)
		fmt.Println(i)
	}
	//接收者
	rabbitmq = NewRabbitMQRouting("exImooc", "imooc_one")
	rabbitmq.RecieveRouting()

	// =========== Topic模式发送者 ===========
	imoocOne = NewRabbitMQTopic("exImoocTopic", "imooc.topic88.three")
	imoocTwo = NewRabbitMQTopic("exImoocTopic", "imooc.topic88.four")

	for i := 0; i <= 10; i++ {
		imoocOne.PublishTopic("hello imooc topic three!" + strconv.Itoa(i))
		imoocTwo.PublishTopic("hello imooc topic four!" + strconv.Itoa(i))
		time.Sleep(1 * time.Second)
		fmt.Println(i)
	}
	//Topic接收者
	rabbitmq = NewRabbitMQTopic("exImoocTopic", "#")
	rabbitmq.RecieveTopic()
}
