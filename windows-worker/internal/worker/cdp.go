package worker

import (
	"bufio"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type cdpLogger func(format string, args ...any)

type cdpClient struct {
	conn   net.Conn
	reader *bufio.Reader
	nextID int
}

func loadExtensionAndNavigate(profileDir, extensionDir, agentURL string, log cdpLogger) error {
	client, err := connectCDP(profileDir, 10*time.Second)
	if err != nil {
		return err
	}
	defer client.close()

	result, err := client.call("", "Extensions.loadUnpacked", map[string]any{
		"path":              extensionDir,
		"enableInIncognito": false,
	}, 8*time.Second)
	if err != nil {
		return err
	}
	var loadResult struct {
		ID string `json:"id"`
	}
	if err = json.Unmarshal(result, &loadResult); err == nil && loadResult.ID != "" {
		log("cdp loaded Automa extension id=%s path=%q", loadResult.ID, extensionDir)
	}
	return nil
}

func navigateBlankTarget(client *cdpClient, agentURL string) error {
	result, err := client.call("", "Target.getTargets", map[string]any{}, 5*time.Second)
	if err != nil {
		return err
	}
	var targets struct {
		TargetInfos []struct {
			TargetID string `json:"targetId"`
			Type     string `json:"type"`
			URL      string `json:"url"`
		} `json:"targetInfos"`
	}
	if err = json.Unmarshal(result, &targets); err != nil {
		return err
	}

	var targetID string
	for _, target := range targets.TargetInfos {
		if target.Type != "page" {
			continue
		}
		if target.URL == "" || target.URL == "about:blank" || strings.HasPrefix(target.URL, "chrome://newtab") {
			targetID = target.TargetID
			break
		}
	}
	if targetID == "" {
		_, err = client.call("", "Target.createTarget", map[string]any{"url": agentURL}, 5*time.Second)
		return err
	}

	result, err = client.call("", "Target.attachToTarget", map[string]any{
		"targetId": targetID,
		"flatten":  true,
	}, 5*time.Second)
	if err != nil {
		return err
	}
	var attachResult struct {
		SessionID string `json:"sessionId"`
	}
	if err = json.Unmarshal(result, &attachResult); err != nil {
		return err
	}
	if attachResult.SessionID == "" {
		return errors.New("CDP attachToTarget did not return sessionId")
	}

	_, err = client.call(attachResult.SessionID, "Page.navigate", map[string]any{"url": agentURL}, 5*time.Second)
	return err
}

func connectCDP(profileDir string, timeout time.Duration) (*cdpClient, error) {
	port, wsPath, err := waitDevToolsActivePort(profileDir, timeout)
	if err != nil {
		return nil, err
	}

	address := net.JoinHostPort("127.0.0.1", port)
	conn, err := net.DialTimeout("tcp", address, 5*time.Second)
	if err != nil {
		return nil, err
	}

	keyBytes := make([]byte, 16)
	if _, err = rand.Read(keyBytes); err != nil {
		_ = conn.Close()
		return nil, err
	}
	key := base64.StdEncoding.EncodeToString(keyBytes)
	request := fmt.Sprintf("GET %s HTTP/1.1\r\nHost: %s\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Key: %s\r\nSec-WebSocket-Version: 13\r\n\r\n", wsPath, address, key)
	if _, err = conn.Write([]byte(request)); err != nil {
		_ = conn.Close()
		return nil, err
	}

	reader := bufio.NewReader(conn)
	status, err := reader.ReadString('\n')
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	if !strings.Contains(status, " 101 ") {
		_ = conn.Close()
		return nil, fmt.Errorf("CDP websocket handshake failed: %s", strings.TrimSpace(status))
	}
	for {
		line, readErr := reader.ReadString('\n')
		if readErr != nil {
			_ = conn.Close()
			return nil, readErr
		}
		if strings.TrimSpace(line) == "" {
			break
		}
	}

	return &cdpClient{conn: conn, reader: reader}, nil
}

func waitDevToolsActivePort(profileDir string, timeout time.Duration) (string, string, error) {
	activePortPath := filepath.Join(profileDir, "DevToolsActivePort")
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		body, err := os.ReadFile(activePortPath)
		if err == nil {
			lines := strings.Split(strings.TrimSpace(string(body)), "\n")
			if len(lines) >= 2 {
				port := strings.TrimSpace(lines[0])
				wsPath := strings.TrimSpace(lines[1])
				if port != "" && wsPath != "" {
					return port, wsPath, nil
				}
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	return "", "", fmt.Errorf("waiting for %s timed out", activePortPath)
}

func (c *cdpClient) call(sessionID, method string, params map[string]any, timeout time.Duration) (json.RawMessage, error) {
	c.nextID++
	id := c.nextID
	message := map[string]any{
		"id":     id,
		"method": method,
	}
	if sessionID != "" {
		message["sessionId"] = sessionID
	}
	if params != nil {
		message["params"] = params
	}
	body, err := json.Marshal(message)
	if err != nil {
		return nil, err
	}
	if err = c.writeText(body); err != nil {
		return nil, err
	}

	deadline := time.Now().Add(timeout)
	for {
		if err = c.conn.SetReadDeadline(deadline); err != nil {
			return nil, err
		}
		payload, err := c.readText()
		if err != nil {
			return nil, err
		}
		var response struct {
			ID     int             `json:"id"`
			Result json.RawMessage `json:"result"`
			Error  *struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
			} `json:"error"`
		}
		if err = json.Unmarshal(payload, &response); err != nil {
			return nil, err
		}
		if response.ID != id {
			continue
		}
		if response.Error != nil {
			return nil, fmt.Errorf("CDP %s failed: %s", method, response.Error.Message)
		}
		return response.Result, nil
	}
}

func (c *cdpClient) writeText(payload []byte) error {
	header := []byte{0x81}
	length := len(payload)
	switch {
	case length < 126:
		header = append(header, byte(0x80|length))
	case length <= 0xffff:
		header = append(header, 0x80|126, byte(length>>8), byte(length))
	default:
		header = append(header, 0x80|127)
		size := make([]byte, 8)
		binary.BigEndian.PutUint64(size, uint64(length))
		header = append(header, size...)
	}

	mask := make([]byte, 4)
	if _, err := rand.Read(mask); err != nil {
		return err
	}
	header = append(header, mask...)
	masked := make([]byte, len(payload))
	for index, value := range payload {
		masked[index] = value ^ mask[index%4]
	}

	if _, err := c.conn.Write(header); err != nil {
		return err
	}
	_, err := c.conn.Write(masked)
	return err
}

func (c *cdpClient) readText() ([]byte, error) {
	for {
		header := make([]byte, 2)
		if _, err := io.ReadFull(c.reader, header); err != nil {
			return nil, err
		}
		opcode := header[0] & 0x0f
		length := uint64(header[1] & 0x7f)
		switch length {
		case 126:
			size := make([]byte, 2)
			if _, err := io.ReadFull(c.reader, size); err != nil {
				return nil, err
			}
			length = uint64(binary.BigEndian.Uint16(size))
		case 127:
			size := make([]byte, 8)
			if _, err := io.ReadFull(c.reader, size); err != nil {
				return nil, err
			}
			length = binary.BigEndian.Uint64(size)
		}

		masked := header[1]&0x80 != 0
		var mask []byte
		if masked {
			mask = make([]byte, 4)
			if _, err := io.ReadFull(c.reader, mask); err != nil {
				return nil, err
			}
		}

		payload := make([]byte, length)
		if _, err := io.ReadFull(c.reader, payload); err != nil {
			return nil, err
		}
		if masked {
			for index := range payload {
				payload[index] ^= mask[index%4]
			}
		}

		switch opcode {
		case 0x1:
			return payload, nil
		case 0x8:
			return nil, errors.New("CDP websocket closed")
		case 0x9:
			_ = c.writePong(payload)
		}
	}
}

func (c *cdpClient) writePong(payload []byte) error {
	frame := []byte{0x8a, byte(0x80 | len(payload))}
	mask := make([]byte, 4)
	if _, err := rand.Read(mask); err != nil {
		return err
	}
	frame = append(frame, mask...)
	for index, value := range payload {
		frame = append(frame, value^mask[index%4])
	}
	_, err := c.conn.Write(frame)
	return err
}

func (c *cdpClient) close() {
	_ = c.conn.Close()
}
