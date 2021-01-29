package main

import (
	"hallserver/module/hall"
	"jarvis/base/database"
	"jarvis/base/network"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	// 自定义服务最大连接数量
	CustomMaxConnection = 5000
	// 自定义服务消息管道大小
	CustomIntoStreamSize = 10000
	// Socket 监听地址
	SocketListenAddress = ":8083"
	// WebSocket 监听地址
	WebSocketListenAddress = ":8084"
	// gRPC 监听地址
	GRPCListenAddress = ":8085"
)

var (
	service network.Service
)

func init() {
	// 实例化服务
	service = network.NewService(
		CustomMaxConnection,
		CustomIntoStreamSize,
	)

	// 初始化 MySQL
	if err := database.InitializeMySQL(
		"frank",
		"frank123",
		"localhost",
		7000,
		"jarvis",
	); err != nil {
		log.Panicf("Initialize MySQL error : %s", err.Error())
		return
	}

	// 设置 MySQL
	database.SetUpMySQL(time.Minute*time.Duration(5), 10, 5000)

	// 初始化 Redis
	database.InitializeRedis(time.Minute*time.Duration(5), 10, 5000, "localhost", 8000, "frank123")

	// 初始化 Mongo
	if err := database.InitializeMongo(
		"frank",
		"frank123",
		"jarvis",
		"localhost",
		9000, time.Minute*time.Duration(5), 5000); err != nil {
		log.Panicf("Initialize Mongo error : %s", err.Error())
		return
	}
}

func main() {
	// 1.添加全局中间件
	if err := service.UseMiddleware(logMiddleware); err != nil {
		log.Fatalf("Use middleware error : %s", err)
		return
	}

	// 2.注册观察者
	if err := service.RegisterObserver(hall.NewObserver()); err != nil {
		log.Fatalf("Register observer error : %s", err)
		return
	}

	// 3.注册模块
	if err := service.RegisterModule(hall.NewModule()); err != nil {
		log.Fatalf("Register module error : %s", err)
		return
	}

	// 4.启动
	if err := service.Run(
		network.NewSocketGate(SocketListenAddress),       // Socket 入口
		network.NewWebSocketGate(WebSocketListenAddress), // WebSocket 入口
		network.NewGRPCGate(GRPCListenAddress),           // gRPC 入口
	); err != nil {
		log.Fatalf("Register observer error : %s", err)
		return
	}

	// 5.监听系统信号
	monitorSystemSignal()
}

// 监听系统信号
// kill -SIGQUIT [进程号] : 杀死当前进程
func monitorSystemSignal() {
	sc := make(chan os.Signal)
	signal.Notify(sc, syscall.SIGQUIT)
	select {
	case <-sc:
		log.Println("Done")
	}
}

// 打印中间件
func logMiddleware(ctx network.Context) {
	log.Printf("%+v", ctx.Request())
}
