# stomp-ws-go - STOMP 协议 golang 客户端，支持 WebSocket 连接 #

## 说明 ##
本项目在 [stompngo](https://github.com/gmallard/stompngo) 项目的基础上改造而来，增加了对基于 WebSocket 实现的 STOMP 协议支持。

## 特性 ##

- 支持 STOMP 1.0，1.1 和 1.2，相关规范请参考 [STOMP] (http://stomp.github.io/) 官方说明。
- 已针对 ActiveMQ、Apache Apollo、RabbitMQ 最近发布的版本测试通过（基于 TCP 连接），参见 [stompngo](https://github.com/gmallard/stompngo)
- 已针对 Spring STOMP 框架测试通过（基于 WebSocket 连接）

## 安装 ##

```console
go get github.com/drawdy/stomp-ws-go
```

## 示例 ##

- 基于 TCP 连接的示例参见 [stompngo_examples at github](https://github.com/gmallard/stompngo_examples)
- 基于 Websocket 连接的简单示例如下，完整内容参见 [example-stomp-ws-go](https://github.com/drawdy/example-stomp-ws-go/)

```golang
func main() {

	u := url.URL{
		Scheme: "ws",
		Host:   "localhost:8080",
		Path:   "/stomp-ws",
	}
	conn, resp, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatalf("couldn't connect to %v: %v", u.String(), err)
	}
	log.Printf("response status: %v\n", resp.Status)
	log.Print("Websocket connection succeeded.")

	h := stomp.Headers{
		stomp.HK_ACCEPT_VERSION, "1.2,1.1,1.0",
		stomp.HK_HEART_BEAT, "3000,3000",
		stomp.HK_HOST, u.Host,
	}
	sc, err := stomp.ConnectOverWS(conn, h)
	if err != nil {
		log.Fatalf("couldn't create stomp connection: %v", err)
	}

	mdCh, err := sc.Subscribe(stomp.Headers{
		stomp.HK_DESTINATION, "/topic/greeting.back",
		stomp.HK_ID, stomp.Uuid(),
	})
	if err != nil {
		log.Fatalf("failed to suscribe greeting message: %v", err)
	}

	err = sc.Send(stomp.Headers{
		stomp.HK_DESTINATION, "/app/greeting",
		stomp.HK_ID, stomp.Uuid(),
	}, "hello STOMP!")
	if err != nil {
		log.Fatalf("failed to send greeting message: %v", err)
	}

	md := <-mdCh
	if md.Error != nil {
		log.Fatalf("receive greeting message caught error: %v", md.Error)
	}

	fmt.Printf("----> receive new message: %v\n", md.Message.BodyString())

	err = sc.Disconnect(stomp.NoDiscReceipt)
	if err != nil {
		log.Fatalf("failed to disconnect: %v", err)
	}

	log.Print("Disconnected.")
}
```
