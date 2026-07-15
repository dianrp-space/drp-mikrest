package routeros

import (
	"errors"
	"fmt"
	"strings"

	ros "github.com/go-routeros/routeros/v3"
	"github.com/go-routeros/routeros/v3/proto"
)

// Client membungkus koneksi RouterOS API native (port 8728 default).
type Client struct {
	addr string // host:port
	user string
	pass string
	c    *ros.Client
}

// Dial membuka koneksi + login ke perangkat RouterOS.
func Dial(host string, port int, user, pass string) (*Client, error) {
	addr := fmt.Sprintf("%s:%d", host, port)
	c, err := ros.Dial(addr, user, pass)
	if err != nil {
		return nil, fmt.Errorf("dial routeros %s: %w", addr, err)
	}
	return &Client{addr: addr, user: user, pass: pass, c: c}, nil
}

// Close menutup koneksi.
func (cl *Client) Close() {
	if cl != nil && cl.c != nil {
		cl.c.Close()
	}
}

// Run menjalankan sentence dan mengembalikan daftar sentence reply.
func (cl *Client) Run(sentence ...string) ([]*proto.Sentence, error) {
	if cl == nil || cl.c == nil {
		return nil, errors.New("klien routeros belum terhubung")
	}
	reply, err := cl.c.RunArgs(sentence)
	if err != nil {
		return nil, err
	}
	return reply.Re, nil
}

// RunMap menjalankan sentence dan mengembalikan list map per baris reply.
func (cl *Client) RunMap(sentence ...string) ([]map[string]string, error) {
	sentences, err := cl.Run(sentence...)
	if err != nil {
		return nil, err
	}
	out := make([]map[string]string, 0, len(sentences))
	for _, s := range sentences {
		out = append(out, sentenceToMap(s))
	}
	return out, nil
}

// Add menjalankan perintah add dengan pasangan key=value.
// Contoh: Add("/ip/hotspot/user/add", "name", "AB12", "password", "AB12", "profile", "1h")
func (cl *Client) Add(command string, kv ...string) (map[string]string, error) {
	if len(kv)%2 != 0 {
		return nil, errors.New("kv harus pasangan key-value")
	}
	args := []string{command}
	for i := 0; i < len(kv); i += 2 {
		args = append(args, fmt.Sprintf("=%s=%s", kv[i], kv[i+1]))
	}
	rows, err := cl.RunMap(args...)
	if err != nil {
		return nil, err
	}
	if len(rows) > 0 {
		return rows[0], nil
	}
	return map[string]string{}, nil
}

// Set menjalankan perintah set berdasarkan .id.
func (cl *Client) Set(command, id string, kv ...string) error {
	if len(kv)%2 != 0 {
		return errors.New("kv harus pasangan key-value")
	}
	args := []string{command, "=.id=" + id}
	for i := 0; i < len(kv); i += 2 {
		args = append(args, fmt.Sprintf("=%s=%s", kv[i], kv[i+1]))
	}
	_, err := cl.Run(args...)
	return err
}

// Remove menghapus item berdasarkan .id.
func (cl *Client) Remove(command, id string) error {
	_, err := cl.Run(command, "=.id="+id)
	return err
}

// Print mengembalikan semua baris dari perintah print.
func (cl *Client) Print(command string, kv ...string) ([]map[string]string, error) {
	args := []string{command + "/print"}
	for i := 0; i+1 < len(kv); i += 2 {
		args = append(args, fmt.Sprintf("?%s=%s", kv[i], kv[i+1]))
	}
	return cl.RunMap(args...)
}

// PrintQuery print dengan filter ?field=value
func (cl *Client) PrintQuery(command string, filters ...string) ([]map[string]string, error) {
	return cl.Print(command, filters...)
}

func sentenceToMap(s *proto.Sentence) map[string]string {
	m := make(map[string]string, len(s.List))
	for _, p := range s.List {
		m[p.Key] = p.Value
	}
	return m
}

// Ping memastikan koneksi valid dengan menjalankan /system/identity/print.
func (cl *Client) Ping() (string, error) {
	rows, err := cl.Print("/system/identity")
	if err != nil {
		return "", err
	}
	if len(rows) == 0 {
		return "", nil
	}
	return rows[0]["name"], nil
}

// Exec menjalankan command RouterOS mentah dan mengembalikan output terminasi
// dalam bentuk teks (seperti terminal CLI).
// Contoh: Exec("/ip/hotspot/active/print count-only=yes")
func (cl *Client) Exec(command string) (string, error) {
	args := strings.Fields(command)
	if len(args) == 0 {
		return "", errors.New("command kosong")
	}
	rows, err := cl.RunMap(args...)
	if err != nil {
		return "", err
	}

	var buf strings.Builder
	if len(rows) == 0 {
		buf.WriteString("-- empty --\n")
		return buf.String(), nil
	}

	// Kumpulkan semua key yang mungkin muncul
	keys := make(map[string]bool)
	for _, row := range rows {
		for k := range row {
			keys[k] = true
		}
	}
	keyList := make([]string, 0, len(keys))
	for k := range keys {
		keyList = append(keyList, k)
	}

	for _, row := range rows {
		for _, k := range keyList {
			if v, ok := row[k]; ok {
				buf.WriteString(fmt.Sprintf("  %s: %s\n", k, v))
			}
		}
		buf.WriteString("\n")
	}

	return buf.String(), nil
}

var _ = strings.Split
