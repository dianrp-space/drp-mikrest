package routeros

import "strings"

// ActiveUser merepresentasikan /ip/hotspot/active.
type ActiveUser struct {
	ID         string `json:"id"` // .id
	Server     string `json:"server"`
	User       string `json:"user"`
	Address    string `json:"address"`
	MACAddress string `json:"mac-address"`
	LoginBy    string `json:"login-by"`
	Uptime     string `json:"uptime"`
	BytesIn    string `json:"bytes-in"`
	BytesOut   string `json:"bytes-out"`
	SessionID  string `json:"session-id-left"`
}

// ListActiveUsers mengembalikan user hotspot yang sedang aktif.
func (cl *Client) ListActiveUsers() ([]ActiveUser, error) {
	rows, err := cl.Print("/ip/hotspot/active")
	if err != nil {
		return nil, err
	}
	out := make([]ActiveUser, 0, len(rows))
	for _, r := range rows {
		out = append(out, ActiveUser{
			ID:         r[".id"],
			Server:     r["server"],
			User:       r["user"],
			Address:    r["address"],
			MACAddress: r["mac-address"],
			LoginBy:    r["login-by"],
			Uptime:     r["uptime"],
			BytesIn:    r["bytes-in"],
			BytesOut:   r["bytes-out"],
			SessionID:  r["session-id-left"],
		})
	}
	return out, nil
}

// KickActiveUser memutuskan user aktif berdasarkan .id.
func (cl *Client) KickActiveUser(id string) error {
	return cl.Remove("/ip/hotspot/active/remove", id)
}

// SystemResource untuk dashboard.
type SystemResource struct {
	Uptime      string `json:"uptime"`
	Version     string `json:"version"`
	BoardName   string `json:"board-name"`
	CPULoad     string `json:"cpu-load"`
	FreeMemory  string `json:"free-memory"`
	TotalMemory string `json:"total-memory"`
	FreeHDD     string `json:"free-hdd-space"`
	TotalHDD    string `json:"total-hdd-space"`
}

// SystemResource mengambil /system/resource/print.
func (cl *Client) SystemResource() (SystemResource, error) {
	rows, err := cl.Print("/system/resource")
	if err != nil {
		return SystemResource{}, err
	}
	if len(rows) == 0 {
		return SystemResource{}, nil
	}
	r := rows[0]
	return SystemResource{
		Uptime:      r["uptime"],
		Version:     r["version"],
		BoardName:   r["board-name"],
		CPULoad:     r["cpu-load"],
		FreeMemory:  r["free-memory"],
		TotalMemory: r["total-memory"],
		FreeHDD:     r["free-hdd-space"],
		TotalHDD:    r["total-hdd-space"],
	}, nil
}

// SystemIdentity mengambil nama router.
func (cl *Client) SystemIdentity() (string, error) {
	return cl.Ping()
}

func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
