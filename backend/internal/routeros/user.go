package routeros

import "fmt"

// HotspotUser merepresentasikan entri /ip/hotspot/user.
type HotspotUser struct {
	ID              string `json:"id"` // .id dari RouterOS
	Name            string `json:"name"`
	Password        string `json:"password"`
	Profile         string `json:"profile"`
	Comment         string `json:"comment"`
	Disabled        bool   `json:"disabled"`
	Uptime          string `json:"uptime"`
	BytesIn         string `json:"bytes-in"`
	BytesOut        string `json:"bytes-out"`
	Dynamic         bool   `json:"dynamic"`
	LimitUptime     string `json:"limit-uptime"`
	LimitBytesTotal string `json:"limit-bytes-total"`
}

// ListHotspotUsers mengembalikan semua hotspot user dari router.
func (cl *Client) ListHotspotUsers() ([]HotspotUser, error) {
	rows, err := cl.Print("/ip/hotspot/user")
	if err != nil {
		return nil, err
	}
	out := make([]HotspotUser, 0, len(rows))
	for _, r := range rows {
		out = append(out, rowToHotspotUser(r))
	}
	return out, nil
}

// FindHotspotUser mencari user berdasarkan name.
func (cl *Client) FindHotspotUser(name string) (HotspotUser, error) {
	rows, err := cl.Print("/ip/hotspot/user", "name", name)
	if err != nil {
		return HotspotUser{}, err
	}
	if len(rows) == 0 {
		return HotspotUser{}, fmt.Errorf("user %q tidak ditemukan", name)
	}
	return rowToHotspotUser(rows[0]), nil
}

// AddHotspotUser membuat hotspot user baru. Mengembalikan row termasuk .id.
// RouterOS add tidak selalu mengembalikan .id di reply, jadi kita print ulang
// untuk mendapatkan .id yang benar.
// limitUptime mengikuti format RouterOS: "1h", "30m", "1d", "1d12h", dst.
// comment bisa dikosongkan.
// limitBytesTraffic bisa dikosongkan (mis. "1G", "500M").
func (cl *Client) AddHotspotUser(name, password, profile, limitUptime, limitBytesTraffic, comment string) (HotspotUser, error) {
	kv := []string{"name", name, "password", password, "disabled", "no"}
	if profile != "" {
		kv = append(kv, "profile", profile)
	}
	if limitUptime != "" {
		kv = append(kv, "limit-uptime", limitUptime)
	}
	if limitBytesTraffic != "" {
		kv = append(kv, "limit-bytes-total", limitBytesTraffic)
	}
	if comment != "" {
		kv = append(kv, "comment", comment)
	}
	_, err := cl.Add("/ip/hotspot/user/add", kv...)
	if err != nil {
		return HotspotUser{}, fmt.Errorf("add hotspot user %s: %w", name, err)
	}
	// ambil .id yang benar via print
	rows, err := cl.Print("/ip/hotspot/user", "name", name)
	if err != nil {
		// add berhasil tapi gagal ambil id - tetap return data dasar
		return HotspotUser{Name: name, Password: password, Profile: profile, Comment: comment, LimitUptime: limitUptime, LimitBytesTotal: limitBytesTraffic}, nil
	}
	for _, r := range rows {
		if r["name"] == name {
			u := rowToHotspotUser(r)
			u.Password = password
			u.Profile = profile
			u.Comment = comment
			u.LimitUptime = limitUptime
			u.LimitBytesTotal = limitBytesTraffic
			return u, nil
		}
	}
	return HotspotUser{Name: name, Password: password, Profile: profile, Comment: comment, LimitUptime: limitUptime, LimitBytesTotal: limitBytesTraffic}, nil
}

// DisableHotspotUser men-disable user berdasarkan .id.
func (cl *Client) DisableHotspotUser(id string) error {
	return cl.Set("/ip/hotspot/user/set", id, "disabled", "yes")
}

// EnableHotspotUser men-enable user berdasarkan .id.
func (cl *Client) EnableHotspotUser(id string) error {
	return cl.Set("/ip/hotspot/user/set", id, "disabled", "no")
}

// RemoveHotspotUser menghapus user berdasarkan .id.
func (cl *Client) RemoveHotspotUser(id string) error {
	return cl.Remove("/ip/hotspot/user/remove", id)
}

// SetHotspotUserComment mengubah comment user.
func (cl *Client) SetHotspotUserComment(id, comment string) error {
	return cl.Set("/ip/hotspot/user/set", id, "comment", comment)
}

func rowToHotspotUser(r map[string]string) HotspotUser {
	return HotspotUser{
		ID:              r[".id"],
		Name:            r["name"],
		Password:        r["password"],
		Profile:         r["profile"],
		Comment:         r["comment"],
		Disabled:        r["disabled"] == "true",
		Uptime:          r["uptime"],
		BytesIn:         r["bytes-in"],
		BytesOut:        r["bytes-out"],
		Dynamic:         r["dynamic"] == "true",
		LimitUptime:     r["limit-uptime"],
		LimitBytesTotal: r["limit-bytes-total"],
	}
}
