package common

import (
	"bufio"
	"log"
	"net"
)

/*
客户端结构体
*/
type Client struct {
	Key    string        //客户端连接的唯标志
	Conn   net.Conn      //连接
	In     InMessage     //输入消息
	Out    OutMessage    //输出消息
	Reader *bufio.Reader //读取
	Writer *bufio.Writer //输出
	Quit   chan *Client  //退出
}
type ClientTable map[string]*Client //客户端列表
/*
获取客户端名称
*/
func (this *Client) GetKey() string {
	return this.Key
}

/*
设置客户端名称
*/
func (this *Client) SetKey(key string) {
	this.Key = key
}

/*
获取输入消息
*/
func (this *Client) GetIn() IMRequest {
	return <-this.In
}

/*
设置输出消息
*/
func (this *Client) PutOut(resp *IMResponse) {
	this.Out <- *resp
}

/*
创建客户端
*/
func CreateClient(key string, conn net.Conn) *Client {
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)
	client := &Client{
		Key:    key,
		Conn:   conn,
		In:     make(InMessage),
		Out:    make(OutMessage),
		Quit:   make(chan *Client),
		Reader: reader,
		Writer: writer,
	}
	client.Listen()
	return client
}

/*
自动读入或者写出消息
*/
func (this *Client) Listen() {
	go this.read()
	go this.write()
}

/*
退出了一个连接
*/
func (this *Client) Quiting() {
	this.Quit <- this
}

/*
关闭连接通道
*/
func (this *Client) Close() {
	this.Conn.Close()
}

/*
读取消息
*/
func (this *Client) read() {
	for {
		if line, _, err := this.Reader.ReadLine(); err == nil {
			req, err := DecodeIMRequest(line)
			if err == nil {
				req.Client = this
				this.In <- *req
			} else {
				// 忽略消息，连命令都不知道，没办法处理
				log.Printf("解析JSON错误: %s", line)
			}
		} else {
			log.Printf("Read error: %s\n", err)
			this.Quiting()
			return
		}
	}
}

/*
输出消息
*/
func (this *Client) write() {
	for resp := range this.Out {
		if _, err := this.Writer.WriteString(string(resp.Encode()) + "\n"); err != nil {
			this.Quiting()
			return
		}
		if err := this.Writer.Flush(); err != nil {
			log.Printf("Write error: %s\n", err)
			this.Quiting()
			return
		}
	}
}