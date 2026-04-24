package qrpro

type Code struct {
	ID             string   `json:"id"`
	Slug           string   `json:"slug"`
	DestinationURL string   `json:"destination_url"`
	Label          string   `json:"label,omitempty"`
	Tags           []string `json:"tags,omitempty"`
	GroupID        *string  `json:"group_id,omitempty"`
	CreatedAt      string   `json:"created_at"`
	UpdatedAt      string   `json:"updated_at"`
	ScanCount      int      `json:"scan_count,omitempty"`
	LastScanAt     *string  `json:"last_scan_at,omitempty"`
	Active         bool     `json:"active"`
	ShortURL       string   `json:"short_url"`
}

type CodeCreate struct {
	DestinationURL string   `json:"destination_url"`
	Slug           string   `json:"slug,omitempty"`
	Label          string   `json:"label,omitempty"`
	Tags           []string `json:"tags,omitempty"`
	GroupID        *string  `json:"group_id,omitempty"`
}

type CodePatch struct {
	DestinationURL *string  `json:"destination_url,omitempty"`
	Label          *string  `json:"label,omitempty"`
	Tags           []string `json:"tags,omitempty"`
	GroupID        *string  `json:"group_id,omitempty"`
	Active         *bool    `json:"active,omitempty"`
}

type Group struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	CreatedAt   string `json:"created_at"`
	CodeCount   int    `json:"code_count,omitempty"`
}

type Webhook struct {
	ID        string   `json:"id"`
	URL       string   `json:"url"`
	Events    []string `json:"events"`
	Active    bool     `json:"active"`
	CreatedAt string   `json:"created_at"`
	Secret    string   `json:"secret,omitempty"` // only on creation
}

type Pagination struct {
	Cursor  *string `json:"cursor,omitempty"`
	Limit   int     `json:"limit,omitempty"`
	HasMore bool    `json:"has_more,omitempty"`
}

type ListResult[T any] struct {
	Data       []T        `json:"data"`
	Pagination Pagination `json:"pagination"`
}

type Analytics struct {
	CodeID     string `json:"code_id"`
	TotalScans int    `json:"total_scans"`
	UniqueDays int    `json:"unique_days"`
	ByDay      []struct {
		Date  string `json:"date"`
		Scans int    `json:"scans"`
	} `json:"by_day"`
	ByCountry []struct {
		Country string `json:"country"`
		Scans   int    `json:"scans"`
	} `json:"by_country"`
	ByDevice []struct {
		DeviceClass string `json:"device_class"`
		Scans       int    `json:"scans"`
	} `json:"by_device"`
}
