package cmdsrv

import (
	"fmt"
	"reflect"

	"github.com/gogf/gf/encoding/gjson"
	"github.com/gogf/gf/util/gvalid"
)

type HandlerFunc = func(*Context)

type Context struct {
	*Response
	SID          string
	Srv          *Srv
	Server       ServerAdapter
	handlers     []HandlerFunc
	handlerIndex int
	handlerAbort bool
}

func (c *Context) Next() {
	if c.handlerAbort {
		return
	}
	if c.handlerIndex < len(c.handlers)-1 {
		c.handlerIndex++
		c.handlers[c.handlerIndex](c)
	} else {
		c.Abort()
	}
}
func (c *Context) Abort() {
	c.handlerAbort = true
}

func (c *Context) Parse(pointer interface{}, mapping ...map[string]string) error {
	var (
		rv = reflect.ValueOf(pointer)
		rk = rv.Kind()
	)
	if rk != reflect.Ptr {
		return fmt.Errorf(
			"parameter should be type of *struct/**struct/*[]struct/*[]*struct/*map[string]interface{}, but got: %v",
			rk,
		)
	}
	rv = rv.Elem()
	rk = rv.Kind()

	// parse request data
	data, err := gjson.LoadContent(c.Response.Request.RawData)
	if err != nil {
		return err
	}

	// if struct or array/slice, validate pointer
	switch rk {
	case reflect.Ptr, reflect.Struct:
		if err := data.GetStruct(".", pointer, mapping...); err != nil {
			return err
		}
		if err := gvalid.CheckStruct(pointer, nil); err != nil {
			return err
		}
	case reflect.Array, reflect.Slice:
		if err := data.GetStructs(".", pointer, mapping...); err != nil {
			return err
		}
		for i := 0; i < rv.Len(); i++ {
			if err := gvalid.CheckStruct(rv.Index(i), nil); err != nil {
				return err
			}
		}
	case reflect.Map:
		if err := data.ToMapToMapDeep(pointer); err != nil {
			return err
		}
	}

	return nil
}

func (c *Context) Get(key string) interface{} {
	return c.Server.GetState(c.SID, key)
}
func (c *Context) Set(key string, v interface{}) {
	c.Server.SetState(c.SID, key, v)
}

func (c *Context) Err(err error, code int) {
	if err != nil {
		c.Response.Code = code
		c.Response.Msg = err.Error()
	}
}

func (c *Context) OK(data ...interface{}) {
	c.Response.Code = 0
	c.Response.Msg = msgOk
	if len(data) == 0 {
		c.Response.Data = struct{}{}
	} else {
		c.Response.Data = data[0]
	}
}
func (c *Context) Resp(code int, msg string, data ...interface{}) {
	c.Response.Code = code
	c.Response.Msg = msg
	if len(data) > 0 {
		c.Response.Data = data[0]
	}
}

func (c *Context) Push(data *Response) error {
	return c.Srv.Push(c.SID, data)
}

func (c *Context) GetAllSID() []string {
	return c.Server.GetAllSID()
}

func (c *Context) Broadcast(data *Response) {
	c.Srv.Broadcast(data)
}

func RouteNotFound(c *Context) {
	c.Resp(-1, msgUnsupportCmd)
}
