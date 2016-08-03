// Implementation of a MultiEchoServer. Students should write their code in this file.

package p0

import (
        "strconv"
        "bufio"
        "net"
        "time"
       )

const sendChannelBufferSize int = 100

type multiEchoServer struct {
    clientsNum      int
    listener        net.Listener
    channelMap      map[net.Conn]chan []byte
    readChan        chan []byte
}

// New creates and returns (but does not start) a new MultiEchoServer.
func New() MultiEchoServer {
    var server multiEchoServer

    server.clientsNum = 0
    server.listener = nil
    server.channelMap = make(map[net.Conn]chan []byte)
    server.readChan = make(chan []byte)

	return &server
}

func (mes *multiEchoServer) Start(port int) error {
    laddr := ":" + strconv.Itoa(port)

    listener,err := net.Listen("tcp",laddr)
    if err != nil{
        return err
    }
    mes.listener = listener
    go mes.handleOutStuff()

    go func(){
        for{
            conn,err := mes.listener.Accept()
            if err != nil{
                return
            }
            mes.clientsNum++
            go mes.readStuff(conn)
        }
    }()

	return nil
}

func (mes *multiEchoServer) Close() {
    if mes.listener != nil{
        mes.listener.Close()
    }
    mes.clientsNum = 0
}

func (mes *multiEchoServer) Count() int {
	return mes.clientsNum
}

func (mes *multiEchoServer)handleOutStuff(){
    for{
        msg := <-mes.readChan
        for _,sendChan := range mes.channelMap{
            if len(sendChan) < sendChannelBufferSize + 10{
                sendChan<-msg
            }else if len(sendChan) < sendChannelBufferSize{
                sendChan<-msg
                time.Sleep(time.Duration(10) * time.Millisecond)
            }
        }
    }
}

func (mes *multiEchoServer)readStuff(conn net.Conn){
    defer conn.Close()
    bufReader := bufio.NewReader(conn)
    mes.channelMap[conn] = make(chan []byte,sendChannelBufferSize)

    go mes.sendStuff(conn)

    for{
        line,err := bufReader.ReadBytes('\n')
        if err != nil{
            delete(mes.channelMap,conn)
            mes.clientsNum--
            break
        }
        mes.readChan<-line
    }
}

func (mes *multiEchoServer)sendStuff(conn net.Conn){
    sendChannel,ok := mes.channelMap[conn]
    if !ok{
        return
    }
    for{
        msg := <-sendChannel
        _,err := conn.Write(msg)
        if err != nil{
            delete(mes.channelMap,conn)
        }
    }
}
