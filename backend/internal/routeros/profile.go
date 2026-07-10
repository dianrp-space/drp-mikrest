package routeros

import "fmt"

// HotspotProfile merepresentasikan /ip/hotspot/user/profile.
// Profile di RouterOS TIDAK punya property `comment` (comment hanya di /ip/hotspot/user).
type HotspotProfile struct {
	ID               string   `json:"id"` // .id
	Name             string   `json:"name"`
	RateLimit        string   `json:"rate-limit"`
	SessionTimeout   string   `json:"session-timeout"`
	IdleTimeout      string   `json:"idle-timeout"`
	SharedUsers      string   `json:"shared-users"`
	KeepaliveTimeout string   `json:"keepalive-timeout"`
	LoginBy          []string `json:"login-by"`
}

// ListHotspotProfiles mengembalikan semua profile dari router.
func (cl *Client) ListHotspotProfiles() ([]HotspotProfile, error) {
	rows, err := cl.Print("/ip/hotspot/user/profile")
	if err != nil {
		return nil, err
	}
	out := make([]HotspotProfile, 0, len(rows))
	for _, r := range rows {
		out = append(out, rowToHotspotProfile(r))
	}
	return out, nil
}

// AddHotspotProfile membuat profile baru di router.
func (cl *Client) AddHotspotProfile(p HotspotProfile) (HotspotProfile, error) {
	kv := []string{"name", p.Name}
	if p.RateLimit != "" {
		kv = append(kv, "rate-limit", p.RateLimit)
	}
	if p.SessionTimeout != "" {
		kv = append(kv, "session-timeout", p.SessionTimeout)
	}
	if p.IdleTimeout != "" {
		kv = append(kv, "idle-timeout", p.IdleTimeout)
	}
	if p.SharedUsers != "" {
		kv = append(kv, "shared-users", p.SharedUsers)
	}
	if p.KeepaliveTimeout != "" {
		kv = append(kv, "keepalive-timeout", p.KeepaliveTimeout)
	}
	row, err := cl.Add("/ip/hotspot/user/profile/add", kv...)
	if err != nil {
		return HotspotProfile{}, fmt.Errorf("add profile %s: %w", p.Name, err)
	}
	out := rowToHotspotProfile(row)
	if out.Name == "" {
		out.Name = p.Name
	}
	return out, nil
}

// UpdateHotspotProfile mengubah profile berdasarkan .id.
func (cl *Client) UpdateHotspotProfile(id string, p HotspotProfile) error {
	kv := []string{"name", p.Name}
	if p.RateLimit != "" {
		kv = append(kv, "rate-limit", p.RateLimit)
	}
	if p.SessionTimeout != "" {
		kv = append(kv, "session-timeout", p.SessionTimeout)
	}
	if p.IdleTimeout != "" {
		kv = append(kv, "idle-timeout", p.IdleTimeout)
	}
	if p.SharedUsers != "" {
		kv = append(kv, "shared-users", p.SharedUsers)
	}
	return cl.Set("/ip/hotspot/user/profile/set", id, kv...)
}

// RemoveHotspotProfile menghapus profile berdasarkan .id.
func (cl *Client) RemoveHotspotProfile(id string) error {
	return cl.Remove("/ip/hotspot/user/profile/remove", id)
}

func rowToHotspotProfile(r map[string]string) HotspotProfile {
	return HotspotProfile{
		ID:               r[".id"],
		Name:             r["name"],
		RateLimit:        r["rate-limit"],
		SessionTimeout:   r["session-timeout"],
		IdleTimeout:      r["idle-timeout"],
		SharedUsers:      r["shared-users"],
		KeepaliveTimeout: r["keepalive-timeout"],
		LoginBy:          splitCSV(r["login-by"]),
	}
}
