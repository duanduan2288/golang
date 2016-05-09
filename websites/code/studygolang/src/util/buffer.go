// Copyright 2013 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author：polaris	studygolang@gmail.com

package util

import (
	"bytes"
	"log"
	"strconv"
)

// 内嵌bytes.Buffer，支持连写
type Buffer struct {
	*bytes.Buffer
}

func NewBuffer() *Buffer {
	return &Buffer{Buffer: new(bytes.Buffer)}
}

func (this *Buffer) Append(s string) *Buffer {
	defer func() {
		if err := recover(); err != nil {
			log.Println("*****内存不够了！******")
		}
	}()
	this.Buffer.WriteString(s)
	return this
}

func (this *Buffer) AppendInt(i int) *Buffer {
	return this.Append(strconv.Itoa(i))
}

func (this *Buffer) AppendBytes(p []byte) *Buffer {
	defer func() {
		if err := recover(); err != nil {
			log.Println("*****内存不够了！******")
		}
	}()

	this.Buffer.Write(p)

	return this
}
