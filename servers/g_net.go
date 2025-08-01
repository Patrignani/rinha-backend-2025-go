package servers

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Patrignani/patrignani-rinha-backend-go/internal/services"
	"github.com/panjf2000/gnet/v2"
)

var (
	bufPool = sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}
	prefix = []byte("HTTP/1.1 200 OK\r\nContent-Type: application/json\r\nContent-Length: 0\r\nConnection: keep-alive\r\n\r\n")
)

type GNetServer struct {
	*gnet.BuiltinEventEngine
	paymentService *services.PaymentService
	keepAlive      bool
}

func NewGNetServer(paymentService *services.PaymentService, keepAlive bool) *GNetServer {
	return &GNetServer{paymentService: paymentService, keepAlive: keepAlive}
}

func writeResponse(c gnet.Conn, statusCode int, body []byte, keepAlive bool) {
	statusText := "OK"
	if statusCode != 200 {
		statusText = "Error"
	}

	connHdr := "Connection: keep-alive\r\n"
	if !keepAlive {
		connHdr = "Connection: close\r\n"
	}

	buf := bufPool.Get().(*bytes.Buffer)
	buf.Reset()

	buf.WriteString(fmt.Sprintf("HTTP/1.1 %d %s\r\n", statusCode, statusText))
	buf.WriteString("Content-Type: application/json\r\n")
	buf.WriteString(fmt.Sprintf("Content-Length: %d\r\n", len(body)))
	buf.WriteString(connHdr)
	buf.WriteString("\r\n")
	buf.Write(body)

	_ = c.AsyncWrite(buf.Bytes(), nil)

	bufPool.Put(buf)
}

func sendWithBlockingWrite(c gnet.Conn, keepAlive bool) gnet.Action {
	if _, err := c.Write(prefix); err != nil {
		println(err.Error())
	}

	if !keepAlive {
		return gnet.Close
	}
	return gnet.None
}

func readLine(data []byte) (line, rest []byte, ok bool) {
	idx := bytes.Index(data, []byte("\r\n"))
	if idx < 0 {
		return nil, data, false
	}
	return data[:idx], data[idx+2:], true
}

func readHeaders(data []byte) (headers map[string][]byte, rest []byte, ok bool) {
	headers = make(map[string][]byte)
	for {
		var line []byte
		line, data, ok = readLine(data)
		if !ok {
			return nil, data, false
		}
		if len(line) == 0 {
			return headers, data, true
		}
		sep := bytes.IndexByte(line, ':')
		if sep < 0 {
			return nil, data, false
		}
		key := bytes.ToLower(bytes.TrimSpace(line[:sep]))
		val := bytes.TrimSpace(line[sep+1:])
		headers[string(key)] = val
	}
}

func parseQueryString(qs []byte) (map[string]string, bool) {
	result := make(map[string]string)
	pairs := bytes.Split(qs, []byte("&"))
	if len(pairs) != 2 {
		return nil, false
	}
	for _, p := range pairs {
		kv := bytes.SplitN(p, []byte("="), 2)
		if len(kv) != 2 {
			return nil, false
		}
		result[string(kv[0])] = string(kv[1])
	}
	return result, true
}

