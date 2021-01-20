package xtcp

import (
	"bytes"
	"encoding/binary"
)

// DefaultPkgProto 一个默认的私有协议实现
// 协议组成：
// 4字节(自定义数据长度) + 任意字节(json字符串数据)
type DefaultPkgProto struct {
	PoolBuf map[string][]byte
}

// Packer 封包，将数据区域包装成私有协议数据包
func (*DefaultPkgProto) Packer(data []byte) ([]byte, error) {
	bodyLen := uint32(len(data))
	header := make([]byte, 4)
	binary.BigEndian.PutUint32(header, bodyLen)

	pkg := bytesCombine(header, data)
	return pkg, nil
}

// Parser 解包，解析收到的原始数据，原始数据有粘包和半包，返回解析完成的数据
func (p *DefaultPkgProto) Parser(sid string, bt []byte) ([][]byte, error) {
	if p.PoolBuf == nil {
		p.PoolBuf = make(map[string][]byte)
	}
	preBuf, ok := p.PoolBuf[sid]
	if !ok {
		preBuf = make([]byte, 0)
		p.PoolBuf[sid] = preBuf
	}

	buf := bytesCombine(preBuf, bt)
	datas := make([][]byte, 0)

	for {
		if len(buf) < 4 {
			break
		}
		header := buf[:4]
		bodyLen := binary.BigEndian.Uint32(header)
		if uint32(len(buf)) < (4 + bodyLen) {
			break
		}
		pack := buf[4 : 4+bodyLen]
		buf = buf[4+bodyLen:]
		datas = append(datas, pack)
	}
	p.PoolBuf[sid] = buf

	return datas, nil

}

func bytesCombine(pBytes ...[]byte) []byte {
	return bytes.Join(pBytes, []byte(""))
}
