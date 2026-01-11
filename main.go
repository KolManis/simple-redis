package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
)

// *3\r\n      ← массив из 3 элементов
// $3\r\nSET\r\n     ← bulk string "SET" (3 символа)
// $5\r\nmykey\r\n   ← bulk string "mykey" (5 символов)
// $5\r\nhello\r\n   ← bulk string "hello" (5 символов)
type ValueType string

const (
	ARRAY  ValueType = "*" // *3\r\n...      - массив команд
	BULK   ValueType = "$" // $5\r\nhello\r\n - строки произвольной длины
	STRING ValueType = "+" // +OK\r\n        - простые строки
)

type Value struct {
	typ   ValueType
	bulk  string
	str   string
	array []Value
}

func (v *Value) readArray(reader io.Reader) {
	buf := make([]byte, 4)
	reader.Read(buf)
	fmt.Println(string(buf))
	arrLen, err := strconv.Atoi(string(buf[1]))
	if err != nil {
		fmt.Println(err)
		return
	}

	for range arrLen {
		bulk := v.readBulk(reader)
		v.array = append(v.array, bulk)
	}
}
func (v *Value) readBulk(reader io.Reader) Value {
	buf := make([]byte, 4)
	reader.Read(buf)

	n, err := strconv.Atoi(string(buf[1]))
	if err != nil {
		fmt.Println(err)
		return Value{}
	}

	bulkBuf := make([]byte, n+2)
	reader.Read(bulkBuf)
	bulk := string(bulkBuf[:n])
	return Value{typ: BULK, bulk: bulk}
}

func main() {
	l, err := net.Listen("tcp", ":6379")
	if err != nil {
		log.Fatal("cannot listen on :6379")
	}
	defer l.Close()

	conn, err := l.Accept()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer conn.Close()

	for {
		v := Value{typ: ARRAY}
		v.readArray(conn)
		conn.Write([]byte("+OK\r\n"))
	}
}