func (s *GNetServer) OnTraffic(c gnet.Conn) gnet.Action {
	for {
		buf, _ := c.Peek(-1)
		if len(buf) == 0 {
			return gnet.None
		}

		reqLine, rest, ok := readLine(buf)
		if !ok {
			return gnet.None
		}

		parts := bytes.Split(reqLine, []byte(" "))
		if len(parts) < 3 {
			writeResponse(c, 400, []byte(`{"error":"bad request"}`), false)
			return gnet.Close
		}

		method := string(parts[0])
		path := parts[1]

		headers, rest, ok := readHeaders(rest)
		if !ok {
			return gnet.None
		}

		cl := 0
		if val, exists := headers["content-length"]; exists {
			var err error
			cl, err = strconv.Atoi(string(val))
			if err != nil || cl < 0 {
				writeResponse(c, 400, []byte(`{"error":"bad content-length"}`), false)
				return gnet.Close
			}
		}

		if len(rest) < cl {
			return gnet.None
		}

		totalConsumed := len(reqLine) + 2
		for k, v := range headers {
			totalConsumed += len(k) + len(v) + 4
		}
		totalConsumed += 2
		totalConsumed += cl

		if val, exists := headers["connection"]; exists {
			if bytes.EqualFold(val, []byte("close")) {
				s.keepAlive = false
			}
		}

		bodyStart := len(buf) - len(rest)
		body := buf[bodyStart : bodyStart+cl]

		_, _ = c.Discard(totalConsumed)

		if method == "GET" {
			partsPath := bytes.SplitN(path, []byte("?"), 2)
			route := string(partsPath[0])

			if route == "/payments-summary" {

				if len(partsPath) < 2 {
					writeResponse(c, 400, []byte(`{"error":"missing query"}`), s.keepAlive)
					if !s.keepAlive {
						return gnet.Close
					}
					continue
				}

				queryMap, ok := parseQueryString(partsPath[1])
				if !ok {
					writeResponse(c, 400, []byte(`{"error":"invalid query"}`), s.keepAlive)
					if !s.keepAlive {
						return gnet.Close
					}
					continue
				}

				fromStr := queryMap["from"]
				toStr := queryMap["to"]

				var fromTime, toTime *time.Time

				if fromStr != "" {
					t, err := time.Parse(time.RFC3339, fromStr)
					if err != nil {
						writeResponse(c, 400, []byte(`{"error":"invalid 'from' timestamp format"}`), s.keepAlive)
						if !s.keepAlive {
							return gnet.Close
						}
						continue
					}
					fromTime = &t
				}

				if toStr != "" {
					toStr = strings.TrimRight(toStr, "\\")
					t, err := time.Parse(time.RFC3339, toStr)
					if err != nil {
						writeResponse(c, 400, []byte(`{"error":"invalid 'to' timestamp format"}`), s.keepAlive)
						if !s.keepAlive {
							return gnet.Close
						}
						continue
					}
					toTime = &t
				}

				v, err := s.paymentService.GetPaymentSummary(context.TODO(), fromTime, toTime)

				if err != nil {
					writeResponse(c, 400, []byte(fmt.Sprintf(`{"error":"%v"}`, err)), s.keepAlive)
					if !s.keepAlive {
						return gnet.Close
					}
					continue
				}

				jsonBytes, err := v.MarshalJSON()
				if err != nil {
					writeResponse(c, 400, []byte(fmt.Sprintf(`{"error":"%v"}`, err)), s.keepAlive)
					if !s.keepAlive {
						return gnet.Close
					}
					continue
				}

				writeResponse(c, 200, []byte(jsonBytes), s.keepAlive)
				if !s.keepAlive {
					return gnet.Close
				}
			} else {
				println(fmt.Sprintf("ROTA NÃO EXISTE: %s", route))
				writeResponse(c, 400, []byte(`{"error":"not found"}`), s.keepAlive)
				if !s.keepAlive {
					return gnet.Close
				}
				continue
			}

		} else if method == "POST" {
			if !bytes.Equal(path, []byte("/payments")) {
				writeResponse(c, 404, []byte(`{"error":"not found"}`), s.keepAlive)
				if !s.keepAlive {
					return gnet.Close
				}
				continue
			}

			sendWithBlockingWrite(c, s.keepAlive)
			go s.paymentService.RunQueue(context.TODO(), body)

			if !s.keepAlive {
				return gnet.Close
			}

		} else {
			writeResponse(c, 405, []byte(`{"error":"method not allowed"}`), s.keepAlive)
			if !s.keepAlive {
				return gnet.Close
			}
		}
	}
}

func (s *GNetServer) OnOpened(c gnet.Conn) (out []byte, action gnet.Action) {
	_ = c.SetDeadline(time.Now().Add(60 * time.Second))
	return nil, gnet.None
}
