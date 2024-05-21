//go:build js
// +build js

package rpc

import (
	"context"
	"net/http"
	"time"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

// WebsocketHandler returns a handler that serves JSON-RPC to WebSocket connections.
//
// allowedOrigins should be a comma-separated list of allowed origin URLs.
// To allow connections with any origin, pass "*".
func (s *Server) WebsocketHandler(allowedOrigins []string) http.Handler {
	panic("not implemented")
}

func newClientTransportWS(endpoint string, cfg *clientConfig) (reconnectFunc, error) {
	dialURL, header, err := wsClientHeaders(endpoint, "")
	if err != nil {
		return nil, err
	}
	for key, values := range cfg.httpHeaders {
		header[key] = values
	}

	connect := func(ctx context.Context) (ServerCodec, error) {
		header := header.Clone()
		if cfg.httpAuth != nil {
			if err := cfg.httpAuth(header); err != nil {
				return nil, err
			}
		}
		conn, resp, err := websocket.Dial(ctx, dialURL, nil)
		if err != nil {
			hErr := wsHandshakeError{err: err}
			if resp != nil {
				hErr.status = resp.Status
			}
			return nil, hErr
		}
		messageSizeLimit := int64(wsDefaultReadLimit)
		if cfg.wsMessageSizeLimit != nil && *cfg.wsMessageSizeLimit >= 0 {
			messageSizeLimit = *cfg.wsMessageSizeLimit
		}
		return newWebsocketCodec(conn, dialURL, header, messageSizeLimit), nil
	}
	return connect, nil
}

type wrappedNhoorConn struct {
	*websocket.Conn
	writeDeadline time.Time
}

func (wc *wrappedNhoorConn) Close() error {
	return wc.Conn.Close(websocket.StatusNormalClosure, "")
}

func (wc *wrappedNhoorConn) SetWriteDeadline(t time.Time) error {
	wc.writeDeadline = t
	return nil
}

type websocketCodecNhooyr struct {
	*jsonCodec
	conn *websocket.Conn
	info PeerInfo
}

func newWebsocketCodec(conn *websocket.Conn, host string, req http.Header, readLimit int64) ServerCodec {
	conn.SetReadLimit(readLimit)
	wrappedConn := &wrappedNhoorConn{Conn: conn}
	encode := func(v interface{}, isErrorResponse bool) error {
		ctx := context.Background()
		return wsjson.Write(ctx, wrappedConn.Conn, v)
	}
	decode := func(v interface{}) error {
		ctx := context.Background()
		return wsjson.Read(ctx, wrappedConn.Conn, v)
	}
	wc := &websocketCodecNhooyr{
		jsonCodec: NewFuncCodec(wrappedConn, encode, decode).(*jsonCodec),
		conn:      conn,
		info: PeerInfo{
			Transport:  "ws",
			RemoteAddr: "Remote address not available",
		},
	}
	// Fill in connection details.
	wc.info.HTTP.Host = host
	wc.info.HTTP.Origin = req.Get("Origin")
	wc.info.HTTP.UserAgent = req.Get("User-Agent")
	return wc
}

func (wc *websocketCodecNhooyr) close() {
	wc.jsonCodec.close()
}

func (wc *websocketCodecNhooyr) peerInfo() PeerInfo {
	return wc.info
}

func (wc *websocketCodecNhooyr) writeJSON(ctx context.Context, v interface{}, isError bool) error {
	return wc.jsonCodec.writeJSON(ctx, v, isError)
}
